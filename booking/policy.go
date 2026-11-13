package main

import (
	"fmt"
	"strings"
	"time"
)

type Rules struct {
	MaxMinutes       int      `json:"maxMinutes"`
	HorizonDays      int      `json:"horizonDays"`
	OpenFrom         string   `json:"openFrom"` // HH:MM, room time zone
	OpenTo           string   `json:"openTo"`
	OrganizerDomains []string `json:"organizerDomains"`
}

type Room struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Rules    *Rules `json:"rules,omitempty"` // replaces the defaults when set
}

// Decide returns "" to accept, or the reason for declining. held are the
// room's bookings that stand: accepted, answered for their current version
// and not cancelled.
func Decide(ev *Event, held []*Event, room Room, r Rules, loc *time.Location, now time.Time) string {
	switch {
	case ev.ParseErr != nil:
		return "the invitation could not be read: " + ev.ParseErr.Error()
	case ev.Recurring:
		return "recurring bookings are not supported"
	case ev.AllDay:
		return "all-day bookings are not supported"
	case !ev.End.After(ev.Start):
		return "the booking has no duration"
	case !ev.End.After(now):
		return "the booking is in the past"
	case ev.Start.After(now.AddDate(0, 0, r.HorizonDays)):
		return fmt.Sprintf("bookings are taken at most %d days ahead", r.HorizonDays)
	case ev.End.Sub(ev.Start) > time.Duration(r.MaxMinutes)*time.Minute:
		return fmt.Sprintf("bookings may last at most %s", fmtMinutes(r.MaxMinutes))
	}
	if !domainAllowed(ev.Organizer, r.OrganizerDomains) {
		return "the organizer's domain may not book this room"
	}
	if reason := outsideHours(ev, r, loc); reason != "" {
		return reason
	}
	if room.Capacity > 0 && len(ev.Attendees) > room.Capacity {
		return fmt.Sprintf("%d people are invited and the room seats %d", len(ev.Attendees), room.Capacity)
	}
	for _, h := range held {
		if h.UID != ev.UID && ev.Start.Before(h.End) && h.Start.Before(ev.End) {
			return fmt.Sprintf("the room is already booked %s–%s",
				h.Start.In(loc).Format("15:04"), h.End.In(loc).Format("15:04"))
		}
	}
	return ""
}

func domainAllowed(addr string, domains []string) bool {
	_, dom, ok := strings.Cut(addr, "@")
	if !ok {
		return false
	}
	for _, d := range domains {
		if strings.EqualFold(dom, d) {
			return true
		}
	}
	return false
}

func outsideHours(ev *Event, r Rules, loc *time.Location) string {
	s, e := ev.Start.In(loc), ev.End.In(loc)
	from, to := clock(s, r.OpenFrom), clock(s, r.OpenTo)
	if s.Before(from) || e.After(to) || e.YearDay() != s.YearDay() {
		return fmt.Sprintf("the room is open %s–%s", r.OpenFrom, r.OpenTo)
	}
	return ""
}

// clock returns HH:MM on the same day as t, in t's location.
func clock(t time.Time, hhmm string) time.Time {
	var h, m int
	fmt.Sscanf(hhmm, "%d:%d", &h, &m)
	return time.Date(t.Year(), t.Month(), t.Day(), h, m, 0, 0, t.Location())
}

func fmtMinutes(m int) string {
	if m%60 == 0 {
		return fmt.Sprintf("%d h", m/60)
	}
	return fmt.Sprintf("%d min", m)
}
