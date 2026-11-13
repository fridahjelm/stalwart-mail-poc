"""Checks every authentication path and a full send/receive round trip."""
import base64
import imaplib
import socket
import json
import smtplib
import ssl
import sys
import time
import urllib.request
from email.message import EmailMessage

# Every lab name is published in the lab zone, which only the containers use;
# on the host they live in /etc/hosts, which scripts/00-prereqs.sh adds for the
# browser's sake. Mapping them here lets this check run before that step, while
# still presenting the real hostname for certificate validation and SNI.
_getaddrinfo = socket.getaddrinfo


def _lab_names(host, port, *args, **kwargs):
    if isinstance(host, str) and host.endswith(".lab.internal"):
        host = "127.0.0.1"
    return _getaddrinfo(host, port, *args, **kwargs)


socket.getaddrinfo = _lab_names

CA = "certs/rootCA.pem"
KC = "https://kc.lab.internal/realms/lab"
HOST = "mail.lab.internal"
ok = True


def report(name, passed, detail=""):
    global ok
    ok = ok and passed
    print(f"  [{'ok' if passed else 'FAIL'}] {name}{': ' + detail if detail else ''}")


def kc_token(user, password, client="bulwark", secret="bulwark-lab-secret"):
    body = urllib.parse.urlencode({
        "client_id": client, "client_secret": secret, "grant_type": "password",
        "username": user, "password": password, "scope": "openid",
    }).encode()
    ctx = ssl.create_default_context(cafile=CA)
    req = urllib.request.Request(f"{KC}/protocol/openid-connect/token", data=body)
    with urllib.request.urlopen(req, context=ctx, timeout=15) as r:
        return json.load(r)["access_token"]


def imap(port=993):
    return imaplib.IMAP4_SSL(HOST, port, ssl_context=ssl.create_default_context(cafile=CA))


def main():
    users = dict(
        line.split("\t") for line in
        open("app-passwords.txt").read().strip().splitlines()
    )
    demo = open(".env").read()
    demo_pw = [l.split("=", 1)[1] for l in demo.splitlines()
               if l.startswith("DEMO_USER_PASSWORD=")][0]

    print("keycloak")
    token = kc_token("alice", demo_pw)
    claims = json.loads(base64.urlsafe_b64decode(
        token.split(".")[1] + "=" * (-len(token.split(".")[1]) % 4)))
    report("issuer matches the lab URL", claims["iss"] == KC, claims["iss"])
    report("email claim present", claims.get("email") == "alice@lab.internal")

    print("jmap over the oidc directory")
    ctx = ssl.create_default_context(cafile=CA)
    req = urllib.request.Request("https://mail.lab.internal/jmap/session",
                                 headers={"Authorization": f"Bearer {token}"})
    with urllib.request.urlopen(req, context=ctx, timeout=15) as r:
        sess = json.load(r)
    report("keycloak token accepted by stalwart",
           sess.get("username") == "alice@lab.internal", sess.get("username", "?"))

    print("imap")
    m = imap()
    raw = f"n,a=alice@lab.internal,\x01auth=Bearer {token}\x01\x01".encode()
    try:
        m.authenticate("OAUTHBEARER", lambda _: raw)
        report("OAUTHBEARER with a keycloak token", True)
        m.logout()
    except Exception as exc:                                  # noqa: BLE001
        report("OAUTHBEARER with a keycloak token", False, str(exc)[:80])

    m = imap()
    try:
        m.login("alice@lab.internal", users["alice@lab.internal"])
        report("app password login", True)
        m.logout()
    except Exception as exc:                                  # noqa: BLE001
        report("app password login", False, str(exc)[:80])

    print("delivery")
    msg = EmailMessage()
    msg["From"] = "alice@lab.internal"
    msg["To"] = "bob@lab.internal"
    stamp = str(time.time())
    msg["Subject"] = f"lab round trip {stamp}"
    msg.set_content("sent by scripts/verify.py")
    s = smtplib.SMTP(HOST, 587, timeout=20)
    s.starttls(context=ssl.create_default_context(cafile=CA))
    s.login("alice@lab.internal", users["alice@lab.internal"])
    s.send_message(msg)
    s.quit()

    found = None
    for _ in range(20):
        time.sleep(1)
        b = imap()
        b.login("bob@lab.internal", users["bob@lab.internal"])
        b.select("INBOX")
        typ, data = b.search(None, "SUBJECT", f'"lab round trip {stamp}"')
        if typ == "OK" and data[0].split():
            num = data[0].split()[-1]
            found = b.fetch(num, "(BODY.PEEK[HEADER])")[1][0][1].decode(errors="replace")
            b.logout()
            break
        b.logout()
    report("alice -> bob delivered", found is not None)
    if found:
        report("message was dkim signed", "DKIM-Signature" in found)

    print("outbound containment")
    out = EmailMessage()
    out["From"] = "alice@lab.internal"
    out["To"] = "nobody@iana.org"
    out["Subject"] = f"containment {stamp}"
    out.set_content("must not leave the lab")
    s = smtplib.SMTP(HOST, 587, timeout=20)
    s.starttls(context=ssl.create_default_context(cafile=CA))
    s.login("alice@lab.internal", users["alice@lab.internal"])
    s.send_message(out)
    s.quit()

    caught = False
    for _ in range(15):
        time.sleep(1)
        with urllib.request.urlopen(
                "http://127.0.0.1:8025/api/v1/messages?limit=50", timeout=10) as r:
            msgs = json.load(r).get("messages", [])
        if any(m["Subject"] == f"containment {stamp}" for m in msgs):
            caught = True
            break
    report("mail for a non-lab domain went to mailpit, not the internet", caught)

    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
