package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Doc is an iCalendar object as unfolded content lines. Edits work on the
// lines directly, so everything the service does not touch is written back
// byte for byte (apart from line folding).
type Doc struct {
	Lines []string
}

func ParseDoc(s string) *Doc {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		if (strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t")) && len(lines) > 0 {
			lines[len(lines)-1] += l[1:]
			continue
		}
		if l != "" {
			lines = append(lines, l)
		}
	}
	return &Doc{Lines: lines}
}

func (d *Doc) String() string {
	var b strings.Builder
	for _, l := range d.Lines {
		b.WriteString(fold(l))
		b.WriteString("\r\n")
	}
	return b.String()
}

// fold splits a line into 75-octet pieces without cutting a UTF-8 sequence.
func fold(l string) string {
	if len(l) <= 75 {
		return l
	}
	var b strings.Builder
	limit := 75
	for len(l) > limit {
		cut := limit
		for cut > 0 && l[cut]&0xC0 == 0x80 {
			cut--
		}
		b.WriteString(l[:cut])
		b.WriteString("\r\n ")
		l = l[cut:]
		limit = 74
	}
	b.WriteString(l)
	return b.String()
}

// Prop is one content line: NAME;PARAM=VALUE;...:VALUE
type Prop struct {
	Name   string
	Params []Param
	Value  string
}

type Param struct {
	Name  string
	Value string // as written, quotes included
}

func ParseProp(line string) Prop {
	var p Prop
	i, inQuote := 0, false
	for ; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			inQuote = !inQuote
		} else if !inQuote && (c == ':') {
			break
		}
	}
	head := line[:i]
	if i < len(line) {
		p.Value = line[i+1:]
	}
	parts := splitUnquoted(head, ';')
	p.Name = strings.ToUpper(parts[0])
	for _, raw := range parts[1:] {
		k, v, _ := strings.Cut(raw, "=")
		p.Params = append(p.Params, Param{Name: strings.ToUpper(k), Value: v})
	}
	return p
}

func splitUnquoted(s string, sep byte) []string {
	var out []string
	start, inQuote := 0, false
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '"':
			inQuote = !inQuote
		case s[i] == sep && !inQuote:
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func (p Prop) String() string {
	var b strings.Builder
	b.WriteString(p.Name)
	for _, pa := range p.Params {
		b.WriteString(";" + pa.Name + "=" + pa.Value)
	}
	b.WriteString(":" + p.Value)
	return b.String()
}

func (p Prop) Param(name string) string {
	for _, pa := range p.Params {
		if pa.Name == name {
			return strings.Trim(pa.Value, `"`)
		}
	}
	return ""
}

func (p *Prop) SetParam(name, value string) {
	for i := range p.Params {
		if p.Params[i].Name == name {
			p.Params[i].Value = value
			return
		}
	}
	p.Params = append(p.Params, Param{Name: name, Value: value})
}

func (p *Prop) DelParam(name string) {
	out := p.Params[:0]
	for _, pa := range p.Params {
		if pa.Name != name {
			out = append(out, pa)
		}
	}
	p.Params = out
}

// Address returns the lower-cased address of a mailto: value.
func (p Prop) Address() string {
	v := p.Value
	if len(v) >= 7 && strings.EqualFold(v[:7], "mailto:") {
		v = v[7:]
	}
	return strings.ToLower(strings.TrimSpace(v))
}

// Span is a [begin, end] pair of line indexes for a component, inclusive.
type Span struct{ Begin, End int }

// Components returns the top-level spans of the named component, so a
// VALARM inside a VEVENT is not returned as a VEVENT of its own.
func (d *Doc) Components(name string) []Span {
	var out []Span
	depth, begin := 0, -1
	for i, l := range d.Lines {
		up := strings.ToUpper(l)
		switch {
		case strings.HasPrefix(up, "BEGIN:"):
			if up == "BEGIN:"+name && depth == 1 {
				begin = i
			}
			depth++
		case strings.HasPrefix(up, "END:"):
			depth--
			if up == "END:"+name && depth == 1 && begin >= 0 {
				out = append(out, Span{begin, i})
				begin = -1
			}
		}
	}
	return out
}

// Props returns the properties directly inside the span, with their line
// indexes; lines of nested components are skipped.
func (d *Doc) Props(s Span) map[int]Prop {
	out := map[int]Prop{}
	depth := 0
	for i := s.Begin + 1; i < s.End; i++ {
		up := strings.ToUpper(d.Lines[i])
		switch {
		case strings.HasPrefix(up, "BEGIN:"):
			depth++
		case strings.HasPrefix(up, "END:"):
			depth--
		case depth == 0:
			out[i] = ParseProp(d.Lines[i])
		}
	}
	return out
}

var reDuration = regexp.MustCompile(`^([+-])?P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

func ParseDuration(s string) (time.Duration, error) {
	m := reDuration.FindStringSubmatch(strings.ToUpper(s))
	if m == nil {
		return 0, fmt.Errorf("bad duration %q", s)
	}
	n := func(i int) time.Duration {
		v, _ := strconv.Atoi(m[i])
		return time.Duration(v)
	}
	d := n(2)*7*24*time.Hour + n(3)*24*time.Hour + n(4)*time.Hour + n(5)*time.Minute + n(6)*time.Second
	if m[1] == "-" {
		d = -d
	}
	return d, nil
}

// ParseDateTime reads DTSTART/DTEND. allDay is true for VALUE=DATE.
// Floating times are read in the room's time zone.
func ParseDateTime(p Prop, floating *time.Location) (t time.Time, allDay bool, err error) {
	v := p.Value
	if strings.EqualFold(p.Param("VALUE"), "DATE") || len(v) == 8 {
		t, err = time.ParseInLocation("20060102", v, floating)
		return t, true, err
	}
	if strings.HasSuffix(v, "Z") {
		t, err = time.Parse("20060102T150405Z", v)
		return t, false, err
	}
	loc := floating
	if tz := p.Param("TZID"); tz != "" {
		if loc, err = time.LoadLocation(strings.TrimPrefix(tz, "/")); err != nil {
			return t, false, fmt.Errorf("unknown time zone %q", tz)
		}
	}
	t, err = time.ParseInLocation("20060102T150405", v, loc)
	return t, false, err
}
