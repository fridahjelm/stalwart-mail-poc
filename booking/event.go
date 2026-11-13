package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// markerProp records which version of an event the service last answered.
// Stalwart keeps it on the room's copy when the organizer sends an update,
// while that update changes SEQUENCE, the times or the attendees, so a
// mismatch means the event has to be judged again.
const markerProp = "X-LAB-BOOKING"

// Event is the room's copy of one invitation.
type Event struct {
	Href, ETag string
	Doc        *Doc
	span       Span

	UID, Summary, Status string
	Organizer            string
	Sequence             int
	Stamp                string
	Start, End           time.Time
	AllDay, Recurring    bool
	Attendees            []string // people only: rooms and resources left out
	PartStat             string   // the room's own answer
	Marker               string
	ParseErr             error
}

func ParseEvent(href, etag, data, room string, loc *time.Location) *Event {
	ev := &Event{Href: href, ETag: etag, Doc: ParseDoc(data), PartStat: "NEEDS-ACTION"}
	spans := ev.Doc.Components("VEVENT")
	if len(spans) == 0 {
		ev.ParseErr = errors.New("no VEVENT")
		return ev
	}
	// More than one VEVENT means overrides of a recurring event.
	ev.Recurring = len(spans) > 1
	ev.span = spans[0]

	var dtend, duration *Prop
	people := map[string]bool{}
	for _, p := range ev.Doc.Props(ev.span) {
		p := p
		switch p.Name {
		case "UID":
			ev.UID = p.Value
		case "SUMMARY":
			ev.Summary = p.Value
		case "STATUS":
			ev.Status = strings.ToUpper(p.Value)
		case "SEQUENCE":
			ev.Sequence, _ = strconv.Atoi(p.Value)
		case "DTSTAMP":
			ev.Stamp = p.Value
		case "ORGANIZER":
			ev.Organizer = p.Address()
		case "RRULE", "RDATE":
			ev.Recurring = true
		case "DTSTART":
			var err error
			ev.Start, ev.AllDay, err = ParseDateTime(p, loc)
			if err != nil {
				ev.ParseErr = err
			}
		case "DTEND":
			dtend = &p
		case "DURATION":
			duration = &p
		case markerProp:
			ev.Marker = p.Value
		case "ATTENDEE":
			addr := p.Address()
			if addr == room {
				if ps := strings.ToUpper(p.Param("PARTSTAT")); ps != "" {
					ev.PartStat = ps
				}
				continue
			}
			switch strings.ToUpper(p.Param("CUTYPE")) {
			case "ROOM", "RESOURCE":
				continue
			}
			people[addr] = true
		}
	}
	if ev.Organizer != "" && ev.Organizer != room {
		people[ev.Organizer] = true
	}
	for a := range people {
		ev.Attendees = append(ev.Attendees, a)
	}
	sort.Strings(ev.Attendees)

	switch {
	case ev.ParseErr != nil:
	case dtend != nil:
		end, _, err := ParseDateTime(*dtend, loc)
		ev.End, ev.ParseErr = end, err
	case duration != nil:
		d, err := ParseDuration(duration.Value)
		ev.End, ev.ParseErr = ev.Start.Add(d), err
	case ev.AllDay:
		ev.End = ev.Start.AddDate(0, 0, 1)
	default:
		ev.End = ev.Start
	}
	return ev
}

// Fingerprint is what the marker stores: the parts of an event whose change
// makes an earlier answer invalid.
func (ev *Event) Fingerprint() string {
	return fmt.Sprintf("seq=%d;start=%s;end=%s;people=%d", ev.Sequence,
		ev.Start.UTC().Format("20060102T150405Z"), ev.End.UTC().Format("20060102T150405Z"), len(ev.Attendees))
}

// Answered reports whether the room's current answer is for this version.
func (ev *Event) Answered() bool {
	return ev.Marker != "" && ev.Marker == ev.Fingerprint() && ev.PartStat != "NEEDS-ACTION"
}

// SetAnswer sets the room's PARTSTAT and the marker on the room's copy.
func (ev *Event) SetAnswer(room, partStat string) {
	props := ev.Doc.Props(ev.span)
	markerLine := -1
	for i, p := range props {
		switch {
		case p.Name == "ATTENDEE" && p.Address() == room:
			p.SetParam("PARTSTAT", partStat)
			p.DelParam("RSVP")
			ev.Doc.Lines[i] = p.String()
		case p.Name == markerProp:
			markerLine = i
		}
	}
	marker := markerProp + ":" + ev.Fingerprint()
	if markerLine >= 0 {
		ev.Doc.Lines[markerLine] = marker
	} else {
		lines := append([]string{}, ev.Doc.Lines[:ev.span.End]...)
		lines = append(lines, marker)
		ev.Doc.Lines = append(lines, ev.Doc.Lines[ev.span.End:]...)
		ev.span.End++
	}
	ev.PartStat = partStat
	ev.Marker = ev.Fingerprint()
}
