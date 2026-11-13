"""Mailbox access for the lab accounts.

  python3 -I scripts/mailbox.py list    <address> [mailbox]
  python3 -I scripts/mailbox.py read    <address> <uid-or-number> [mailbox]
  python3 -I scripts/mailbox.py reply   <address> <number> <<'EOF' ... EOF
  python3 -I scripts/mailbox.py send    <address> <to> <subject> <<'EOF' ... EOF

Credentials come from app-passwords.txt. reply keeps the thread by copying
In-Reply-To and References from the message being answered, and copies every
other recipient of that message so the thread stays together.
"""
import email
import imaplib
import smtplib
import socket
import ssl
import sys
import time
from email.message import EmailMessage
from email.policy import default as POLICY
from email.utils import formatdate, getaddresses, make_msgid

_getaddrinfo = socket.getaddrinfo
socket.getaddrinfo = lambda h, p, *a, **k: _getaddrinfo(
    "127.0.0.1" if isinstance(h, str) and h.endswith(".lab.internal") else h, p, *a, **k)

HOST = "mail.lab.internal"
CTX = ssl.create_default_context(cafile="certs/rootCA.pem")


def secret(address):
    for line in open("app-passwords.txt"):
        if line.strip():
            who, pw = line.rstrip("\n").split("\t")
            if who == address:
                return pw
    raise SystemExit(f"no app password for {address}; run scripts/35-app-passwords.sh")


def imap(address, mailbox="INBOX"):
    c = imaplib.IMAP4_SSL(HOST, 993, ssl_context=CTX)
    c.login(address, secret(address))
    c.select(mailbox)
    return c


def fetch(c, number):
    return email.message_from_bytes(c.fetch(str(number), "(RFC822)")[1][0][1], policy=POLICY)


def body_of(msg):
    part = msg.get_body(("plain",))
    return part.get_content().strip() if part else "(no plain text part)"


def deliver(sender, msg):
    s = smtplib.SMTP(HOST, 587, timeout=20)
    s.starttls(context=CTX)
    s.login(sender, secret(sender))
    s.send_message(msg)
    s.quit()
    c = imaplib.IMAP4_SSL(HOST, 993, ssl_context=CTX)
    c.login(sender, secret(sender))
    c.append("Sent Items", "\\Seen", imaplib.Time2Internaldate(time.time()), msg.as_bytes())
    c.logout()
    print(f"sent {msg['Message-ID']} to {msg['To']}" + (f" cc {msg['Cc']}" if msg["Cc"] else ""))


def cmd_list(address, mailbox="INBOX"):
    c = imap(address, mailbox)
    typ, data = c.search(None, "ALL")
    for n in data[0].split():
        m = email.message_from_bytes(c.fetch(n, "(BODY.PEEK[HEADER])")[1][0][1], policy=POLICY)
        print(f"{n.decode():>4}  {m['Date']}  {m['From']}  {m['Subject']}")
    c.logout()


def cmd_read(address, number, mailbox="INBOX"):
    c = imap(address, mailbox)
    m = fetch(c, number)
    for h in ("Date", "From", "To", "Cc", "Subject", "Message-ID"):
        if m[h]:
            print(f"{h}: {m[h]}")
    print("-" * 60)
    print(body_of(m))
    c.logout()


def cmd_reply(address, number):
    text = sys.stdin.read()
    c = imap(address)
    orig = fetch(c, number)
    c.logout()
    recipients = [a for _, a in getaddresses(
        [orig["From"] or "", orig["To"] or "", orig["Cc"] or ""])]
    to = [a for _, a in getaddresses([orig["From"] or ""])]
    cc = [a for a in dict.fromkeys(recipients) if a not in to and a != address]
    r = EmailMessage()
    r["From"] = address
    r["To"] = ", ".join(to)
    if cc:
        r["Cc"] = ", ".join(cc)
    subject = orig["Subject"] or ""
    r["Subject"] = subject if subject.lower().startswith("re:") else f"Re: {subject}"
    r["Date"] = formatdate(localtime=True)
    r["Message-ID"] = make_msgid(domain=address.split("@")[1])
    r["In-Reply-To"] = orig["Message-ID"]
    r["References"] = " ".join(
        x for x in [orig.get("References"), orig["Message-ID"]] if x)
    r.set_content(text)
    deliver(address, r)


def cmd_send(address, to, subject):
    text = sys.stdin.read()
    m = EmailMessage()
    m["From"] = address
    m["To"] = to
    m["Subject"] = subject
    m["Date"] = formatdate(localtime=True)
    m["Message-ID"] = make_msgid(domain=address.split("@")[1])
    m.set_content(text)
    deliver(address, m)


if __name__ == "__main__":
    if len(sys.argv) < 3:
        raise SystemExit(__doc__)
    action, rest = sys.argv[1], sys.argv[2:]
    {"list": cmd_list, "read": cmd_read, "reply": cmd_reply, "send": cmd_send}[action](*rest)
