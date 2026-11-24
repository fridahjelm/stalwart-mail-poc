"""Checks how Stalwart answers invitations sent to the room resource accounts.

Books room-a as Bob, books an overlapping slot, and books room-b for more
people than it seats. Prints each room's PARTSTAT and removes the events it
created. Run it with the lab up and after scripts/30-stalwart-bootstrap.sh.
"""
import base64
import re
import socket
import ssl
import sys
import time
import urllib.error
import urllib.request
import uuid

# The lab names are in /etc/hosts for the browser's sake; mapping them here too
# lets this run before scripts/00-prereqs.sh, while still presenting the real
# hostname for certificate validation.
_getaddrinfo = socket.getaddrinfo
socket.getaddrinfo = lambda h, p, *a, **k: _getaddrinfo(
    "127.0.0.1" if isinstance(h, str) and h.endswith(".lab.internal") else h, p, *a, **k)

BASE = "https://mail.lab.internal"
CTX = ssl.create_default_context(cafile="certs/rootCA.pem")
ORGANIZER = "bob@lab.internal"
DAY = time.strftime("%Y%m%d", time.gmtime(time.time() + 14 * 86400))
ok = True


def secret(address):
    for line in open("app-passwords.txt"):
        if line.strip():
            who, pw = line.rstrip("\n").split("\t")
            if who == address:
                return pw
    raise SystemExit(f"no app password for {address}; run scripts/35-app-passwords.sh")


def dav(method, path, body=None, headers=None):
    auth = base64.b64encode(f"{ORGANIZER}:{secret(ORGANIZER)}".encode()).decode()
    h = {"Authorization": f"Basic {auth}", "Content-Type": "text/calendar; charset=utf-8"}
    h.update(headers or {})
    req = urllib.request.Request(BASE + path, method=method,
                                 data=body.encode() if body else None, headers=h)
    try:
        with urllib.request.urlopen(req, context=CTX, timeout=20) as r:
            return r.status, r.read().decode(errors="replace")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode(errors="replace")


def book(room, start, end, guests=()):
    uid = str(uuid.uuid4())
    extra = "".join(
        f"ATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:{g}\r\n"
        for g in guests)
    ics = (
        "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//lab//room-booking-test//EN\r\n"
        "METHOD:REQUEST\r\nBEGIN:VEVENT\r\n"
        f"UID:{uid}\r\nDTSTAMP:{start}\r\nDTSTART:{start}\r\nDTEND:{end}\r\n"
        f"SUMMARY:room booking test {uid[:8]}\r\n"
        f"ORGANIZER:mailto:{ORGANIZER}\r\n"
        f"ATTENDEE;ROLE=CHAIR;PARTSTAT=ACCEPTED:mailto:{ORGANIZER}\r\n"
        f"{extra}"
        "ATTENDEE;CUTYPE=ROOM;ROLE=REQ-PARTICIPANT;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:"
        f"mailto:{room}\r\n"
        "END:VEVENT\r\nEND:VCALENDAR\r\n")
    path = f"/dav/cal/{ORGANIZER.replace('@', '%40')}/default/{uid}.ics"
    status, _ = dav("PUT", path, ics)
    if status not in (200, 201, 204):
        raise SystemExit(f"could not create the event: HTTP {status}")
    return path


def answer(path, room, timeout=20):
    """The room's PARTSTAT on the organizer's copy, once it stops being NEEDS-ACTION."""
    for _ in range(timeout):
        time.sleep(1)
        _, body = dav("GET", path)
        # ICS folds long lines; unfold before matching the attendee.
        for line in body.replace("\r\n ", "").replace("\n ", "").splitlines():
            if line.startswith("ATTENDEE") and room in line:
                match = re.search(r"PARTSTAT=([A-Z-]+)", line)
                if match and match.group(1) != "NEEDS-ACTION":
                    return match.group(1)
    return "NEEDS-ACTION"


def report(name, passed, detail=""):
    global ok
    ok = ok and passed
    print(f"  [{'ok' if passed else 'FAIL'}] {name}{': ' + detail if detail else ''}")


def main():
    created = []
    try:
        print("automatic booking")
        first = book("room-a@lab.internal", f"{DAY}T090000Z", f"{DAY}T100000Z")
        created.append(first)
        got = answer(first, "room-a@lab.internal")
        report("free slot accepted", got == "ACCEPTED", got)

        second = book("room-a@lab.internal", f"{DAY}T093000Z", f"{DAY}T103000Z")
        created.append(second)
        got = answer(second, "room-a@lab.internal")
        report("overlapping slot declined", got == "DECLINED", got)

        print("capacity")
        # Plus-addressed so that the count exceeds the room's capacity using
        # only the accounts the bootstrap creates; the room answers on the
        # number of attendees, not on whether they exist.
        guests = ["alice@lab.internal", "carol@other.lab.internal",
                  "alice+one@lab.internal", "alice+two@lab.internal",
                  "alice+three@lab.internal"]
        over = book("room-b@lab.internal", f"{DAY}T140000Z", f"{DAY}T150000Z", guests)
        created.append(over)
        got = answer(over, "room-b@lab.internal")
        # room-b seats 4. Capacity is metadata, not a booking rule, so Stalwart
        # accepts this; the check records the behaviour rather than asserting a fix.
        report("6 people in a 4-seat room still accepted", got == "ACCEPTED", got)
    finally:
        for path in created:
            dav("DELETE", path)
        print(f"removed {len(created)} test event(s)")

    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
