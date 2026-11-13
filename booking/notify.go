package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Mailer sends the decline reason, which Stalwart's iTIP reply does not
// carry: a COMMENT on the room's copy never reaches the organizer.
type Mailer struct {
	Addr       string // host:port of the submission listener
	ServerName string
	Roots      *x509.CertPool
}

func NewMailer(addr, caFile string) (*Mailer, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	m := &Mailer{Addr: addr, ServerName: host}
	if caFile != "" {
		pem, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		m.Roots = x509.NewCertPool()
		if !m.Roots.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificates in %s", caFile)
		}
	}
	return m, nil
}

func (m *Mailer) SendDecline(room Room, password string, ev *Event, reason string, loc *time.Location) error {
	c, err := smtp.Dial(m.Addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.StartTLS(&tls.Config{ServerName: m.ServerName, RootCAs: m.Roots}); err != nil {
		return err
	}
	if err := c.Auth(smtp.PlainAuth("", room.Address, password, m.ServerName)); err != nil {
		return err
	}
	if err := c.Mail(room.Address); err != nil {
		return err
	}
	if err := c.Rcpt(ev.Organizer); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	when := fmt.Sprintf("%s %s–%s", ev.Start.In(loc).Format("Mon 2 Jan 2006"),
		ev.Start.In(loc).Format("15:04"), ev.End.In(loc).Format("15:04"))
	subject := fmt.Sprintf("%s declined: %s", room.Name, oneLine(ev.Summary))
	body := fmt.Sprintf("%s declined \"%s\" (%s): %s.\r\n\r\nThis message comes from the lab booking service.\r\n",
		room.Name, oneLine(ev.Summary), when, reason)
	msg := strings.Join([]string{
		"From: " + room.Name + " <" + room.Address + ">",
		"To: " + ev.Organizer,
		"Subject: " + subject,
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: 8bit",
		"Auto-Submitted: auto-replied",
		"", body}, "\r\n")
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// oneLine keeps a summary from injecting header lines into the message.
func oneLine(s string) string {
	s = strings.NewReplacer("\r", " ", "\n", " ", `\n`, " ", `\,`, ",", `\;`, ";").Replace(s)
	return strings.TrimSpace(s)
}
