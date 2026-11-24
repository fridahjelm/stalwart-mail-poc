"""Checks the pieces the office integration needs, as far as they can be
checked without a signed-in browser session.

Bulwark's and Nextcloud's own "is office editing available" endpoints require
a session, so the last step is a click: open the message with the .docx
attachment in Bulwark and see whether it renders.
"""
import socket
import ssl
import sys
import urllib.error
import urllib.request

_getaddrinfo = socket.getaddrinfo
socket.getaddrinfo = lambda h, p, *a, **k: _getaddrinfo(
    "127.0.0.1" if isinstance(h, str) and h.endswith(".lab.internal") else h, p, *a, **k)

CTX = ssl.create_default_context(cafile="certs/rootCA.pem")
ok = True


def report(name, passed, detail=""):
    global ok
    ok = ok and passed
    print(f"  [{'ok' if passed else 'FAIL'}] {name}{': ' + detail if detail else ''}")


def get(url, timeout=20):
    with urllib.request.urlopen(url, context=CTX, timeout=timeout) as r:
        return r.status, r.read().decode(errors="replace")


def main():
    print("collabora")
    discovery = ""
    try:
        status, discovery = get("https://office.lab.internal/hosting/discovery")
        body = discovery
        report("discovery answers", status == 200, f"HTTP {status}")
        # The browser loads the editor from this URL, so it has to be the
        # public name rather than a container address; that mismatch is what
        # breaks the editor behind a reverse proxy.
        report("urlsrc is the public name",
               'urlsrc="https://office.lab.internal/' in body,
               (body.split('urlsrc="', 1)[1].split('"', 1)[0]
                if 'urlsrc="' in body else "no urlsrc"))
        report("docx is a known format", "docx" in body)
    except urllib.error.URLError as exc:
        report("discovery answers", False, str(exc)[:80])

    print("nextcloud")
    try:
        status, body = get("https://cloud.lab.internal/status.php")
        report("reachable through caddy", status == 200 and '"installed":true' in body)
    except urllib.error.URLError as exc:
        report("reachable through caddy", False, str(exc)[:80])

    # Sign-in fails with "could not reach the OpenID Connect provider" both
    # when the lab CA is not trusted and when the SSRF guard refuses the
    # private address, and the browser cannot tell them apart. Asking the
    # login endpoint exercises the same path PHP takes: a 303 to Keycloak
    # means discovery succeeded. curl inside the container proves nothing,
    # because it has its own trust store. The discovery document is cached,
    # so the cache is flushed first.
    import subprocess
    subprocess.run(["docker", "exec", "lab-valkey", "valkey-cli", "flushall"],
                   capture_output=True, text=True)
    try:
        request = urllib.request.Request(
            "https://cloud.lab.internal/apps/user_oidc/login/1")
        opener = urllib.request.build_opener(
            urllib.request.HTTPSHandler(context=CTX),
            type("NoRedirect", (urllib.request.HTTPRedirectHandler,),
                 {"redirect_request": lambda *a, **k: None})())
        with opener.open(request, timeout=30) as r:
            status, target = r.status, r.headers.get("Location", "")
    except urllib.error.HTTPError as exc:
        status, target = exc.code, exc.headers.get("Location", "")
    report("sign-in reaches keycloak",
           status == 303 and "kc.lab.internal" in target,
           f"HTTP {status}" + (" (see data/nextcloud.log)" if status != 303 else ""))

    print("bulwark")
    try:
        status, _ = get("https://webmail.lab.internal/")
        report("reachable through caddy", status == 200, f"HTTP {status}")
    except urllib.error.URLError as exc:
        report("reachable through caddy", False, str(exc)[:80])

    print("the editor must be able to call the hosts back")
    import json
    import subprocess
    # The Collabora image has no shell, so nothing can be run inside it. Name
    # resolution is checked from another container on the same resolver, and
    # Collabora's own settings are read from its environment.
    for name in ("webmail.lab.internal", "cloud.lab.internal"):
        out = subprocess.run(
            ["docker", "exec", "lab-nextcloud", "getent", "hosts", name],
            capture_output=True, text=True)
        addr = out.stdout.split()[0] if out.stdout.split() else "no answer"
        report(f"{name} resolves to caddy on the lab network",
               out.returncode == 0 and addr == "172.28.0.40", addr)

    env = subprocess.run(
        ["docker", "inspect", "lab-collabora", "--format", "{{json .Config.Env}}"],
        capture_output=True, text=True)
    settings = " ".join(json.loads(env.stdout or "[]"))
    for host in ("webmail.lab.internal", "cloud.lab.internal"):
        report(f"{host} is an allowed wopi host", host in settings)

    # Read the policy Collabora actually sends rather than the setting it was
    # given: a value with a space in it can be truncated on the way in, which
    # leaves the editor working in one host's frame and blocked in the other's
    # with nothing logged on either side.
    urlsrc = (discovery.split('urlsrc="', 1)[1].split('"', 1)[0]
              if 'urlsrc="' in discovery else "")
    policy = ""
    if urlsrc:
        request = urllib.request.Request(urlsrc, method="HEAD")
        try:
            with urllib.request.urlopen(request, context=CTX, timeout=20) as r:
                policy = r.headers.get("Content-Security-Policy", "")
        except urllib.error.HTTPError as exc:
            policy = exc.headers.get("Content-Security-Policy", "")
    ancestors = next((part for part in policy.split(";")
                      if "frame-ancestors" in part), "")
    for host in ("webmail.lab.internal:443", "cloud.lab.internal:443"):
        report(f"{host} may frame the editor", host in ancestors,
               ancestors.strip() or "no frame-ancestors header")

    print("\nremaining check, in a browser:")
    print("  open 'Notes from the lab' as alice in https://webmail.lab.internal")
    print("  and click lab-notes.docx; it should render read-only in Collabora.")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
