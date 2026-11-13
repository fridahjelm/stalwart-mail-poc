package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DAV struct {
	Base     string // e.g. http://mail.lab.internal:8080
	User     string
	Password string
	HTTP     *http.Client
}

type Resource struct {
	Href, ETag, Data string
}

// CalendarPath is the room's default calendar, as Stalwart names it.
func CalendarPath(addr string) string {
	return "/dav/cal/" + strings.ReplaceAll(url.PathEscape(addr), "@", "%40") + "/default/"
}

type multistatus struct {
	Responses []struct {
		Href     string `xml:"DAV: href"`
		Propstat []struct {
			Status string `xml:"DAV: status"`
			Prop   struct {
				ETag string `xml:"DAV: getetag"`
				Data string `xml:"urn:ietf:params:xml:ns:caldav calendar-data"`
			} `xml:"DAV: prop"`
		} `xml:"DAV: propstat"`
	} `xml:"DAV: response"`
}

const listBody = `<?xml version="1.0" encoding="utf-8"?>
<propfind xmlns="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"><prop><getetag/><C:calendar-data/></prop></propfind>`

// List returns every event resource in the calendar with its data.
func (d *DAV) List(path string) ([]Resource, error) {
	resp, err := d.do("PROPFIND", path, strings.NewReader(listBody), map[string]string{
		"Depth": "1", "Content-Type": "application/xml; charset=utf-8"})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("PROPFIND %s: %s", path, resp.Status)
	}
	var ms multistatus
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return nil, fmt.Errorf("PROPFIND %s: %w", path, err)
	}
	var out []Resource
	for _, r := range ms.Responses {
		if strings.HasSuffix(r.Href, "/") {
			continue // the collection itself
		}
		for _, ps := range r.Propstat {
			if strings.Contains(ps.Status, " 200 ") && ps.Prop.Data != "" {
				out = append(out, Resource{Href: r.Href, ETag: ps.Prop.ETag, Data: ps.Prop.Data})
			}
		}
	}
	return out, nil
}

// Put replaces a resource only if it is unchanged since it was read; a
// changed resource returns ErrChanged and is handled on the next poll.
func (d *DAV) Put(href, etag, data string) error {
	resp, err := d.do("PUT", href, strings.NewReader(data), map[string]string{
		"Content-Type": "text/calendar; charset=utf-8", "If-Match": etag})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return status(resp, "PUT", href)
}

// Delete removes the room's copy without sending a reply to the organizer.
func (d *DAV) Delete(href, etag string) error {
	resp, err := d.do("DELETE", href, nil, map[string]string{"If-Match": etag, "Schedule-Reply": "F"})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return status(resp, "DELETE", href)
}

var ErrChanged = fmt.Errorf("changed since read")

func status(resp *http.Response, method, href string) error {
	switch {
	case resp.StatusCode == http.StatusPreconditionFailed:
		return ErrChanged
	case resp.StatusCode >= 300:
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("%s %s: %s %s", method, href, resp.Status, strings.TrimSpace(string(b)))
	}
	return nil
}

func (d *DAV) do(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, d.Base+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(d.User, d.Password)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c := d.HTTP
	if c == nil {
		c = &http.Client{Timeout: 20 * time.Second}
	}
	return c.Do(req)
}
