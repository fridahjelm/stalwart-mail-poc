// Command booking answers meeting-room invitations on behalf of room
// accounts in Stalwart. Each room is an ordinary account; the service polls
// its calendar over CalDAV and accepts or declines by setting the room's
// PARTSTAT, which makes Stalwart send the iTIP reply to the organizer.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"strings"
	"time"
	_ "time/tzdata" // the runtime image has no zoneinfo
)

type Config struct {
	PollSeconds int    `json:"pollSeconds"`
	TimeZone    string `json:"timeZone"`
	Defaults    Rules  `json:"defaults"`
	Rooms       []Room `json:"rooms"`
}

func main() {
	cfg := mustConfig(env("BOOKING_CONFIG", "/etc/booking/rooms.json"))
	passwords := mustPasswords(env("BOOKING_PASSWORDS", "/run/booking/passwords.txt"))
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		log.Fatalf("time zone: %v", err)
	}
	mailer, err := NewMailer(env("SMTP_ADDR", "mail.lab.internal:587"), env("SMTP_CA", ""))
	if err != nil {
		log.Fatalf("mailer: %v", err)
	}
	base := env("STALWART_URL", "http://mail.lab.internal:8080")

	started := 0
	for _, room := range cfg.Rooms {
		room.Address = strings.ToLower(room.Address)
		pw, ok := passwords[room.Address]
		if !ok {
			log.Printf("%s: no app password, skipped (run scripts/37-room-passwords.sh)", room.Address)
			continue
		}
		rules := cfg.Defaults
		if room.Rules != nil {
			rules = *room.Rules
		}
		w := &worker{room: room, rules: rules, loc: loc, mailer: mailer, password: pw,
			dav: &DAV{Base: base, User: room.Address, Password: pw}}
		go w.run(time.Duration(cfg.PollSeconds) * time.Second)
		started++
	}
	if started == 0 {
		log.Fatal("no rooms to serve")
	}
	select {}
}

// worker serves one room. Each room has exactly one worker, so two
// invitations for the same slot are judged one after the other and cannot
// both be accepted.
type worker struct {
	room     Room
	rules    Rules
	loc      *time.Location
	dav      *DAV
	mailer   *Mailer
	password string
}

func (w *worker) run(every time.Duration) {
	log.Printf("%s: serving %s", w.room.Address, CalendarPath(w.room.Address))
	for {
		if err := w.poll(time.Now()); err != nil {
			log.Printf("%s: %v", w.room.Address, err)
		}
		time.Sleep(every)
	}
}

func (w *worker) poll(now time.Time) error {
	res, err := w.dav.List(CalendarPath(w.room.Address))
	if err != nil {
		return err
	}
	var held, pending []*Event
	for _, r := range res {
		ev := ParseEvent(r.Href, r.ETag, r.Data, w.room.Address, w.loc)
		switch {
		case ev.Status == "CANCELLED":
			// Stalwart keeps a cancelled event in the attendee's calendar;
			// for a room it only blocks nothing and clutters the calendar.
			if err := w.dav.Delete(ev.Href, ev.ETag); err != nil && !errors.Is(err, ErrChanged) {
				log.Printf("%s: delete cancelled %q: %v", w.room.Address, ev.Summary, err)
			} else if err == nil {
				log.Printf("%s: removed cancelled %q", w.room.Address, ev.Summary)
			}
		case ev.Answered() && ev.PartStat == "DECLINED":
			// Stalwart sends no CANCEL to an attendee who declined, so a
			// declined copy would stay forever once the organizer deletes the
			// event. It blocks nothing; if the organizer sends an update,
			// Stalwart delivers it as a new copy and it is judged again.
			if err := w.dav.Delete(ev.Href, ev.ETag); err != nil && !errors.Is(err, ErrChanged) {
				log.Printf("%s: delete declined %q: %v", w.room.Address, ev.Summary, err)
			}
		case ev.Answered():
			if ev.PartStat == "ACCEPTED" {
				held = append(held, ev)
			}
		default:
			pending = append(pending, ev)
		}
	}
	// Oldest change first, so an earlier request wins a contested slot.
	sort.Slice(pending, func(i, j int) bool {
		if pending[i].Stamp != pending[j].Stamp {
			return pending[i].Stamp < pending[j].Stamp
		}
		return pending[i].UID < pending[j].UID
	})
	for _, ev := range pending {
		reason := Decide(ev, held, w.room, w.rules, w.loc, now)
		answer := "ACCEPTED"
		if reason != "" {
			answer = "DECLINED"
		}
		ev.SetAnswer(w.room.Address, answer)
		if err := w.dav.Put(ev.Href, ev.ETag, ev.Doc.String()); err != nil {
			if !errors.Is(err, ErrChanged) {
				log.Printf("%s: answer %q: %v", w.room.Address, ev.Summary, err)
			}
			continue
		}
		if answer == "ACCEPTED" {
			held = append(held, ev)
			log.Printf("%s: accepted %q %s from %s", w.room.Address, ev.Summary, w.span(ev), ev.Organizer)
			continue
		}
		log.Printf("%s: declined %q %s from %s: %s", w.room.Address, ev.Summary, w.span(ev), ev.Organizer, reason)
		if ev.Organizer != "" {
			if err := w.mailer.SendDecline(w.room, w.password, ev, reason, w.loc); err != nil {
				log.Printf("%s: decline mail to %s: %v", w.room.Address, ev.Organizer, err)
			}
		}
	}
	return nil
}

func (w *worker) span(ev *Event) string {
	return ev.Start.In(w.loc).Format("2006-01-02 15:04") + "–" + ev.End.In(w.loc).Format("15:04")
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustConfig(path string) Config {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		log.Fatalf("config %s: %v", path, err)
	}
	if c.PollSeconds <= 0 {
		c.PollSeconds = 5
	}
	return c
}

// mustPasswords reads "address<TAB>app password" lines, the format the
// scripts in scripts/ write.
func mustPasswords(path string) map[string]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("passwords: %v", err)
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if addr, pw, ok := strings.Cut(strings.TrimSpace(sc.Text()), "\t"); ok {
			out[strings.ToLower(addr)] = pw
		}
	}
	return out
}
