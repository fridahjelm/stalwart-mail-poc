package main

import (
	"strings"
	"testing"
	"time"
)

const room = "room-a@lab.internal"

func invite(start, end string, extra ...string) string {
	lines := []string{
		"BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//test//EN",
		"BEGIN:VTIMEZONE", "TZID:Europe/Stockholm", "BEGIN:STANDARD",
		"DTSTART:19701025T030000", "TZOFFSETFROM:+0200", "TZOFFSETTO:+0100", "END:STANDARD", "END:VTIMEZONE",
		"BEGIN:VEVENT", "UID:u1", "DTSTAMP:20261113T120000Z", "SUMMARY:Planning",
		"DTSTART;TZID=Europe/Stockholm:" + start, "DTEND;TZID=Europe/Stockholm:" + end,
		`ORGANIZER;CN="Andersson, Alice":mailto:alice@lab.internal`,
		"ATTENDEE;PARTSTAT=ACCEPTED:mailto:alice@lab.internal",
		"ATTENDEE;CUTYPE=ROOM;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:" + room,
	}
	lines = append(lines, extra...)
	lines = append(lines, "BEGIN:VALARM", "TRIGGER:-PT15M", "ACTION:DISPLAY", "END:VALARM", "END:VEVENT", "END:VCALENDAR")
	return strings.Join(lines, "\r\n") + "\r\n"
}

func stockholm(t *testing.T) *time.Location {
	loc, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

var rules = Rules{MaxMinutes: 240, HorizonDays: 90, OpenFrom: "07:00", OpenTo: "20:00",
	OrganizerDomains: []string{"lab.internal", "other.lab.internal"}}
var roomA = Room{Address: room, Name: "Room A", Capacity: 3}

func TestFoldRoundTrip(t *testing.T) {
	long := "DESCRIPTION:" + strings.Repeat("åäö ", 40)
	d := ParseDoc("BEGIN:VCALENDAR\r\n" + long + "\r\nEND:VCALENDAR\r\n")
	out := d.String()
	for _, l := range strings.Split(out, "\r\n") {
		if len(l) > 75 {
			t.Fatalf("line of %d octets", len(l))
		}
	}
	if got := ParseDoc(out).Lines[1]; got != long {
		t.Fatalf("round trip changed the line:\n%q\n%q", got, long)
	}
}

func TestParsePropQuoted(t *testing.T) {
	p := ParseProp(`ORGANIZER;CN="Doe: Jane; Ltd";SENT-BY="mailto:x@y":mailto:Jane@Lab.Internal`)
	if p.Name != "ORGANIZER" || p.Param("CN") != "Doe: Jane; Ltd" || p.Address() != "jane@lab.internal" {
		t.Fatalf("%+v", p)
	}
	if p.String() != `ORGANIZER;CN="Doe: Jane; Ltd";SENT-BY="mailto:x@y":mailto:Jane@Lab.Internal` {
		t.Fatalf("not written back unchanged: %s", p.String())
	}
}

func TestParseEvent(t *testing.T) {
	loc := stockholm(t)
	ev := ParseEvent("/h", `"e"`, invite("20261120T100000", "20261120T110000"), room, loc)
	if ev.ParseErr != nil {
		t.Fatal(ev.ParseErr)
	}
	if want := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC); !ev.Start.Equal(want) {
		t.Fatalf("start %v, want %v", ev.Start, want)
	}
	if ev.PartStat != "NEEDS-ACTION" || ev.Organizer != "alice@lab.internal" || len(ev.Attendees) != 1 || ev.Recurring {
		t.Fatalf("%+v", ev)
	}
}

func TestDurationInsteadOfEnd(t *testing.T) {
	ics := strings.Replace(invite("20261120T100000", "X"), "DTEND;TZID=Europe/Stockholm:X", "DURATION:PT1H30M", 1)
	ev := ParseEvent("/h", "", ics, room, stockholm(t))
	if ev.End.Sub(ev.Start) != 90*time.Minute {
		t.Fatalf("duration %v", ev.End.Sub(ev.Start))
	}
}

func TestAnswerMarker(t *testing.T) {
	loc := stockholm(t)
	ev := ParseEvent("/h", "", invite("20261120T100000", "20261120T110000"), room, loc)
	if ev.Answered() {
		t.Fatal("unanswered event reported as answered")
	}
	ev.SetAnswer(room, "ACCEPTED")
	again := ParseEvent("/h", "", ev.Doc.String(), room, loc)
	if !again.Answered() || again.PartStat != "ACCEPTED" {
		t.Fatalf("answer not read back: %+v", again)
	}
	if strings.Contains(ev.Doc.String(), "RSVP=TRUE:mailto:"+room) {
		t.Fatal("RSVP left on the room's attendee line")
	}
	if strings.Count(ev.Doc.String(), "BEGIN:VALARM") != 1 || !strings.Contains(ev.Doc.String(), markerProp+":") {
		t.Fatal("document damaged by SetAnswer")
	}
	// The organizer moves the meeting: the old answer no longer counts.
	moved := strings.Replace(again.Doc.String(), "20261120T100000", "20261120T130000", 1)
	moved = strings.Replace(moved, "20261120T110000", "20261120T140000", 1)
	if ParseEvent("/h", "", moved, room, loc).Answered() {
		t.Fatal("rescheduled event still reported as answered")
	}
}

func TestDecide(t *testing.T) {
	loc := stockholm(t)
	now := time.Date(2026, 11, 13, 12, 0, 0, 0, loc)
	held := ParseEvent("/held", "", strings.Replace(invite("20261120T100000", "20261120T110000"), "UID:u1", "UID:held", 1), room, loc)

	cases := []struct {
		name, start, end string
		extra            []string
		want             string // substring of the reason; "" means accept
	}{
		{"free slot", "20261120T110000", "20261120T120000", nil, ""},
		{"overlap", "20261120T103000", "20261120T113000", nil, "already booked 10:00–11:00"},
		{"too long", "20261120T080000", "20261120T130000", nil, "at most 4 h"},
		{"before opening", "20261120T063000", "20261120T073000", nil, "open 07:00–20:00"},
		{"after closing", "20261120T193000", "20261120T203000", nil, "open 07:00–20:00"},
		{"past", "20261112T100000", "20261112T110000", nil, "in the past"},
		{"beyond horizon", "20270301T100000", "20270301T110000", nil, "90 days ahead"},
		{"recurring", "20261120T130000", "20261120T140000", []string{"RRULE:FREQ=WEEKLY"}, "recurring"},
		{"over capacity", "20261120T130000", "20261120T140000", []string{
			"ATTENDEE:mailto:b@lab.internal", "ATTENDEE:mailto:c@lab.internal", "ATTENDEE:mailto:d@lab.internal"}, "seats 3"},
		{"other rooms do not count", "20261120T130000", "20261120T140000", []string{
			"ATTENDEE;CUTYPE=ROOM:mailto:room-b@lab.internal", "ATTENDEE:mailto:b@lab.internal"}, ""},
	}
	for _, c := range cases {
		ev := ParseEvent("/e", "", invite(c.start, c.end, c.extra...), room, loc)
		got := Decide(ev, []*Event{held}, roomA, rules, loc, now)
		if (c.want == "") != (got == "") || !strings.Contains(got, c.want) {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}

	foreign := ParseEvent("/e", "", strings.ReplaceAll(invite("20261120T130000", "20261120T140000"),
		"mailto:alice@lab.internal", "mailto:eve@example.com"), room, loc)
	if got := Decide(foreign, nil, roomA, rules, loc, now); !strings.Contains(got, "domain") {
		t.Errorf("foreign organizer: got %q", got)
	}
}

func TestParseDuration(t *testing.T) {
	for s, want := range map[string]time.Duration{"PT1H": time.Hour, "P1D": 24 * time.Hour, "PT45M": 45 * time.Minute, "P1W": 7 * 24 * time.Hour} {
		if got, err := ParseDuration(s); err != nil || got != want {
			t.Errorf("%s: %v %v", s, got, err)
		}
	}
}
