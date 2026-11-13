#!/usr/bin/env python3
"""End-to-end test of the booking service.

Organizers create events in their own calendars over CalDAV, inviting a room,
and the test waits for the room's answer to appear in the organizer's copy.
Every event it creates is deleted again at the end, which also exercises the
service's removal of cancelled copies. Run from anywhere:

  python3 -I scripts/60-booking-test.py
"""
import base64
import datetime as dt
import email
import http.client
import imaplib
import os
import re
import socket
import ssl
import sys
import time
import urllib.parse
import uuid
from email.policy import default as POLICY
from zoneinfo import ZoneInfo

os.chdir(os.path.join(os.path.dirname(os.path.abspath(__file__)), ".."))

_getaddrinfo = socket.getaddrinfo
socket.getaddrinfo = lambda h, p, *a, **k: _getaddrinfo(
    "127.0.0.1" if isinstance(h, str) and h.endswith(".lab.internal") else h, p, *a, **k)

HOST = "mail.lab.internal"
CTX = ssl.create_default_context(cafile="certs/rootCA.pem")
TZ = ZoneInfo("Europe/Stockholm")
PW = {}
for path in ("app-passwords.txt", "booking-passwords.txt"):
    for line in open(path):
        if line.strip():
            addr, secret = line.rstrip("\n").split("\t")
            PW[addr] = secret

ALICE, BOB, CAROL = "alice@lab.internal", "bob@lab.internal", "carol@other.lab.internal"
ROOM_A, ROOM_B = "room-a@lab.internal", "room-b@lab.internal"
RUN = uuid.uuid4().hex[:6]
created = []  # (organizer, path)
failures = 0


def dav(user, method, path, body=None, headers=None):
    c = http.client.HTTPSConnection(HOST, 443, context=CTX, timeout=20)
    h = {"Authorization": "Basic " + base64.b64encode(f"{user}:{PW[user]}".encode()).decode()}
    h.update(headers or {})
    c.request(method, path, body=body, headers=h)
    r = c.getresponse()
    data = r.read().decode(errors="replace")
    etag = r.getheader("ETag")
    c.close()
    return r.status, data, etag


def cal(user):
    return f"/dav/cal/{urllib.parse.quote(user)}/default/"


def ics(uid, organizer, room, start, end, summary, seq=0, extra=()):
    people = [organizer, *extra]
    lines = ["BEGIN:VCALENDAR", "VERSION:2.0", "PRODID:-//lab.internal//booking-test//EN",
             "BEGIN:VEVENT", f"UID:{uid}",
             "DTSTAMP:" + dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ"),
             "DTSTART;TZID=Europe/Stockholm:" + start.strftime("%Y%m%dT%H%M%S"),
             "DTEND;TZID=Europe/Stockholm:" + end.strftime("%Y%m%dT%H%M%S"),
             f"SUMMARY:{summary}", f"ORGANIZER:mailto:{organizer}"]
    lines += [f"ATTENDEE;PARTSTAT={'ACCEPTED' if p == organizer else 'NEEDS-ACTION'}:mailto:{p}" for p in people]
    lines += [f"ATTENDEE;CUTYPE=ROOM;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:mailto:{room}",
              f"SEQUENCE:{seq}", "END:VEVENT", "END:VCALENDAR", ""]
    return "\r\n".join(lines)


def book(organizer, room, start, end, summary, extra=()):
    uid = f"booking-test-{RUN}-{uuid.uuid4().hex[:8]}@lab.internal"
    path = cal(organizer) + uid.split("@")[0] + ".ics"
    status, _, _ = dav(organizer, "PUT", path, ics(uid, organizer, room, start, end, summary, 0, extra),
                       {"Content-Type": "text/calendar", "If-None-Match": "*"})
    assert status == 201, f"PUT {path}: {status}"
    created.append((organizer, path))
    return uid, path


def room_answer(organizer, path, room, want, timeout=25):
    """Wait until the room's PARTSTAT in the organizer's copy is `want`; return the last value seen."""
    deadline, seen = time.time() + timeout, "NEEDS-ACTION"
    while time.time() < deadline:
        _, body, _ = dav(organizer, "GET", path)
        body = re.sub(r"\r?\n[ \t]", "", body)
        for line in body.splitlines():
            if line.startswith("ATTENDEE") and line.lower().endswith("mailto:" + room):
                m = re.search(r"PARTSTAT=([A-Z-]+)", line)
                seen = m.group(1) if m else seen
        if seen == want:
            return seen
        time.sleep(1)
    return seen


def reschedule(organizer, path, start, end, seq, ask_again=None):
    """Move an event. ask_again=<room> also resets that room's answer to NEEDS-ACTION,
    which is how a client re-invites an attendee; Stalwart sends nothing to an
    attendee who declined unless it is asked again."""
    _, body, etag = dav(organizer, "GET", path)
    if ask_again:
        body = re.sub(r"\r?\n[ \t]", "", body)
        body = re.sub(rf"^(ATTENDEE[^\r\n]*?)PARTSTAT=[A-Z-]+([^\r\n]*mailto:{re.escape(ask_again)})",
                      r"\1PARTSTAT=NEEDS-ACTION;RSVP=TRUE\2", body, flags=re.M | re.I)
    body = re.sub(r"(DTSTART;TZID=Europe/Stockholm:)\d{8}T\d{6}", r"\g<1>" + start.strftime("%Y%m%dT%H%M%S"), body)
    body = re.sub(r"(DTEND;TZID=Europe/Stockholm:)\d{8}T\d{6}", r"\g<1>" + end.strftime("%Y%m%dT%H%M%S"), body)
    body = re.sub(r"SEQUENCE:\d+", f"SEQUENCE:{seq}", body)
    status = dav(organizer, "PUT", path, body, {"Content-Type": "text/calendar", "If-Match": etag})[0]
    assert status in (200, 204), f"PUT {path}: {status}"


def check(name, got, want):
    global failures
    ok = got == want
    failures += not ok
    print(f"{'ok  ' if ok else 'FAIL'} {name}: {got}" + ("" if ok else f" (expected {want})"))


def decline_mail(organizer, summary, timeout=15):
    deadline = time.time() + timeout
    while time.time() < deadline:
        c = imaplib.IMAP4_SSL(HOST, 993, ssl_context=CTX)
        c.login(organizer, PW[organizer])
        c.select("INBOX")
        ids = c.search(None, "SUBJECT", f'"declined: {summary}"')[1][0].split()
        if ids:
            m = email.message_from_bytes(c.fetch(ids[-1], "(RFC822)")[1][0][1], policy=POLICY)
            c.logout()
            return m.get_body(("plain",)).get_content().strip()
        c.logout()
        time.sleep(1)
    return None


# A Tuesday at least a week ahead, inside the 90-day horizon.
day = dt.datetime.now(TZ).date() + dt.timedelta(days=7)
day += dt.timedelta(days=(1 - day.weekday()) % 7)
at = lambda h, m=0: dt.datetime(day.year, day.month, day.day, h, m, tzinfo=TZ)
print(f"test day {day}, run {RUN}")

try:
    _, a_path = book(ALICE, ROOM_A, at(10), at(11), f"Planning {RUN}")
    check("1 Alice books room-a 10:00-11:00", room_answer(ALICE, a_path, ROOM_A, "ACCEPTED"), "ACCEPTED")

    _, b_path = book(BOB, ROOM_A, at(10, 30), at(11, 30), f"Review {RUN}")
    check("2 Bob books room-a 10:30-11:30 (overlap)", room_answer(BOB, b_path, ROOM_A, "DECLINED"), "DECLINED")
    reason = decline_mail(BOB, f"Review {RUN}")
    check("2 Bob gets the reason by mail", bool(reason and "already booked 10:00–11:00" in reason), True)

    dav(ALICE, "DELETE", a_path)
    created.remove((ALICE, a_path))
    time.sleep(8)
    _, b2_path = book(BOB, ROOM_A, at(10, 30), at(11, 30), f"Review again {RUN}")
    check("3 after Alice cancels, Bob books 10:30-11:30",
          room_answer(BOB, b2_path, ROOM_A, "ACCEPTED"), "ACCEPTED")

    _, c_path = book(ALICE, ROOM_A, at(13), at(14), f"Moved {RUN}")
    check("4 Alice books room-a 13:00-14:00", room_answer(ALICE, c_path, ROOM_A, "ACCEPTED"), "ACCEPTED")
    reschedule(ALICE, c_path, at(11), at(12), 1)
    check("4 Alice moves it to 11:00-12:00 (overlaps Bob)",
          room_answer(ALICE, c_path, ROOM_A, "DECLINED"), "DECLINED")

    time.sleep(6)
    reschedule(ALICE, c_path, at(15), at(16), 2, ask_again=ROOM_A)
    check("5 Alice moves the declined meeting to 15:00-16:00 and asks the room again",
          room_answer(ALICE, c_path, ROOM_A, "ACCEPTED"), "ACCEPTED")

    _, d_path = book(CAROL, ROOM_B, at(9), at(10), f"Workshop {RUN}",
                     extra=[ALICE, BOB, "vera@lab.internal", "chris@lab.internal"])
    check("6 Carol books room-b (seats 4) for 5 people",
          room_answer(CAROL, d_path, ROOM_B, "DECLINED"), "DECLINED")
finally:
    for organizer, path in created:
        dav(organizer, "DELETE", path)
    time.sleep(8)
    # Attendees keep a cancelled copy of each invitation; remove this run's.
    for user in PW:
        if user.startswith("room-"):
            continue
        _, data, _ = dav(user, "PROPFIND", cal(user),
                         '<propfind xmlns="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">'
                         '<prop><C:calendar-data/></prop></propfind>', {"Depth": "1"})
        for href, body in re.findall(r"<D:href>([^<]*\.ics)</D:href>(.*?)</D:response>", data, re.S):
            if f"booking-test-{RUN}" in body:
                dav(user, "DELETE", href, None, {"Schedule-Reply": "F"})
    leftovers = 0
    for room in (ROOM_A, ROOM_B):
        _, data, _ = dav(room, "PROPFIND", cal(room), '<propfind xmlns="DAV:"><prop><getetag/></prop></propfind>',
                         {"Depth": "1"})
        leftovers += len([h for h in re.findall(r"<D:href>([^<]*\.ics)</D:href>", data)])
    print(f"cleanup: {len(created)} test events deleted by their organizers")

check("7 room calendars hold no test copies after cleanup", leftovers, 0)
sys.exit(1 if failures else 0)
