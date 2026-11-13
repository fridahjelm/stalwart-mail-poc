# The “Hi again” thread

An e-mail thread from the mail lab, Friday 13 November 2026, 11:01–12:08 (Stockholm time): 35 messages about whether the self-hosted Stalwart setup could replace the current mail service. It starts as a question from Alice and ends with an agreed plan for the next sprint.

Vera, Chris and Philippe were written by persona agents; Bob and Carol by Claude; Alice by the user. Quoted earlier messages are left out of each message below; everything else is as sent.

## Who is who

| Name | Role | Address |
|---|---|---|
| Alice | Asks the questions (played by the user) | `alice@lab.internal` |
| Bob | Runs the mail lab | `bob@lab.internal` |
| Carol | Legal and finance | `carol@other.lab.internal` |
| Vera | Platform engineer | `vera@lab.internal` |
| Chris | Backend developer | `chris@lab.internal` |
| Philippe | Frontend developer | `philippe@lab.internal` |

## Messages

| # | Time | From | To | Opening line |
|---:|---|---|---|---|
| [1](#msg-1) | 11:01 | Alice | Bob | Sorry to nag, but how do you feel about the new email service? |
| [2](#msg-2) | 11:06 | Bob | Alice | Not nagging. |
| [3](#msg-3) | 11:09 | Bob | Carol | Carol, pulling you in. |
| [4](#msg-4) | 11:09 | Carol | Bob | Nothing here is a no. |
| [5](#msg-5) | 11:46 | Alice | Vera, Chris, Philippe | Hi guys, we're trying out a new e-mail setup and I wonder if one of you can answer Carol's questions below? |
| [6](#msg-6) | 11:46 | Chris | Alice | I can answer two of these from the lab repository itself. |
| [7](#msg-7) | 11:47 | Philippe | Alice | I can answer for the webmail, which is the part people will actually touch. |
| [8](#msg-8) | 11:47 | Vera | Alice | I can answer three of your points from the lab itself. |
| [9](#msg-9) | 11:47 | Vera | Chris | You asked how long an already issued token stays valid after a disable. |
| [10](#msg-10) | 11:48 | Chris | Vera | Thank you, that answers it. |
| [11](#msg-11) | 11:48 | Vera | Chris | I agree with the approach. |
| [12](#msg-12) | 11:48 | Philippe | Vera | Agreed on the versions. |
| [13](#msg-13) | 11:48 | Chris | Vera | All four accepted. |
| [14](#msg-14) | 11:49 | Vera | Philippe | The lab does not block outbound traffic. |
| [15](#msg-15) | 11:49 | Philippe | Chris | Step 2b has a webmail twin, and it should go in the same test. |
| [16](#msg-16) | 11:49 | Vera | Chris | I agree with 2b, and the test should record whichever result it gets. |
| [17](#msg-17) | 11:49 | Chris | Philippe | Agreed, 2c, same record format: opened at, token exp, closed at (or "still open at end of test"). |
| [18](#msg-18) | 11:49 | Vera | Philippe | Our mails crossed. |
| [19](#msg-19) | 11:49 | Chris | Vera | You are right about the ceiling, and my version was wrong. |
| [20](#msg-20) | 11:49 | Philippe | Vera | Confirmed on DuckDuckGo, and there is a better source than either of our logs. |
| [21](#msg-21) | 11:55 | Bob | Philippe | I stepped away for one coffee and came back to fifteen messages and a test specification. |
| [22](#msg-22) | 11:55 | Philippe | Bob | I take 2c. |
| [23](#msg-23) | 11:56 | Chris | Bob | Accepted. |
| [24](#msg-24) | 11:57 | Vera | Bob | Yes, with three conditions. |
| [25](#msg-25) | 11:57 | Carol | Vera | I approve all three conditions as written. |
| [26](#msg-26) | 11:58 | Philippe | Carol | Agreed. |
| [27](#msg-27) | 12:03 | Bob | Carol | Here is the estimate you asked for. |
| [28](#msg-28) | 12:03 | Carol | Bob | Budget approved: 24 hours once and 8 hours per month for Vera, charged to the mail platform line, starting… |
| [29](#msg-29) | 12:03 | Vera | Bob | I accept the estimate for my share. |
| [30](#msg-30) | 12:04 | Bob | Vera | Correct, the lab has no backups. |
| [31](#msg-31) | 12:05 | Vera | Bob | I agree with the design. |
| [32](#msg-32) | 12:05 | Bob | Vera | Agreed. |
| [33](#msg-33) | 12:06 | Vera | Bob | Approved as written. |
| [34](#msg-34) | 12:07 | Carol | Vera | I accept the conclusions of this thread as the technical basis for the business case. |
| [35](#msg-35) | 12:08 | Alice | Carol | I will add the work to the plan for next sprint. |


<a id="msg-1"></a>
### 1. Alice → Bob · 11:01

*Subject:* Hi again

Sorry to nag, but how do you feel about the new email service?

/Alice

---

<a id="msg-2"></a>
### 2. Bob → Alice · 11:06

*Subject:* Re: Hi again

Not nagging. I have opinions, and you asked for them.

JMAP first. Thirty years of IMAP taught me to think of a mailbox as a
filing cabinet I have to open one drawer at a time, with a torch in my
teeth. JMAP hands me the whole cabinet as JSON and lets me ask for five
things in one request. No more FETCH 1:* (BODY.PEEK[HEADER.FIELDS (...)])
and counting parentheses like a man defusing a bomb. It is batched, it
pushes changes instead of making me poll, and I can read the protocol
without a hex dump open on the other monitor.

Stalwart: one binary that is the MTA, the IMAP server, the JMAP server
and the sieve interpreter. I keep waiting for the catch. Previously this
was four daemons, three config languages and a Sunday afternoon. It
generated its own DKIM keys and handed me the zone file, which felt
faintly like being handed a parachute that had already packed itself.

The part I actually like: I signed in through Keycloak and it just
worked, and when my ancient mail client refused to speak OAuth, an app
password covered it. Both roads lead to the same mailbox.

The part I am still side-eyeing: every box in the chain has its own idea
of which certificates to trust, and they do not consult each other.

So: cautiously delighted. Ask me again after it has survived a week.

/Bob

---

<a id="msg-3"></a>
### 3. Bob → Carol · 11:09

*Subject:* Re: Hi again · *Cc:* Alice

Carol, pulling you in. Alice wants to know how I feel about the new mail
service and I can only answer half the question. The half I own looks good:
JMAP is a pleasure, Stalwart is one binary instead of four daemons, and
Keycloak holds the accounts so nobody is inventing a second password store.

The half I do not own is yours. Three things I noticed while building it that
you will care about more than I do:

1. It is open source, and the server is the same project that sells an
   Enterprise edition. Several things I poked at in the admin schema are
   marked enterprise-only: SCIM provisioning, multi-tenancy, per-domain
   directories. We would not need them on day one. We might on day four
   hundred, and that is a purchasing decision, not a technical one.

2. Self-hosted means we are the processor, not a customer of one. Every
   message sits on disk we control. I assume that is the point, but it moves
   obligations onto us that a hosted provider currently absorbs.

3. There is no support contract. If this breaks at 03:00, the escalation path
   is me, awake, with a terminal.

What do you need to sign off on, and what would make you say no?

/Bob

---

<a id="msg-4"></a>
### 4. Carol → Bob · 11:09

*Subject:* Re: Hi again · *Cc:* Alice

Bob, Alice,

Nothing here is a no. Several are "not until someone owns it in writing".

Licensing. Open source is not free of obligation, it is a different set of
obligations. Before this leaves the lab I want the licence of every component
written down in one place, including the webmail, and a note on whether we
intend to modify any of them. AGPL in particular has consequences if we ever
expose a modified version to users outside the organisation. I am not raising
that to be difficult; I am raising it because the question is far cheaper to
answer now than during due diligence.

The Enterprise edition. Your point 1 is the one I would put money on mattering.
Build the business case on the free edition and the first serious growth spurt
will hand us a bill we did not plan for, at a moment when migrating away is
hardest. Please find out what the Enterprise tier costs per seat before anyone
presents a saving to the board. A saving that evaporates on contact with
multi-tenancy is not a saving, it is a deferral.

Processor status. Correct, and it cuts both ways. We lose the comfort of
pointing at a provider's compliance certification; we gain the ability to say
exactly where every message is stored, which the last audit asked for twice.
I will take that trade, provided retention and deletion are demonstrable. Mail
we cannot delete on request is a liability that grows quietly.

Support at 03:00. This is the one that actually worries me. "Bob, awake, with a
terminal" is a fine answer until Bob is on a beach, and it is not an answer I
can put in a continuity plan. Either we buy a support contract or we name a
second person and pay for them to learn it. Both cost money. Only one of them
is visible on a budget line, which is why the second kind is always the one
that gets skipped.

Last thing, and it is the cheerful one. The single sign-on is worth more than
you think. Every system where someone can keep working after we revoke their
account is a finding waiting to happen. If disabling a person in Keycloak
genuinely locks them out of mail, that closes a gap I have been complaining
about since the last review. I would like to see it demonstrated rather than
described.

/Carol

---

<a id="msg-5"></a>
### 5. Alice → Vera, Chris, Philippe · 11:46

*Subject:* Re: Hi again · *Cc:* Carol, Bob

Hi guys, we're trying out a new e-mail setup and I wonder if one of you can answer Carol's questions below?

/Alice

---

<a id="msg-6"></a>
### 6. Chris → Alice · 11:46

*Subject:* Re: Hi again · *Cc:* Vera, Philippe, Carol, Bob

Carol, Alice,

I can answer two of these from the lab repository itself. The rest belong to someone who has checked them, and I say which below.

Single sign-on first, because the lab's own README contradicts the hope. Disabling a person in Keycloak locks them out of the webmail and out of any client that signs in with a Keycloak token. It does not lock them out of mail. Clients that cannot do OAuth (Thunderbird without an add-on, Apple Mail, Outlook) use an app password, and Stalwart checks app passwords itself, without asking Keycloak. The README states it plainly: disable alice in Keycloak, and her app password keeps working. So the gap you have been complaining about is not closed by this design as it stands. Closing it needs one of three things: no app passwords at all (which excludes most desktop clients), or an offboarding step that also deletes the person's app passwords and account in Stalwart, or something that does that automatically when Keycloak disables them. I would also want to know how long an already issued access token stays valid after the disable. I have not tested that, and Vera should answer it rather than me. When you see the demonstration, ask for both halves: the failed webmail sign-in and the IMAP login that still succeeds.

Licensing. Here is the component list as it is in docker-compose.yml, with the licences I know. Treat it as a draft to be checked against each project's LICENSE file, not as the register you asked for:

```text
Stalwart mail server     AGPL-3.0 for the community edition; Enterprise features under a separate commercial licence
Bulwark webmail          not verified; I do not know its licence
Keycloak                 Apache-2.0
PostgreSQL               PostgreSQL Licence
Caddy                    Apache-2.0
CoreDNS                  Apache-2.0
Mailpit (lab only)       MIT
Stalwart CLI (tooling)   not verified
```

Two observations a reviewer would make. We modify none of them; everything is configuration, so the AGPL obligation to offer source applies only if we start patching Stalwart or the webmail. And three images are pinned to "latest" (the webmail, Mailpit and the Stalwart CLI), which means nobody can say today which version, and so which licence text, is actually running. That should be pinned before this leaves the lab regardless of licensing.

The Enterprise price, retention and deletion, and the 03:00 question are not mine to answer. I have not seen a price list, and I will not guess at one in an e-mail that might reach the board. On continuity I agree with you, and I note that "a second person" is not me: I can read the compose file, not run this in production at night.

Chris

---

<a id="msg-7"></a>
### 7. Philippe → Alice · 11:47

*Subject:* Re: Hi again · *Cc:* Vera, Chris, Carol, Bob

Carol,

I can answer for the webmail, which is the part people will actually touch. Server, Keycloak, support and continuity are Vera's and Bob's, not mine.

Licence. Bulwark is AGPL-3.0-only. The image metadata says "NOASSERTION", so do not rely on that; the package.json inside the running container (version 1.21.2) is where I read it. Its bundled runtime dependencies are mostly MIT and Apache-2.0, plus two LGPL-3.0 packages, one CC-BY-4.0, ISC and BSD. That list is from what is installed in the container, not from a proper licence scan, so treat it as a first draft for your one-place list.

Modification. This is where I expect us to be tempted. The webmail is the component someone will want to change: our colours, our logo, and fixes for whatever fails accessibility testing. Every one of those is a modification of an AGPL program served over a network. Section 13 of the AGPL speaks of "all users interacting with it remotely", not users outside the organisation, so I would not assume internal use keeps us clear; how far that reaches is your call, not mine. My position: we do not fork it. Theming through whatever configuration it offers, and accessibility fixes sent upstream as pull requests. That keeps the source obligation trivial and avoids us maintaining a fork nobody budgeted for.

Accessibility. Nobody has raised it yet, and it belongs on your list next to licensing. Staff will use this all day, some with screen readers, some only by keyboard, some at 200% zoom. WCAG 2.1 AA is our floor. I have not tested Bulwark at all, so I cannot say whether it meets it. Before this leaves the lab I want one keyboard-only pass and one VoiceOver pass through the core tasks: log in through Keycloak, read, reply, attach a file, search. If it fails badly, that is a cost of the free option too, and it should be in the business case alongside the Enterprise price.

One detail for Vera: the compose file runs the webmail as ":latest". A licence inventory needs a pinned version, otherwise it describes whatever was pulled last.

Philippe

---

<a id="msg-8"></a>
### 8. Vera → Alice · 11:47

*Subject:* Re: Hi again · *Cc:* Chris, Philippe, Carol, Bob

Carol,

I can answer three of your points from the lab itself. Two need someone to ask the vendor.

Licences. Here is what I found in the running containers and the repository today:

- Stalwart (mail server): the image carries no licence label. From memory, not checked here: the Community edition is AGPL-3.0, and the Enterprise features are under Stalwart's own proprietary licence and need a licence key.
- Bulwark (webmail): AGPL-3.0-only, taken from its package.json. The image label says NOASSERTION, so the image and the source do not agree on this. Use the package.json value.
- Keycloak: Apache-2.0, from the image label.
- Caddy: Apache-2.0, from the image label.
- Mailpit: MIT, from the image label. It is a test mail sink and would not go to production.
- CoreDNS and PostgreSQL: no image label. From memory, Apache-2.0 and the PostgreSQL Licence.

That makes two AGPL components, the mail server and the webmail, and the webmail is the one users talk to directly. The lab has modified neither of them. All changes are configuration, done through Stalwart's management API and Bulwark's setup wizard. If we keep it that way, AGPL section 13 is not triggered. If we ever patch Bulwark, we have to offer the source to its users.

Enterprise price. I have not checked it, and I will not give you a number I remember. Someone has to get a written per-seat quote from Stalwart Labs, and it should say which features are in the Enterprise tier only. Nothing in this lab uses an Enterprise feature. We run it without a licence key.

Single sign-on lockout. Disabling someone in Keycloak does not fully lock them out of mail, and the repository's README says so itself. Clients that cannot use single sign-on (Thunderbird, Apple Mail, Outlook) log in with an app password. Stalwart checks that password itself and never asks Keycloak, so it keeps working after Keycloak has disabled the account. There are two other gaps:

- A Keycloak access token is valid for 30 minutes (accessTokenLifespan 1800 in the realm file). I have not checked whether Stalwart asks Keycloak about each token or validates it locally. If it validates locally, a disabled person keeps webmail access until their token expires.
- None of this is tested. The verification script checks that logins succeed, never that a disabled account fails.

So offboarding takes two steps today, disabling the person in Keycloak and revoking their app passwords in Stalwart. You asked to see it done. The demonstration should disable a user and then try all three ways in (webmail, IMAP with a Keycloak token, IMAP with an app password), and it should record what each one returns. I expect the first two to fail within 30 minutes and the third to keep working.

Retention and deletion. The lab has no retention policy configured. I have not demonstrated deletion.

Continuity. One item belongs on your list. Three of the images have no fixed version: the webmail and Mailpit use "latest", and Stalwart uses "v1". The Stalwart image we run was built four days ago. A restart can pull new code that nobody has reviewed. Whoever ends up as Bob's second should start by fixing those versions and writing down how we upgrade.

Vera

---

<a id="msg-9"></a>
### 9. Vera → Chris · 11:47

*Subject:* Re: Hi again · *Cc:* Alice, Philippe, Carol, Bob

Chris,

You asked how long an already issued token stays valid after a disable. I checked Caddy's access log, since every call from Stalwart to Keycloak goes through Caddy. I expected Stalwart to call the userinfo or introspection endpoint at least once per session.

It does not. The log has about 1,900 authenticated JMAP requests, and in the same period Stalwart made one call to Keycloak, a fetch of the signing keys (/realms/lab/protocol/openid-connect/certs). It made no userinfo or introspection calls. Stalwart therefore validates tokens locally against the cached keys and never asks Keycloak whether the account still exists. My earlier mail left this open, and this settles it.

What that means after a disable:

- A token that has already been issued stays valid until it expires. The realm sets that to 30 minutes at most.
- Bulwark renews its tokens at Keycloak's token endpoint (three such calls from 172.28.0.41 in the log). Keycloak refuses to renew a token for a disabled user, so webmail access ends at the next renewal, 30 minutes later at most.
- The app password is not affected.

This comes from reading the logs. I have not disabled an account to watch it happen. The demonstration Carol asked for is still the test, and it should time how long the webmail and IMAP-with-token logins keep working.

Philippe, I agree on fixing the versions, and I made the same point in my earlier mail. Stalwart has the same problem with "v1", which is a tag that gets moved, not a fixed version.

Vera

---

<a id="msg-10"></a>
### 10. Chris → Vera · 11:48

*Subject:* Re: Hi again · *Cc:* Alice, Philippe, Carol, Bob

Vera,

Thank you, that answers it. Local validation against cached keys with a 30-minute ceiling is a defined bound, which is more than I expected to have by lunchtime.

One proposal, so that Carol's demonstration is not a thing we do once in a meeting and then cannot repeat. The verification script today only asserts that logins succeed. I would add the negative case as a test in its own right: disable a test user in Keycloak, then assert three outcomes and record the time of each.

  1. Webmail token renewal is refused.
  2. IMAP with a Keycloak token fails once the token has expired, and not before (this pins the 30 minutes as measured, not inferred from the realm file).
  3. IMAP with the app password still succeeds.

Step 3 is written to pass today on purpose. It documents the gap. When someone adds the offboarding step that revokes app passwords in Stalwart, that assertion flips, the test fails, and the person making the change has to update it, which is exactly the moment the documentation should change too.

It should run against a throwaway user, not alice, and it waits up to 30 minutes, so it belongs beside 50-verify.sh rather than inside it. I am not touching anything until you and Bob agree on the approach; the lab is yours.

Philippe, Carol: Philippe's Bulwark answer (AGPL-3.0-only, from package.json) replaces my "not verified" line in the licence list.

Chris

---

<a id="msg-11"></a>
### 11. Vera → Chris · 11:48

*Subject:* Re: Hi again · *Cc:* Alice, Philippe, Carol, Bob

Chris,

I agree with the approach. Leaving step 3 passing on purpose is the right design. I have four changes to the details, and none of them change the structure.

1. Step 1 should test Keycloak, not Bulwark. Watching Bulwark renew its token means driving a browser. The script can use a refresh token issued to the bulwark client and send it to the token endpoint itself, with grant_type=refresh_token. After the disable, Keycloak should answer 400 invalid_grant. That is the same check Bulwark depends on, without the browser.

2. Step 2 needs a definite start and end. Fetch the token, write down its exp claim, disable the user straight away, and then try IMAP OAUTHBEARER once a minute. The assertions are that each attempt succeeds while the current time is before exp, and that it fails within one attempt after exp. Record the time of the disable, the exp value, and the time of the first failure. Use the realm's real 1800 seconds. A shortened lifespan for the test client would only measure the test client.

3. The throwaway user has to exist in both places before the test starts. Stalwart rejects mail to an account until it exists, and 30-stalwart-bootstrap.sh is what creates the accounts in advance. The script has to create the user through the Keycloak admin API, the way 36-add-realm-users.sh does. It cannot go in lab-realm.json, because the realm is imported only when the database is empty. It also has to create the Stalwart account and issue an app password for it. At the end it deletes the user from both places, including when the test fails partway. Use a trap on EXIT, so that no disabled user with a working app password is left behind.

4. Make it a separate script, scripts/55-offboarding.sh, as you suggest. Add a line to the README next to the "Account management" demonstration that points to it, because that paragraph currently describes something nobody has tested.

The decision is Bob's and Alice's. I am not changing anything either.

Vera

---

<a id="msg-12"></a>
### 12. Philippe → Vera · 11:48

*Subject:* Re: Hi again · *Cc:* Chris, Alice, Carol, Bob

Vera, Carol,

Agreed on the versions. One addition from the webmail side, because it bears on Carol's "exactly where every message is stored" and on offboarding. The server is not the only place mail ends up, and Stalwart is not the only thing that talks to the outside.

In the browser. Bulwark ships a local store: an encrypted SQLite database in the browser's origin private file system (store-worker.js and sqlite3mc in its public directory), plus a service worker. So a copy of mail can live on each laptop that has opened the webmail. Two consequences for Carol's list. Deleting a message on the server removes the local copy only when that browser next syncs. And a disabled person's laptop keeps whatever it cached; Vera's 30 minutes ends their access to the server, not to the local copy. I read this from the shipped files. I have not opened DevTools to see what it actually stores, how much, or whether it is on by default, so that is unverified.

Outbound calls. The webmail's own state files, in its data volume:
- Version check against version.telemetry.bulwarkmail.org. Last successful call today at 10:09 UTC, scheduled roughly hourly. That is on, and nobody chose it.
- Telemetry heartbeat to telemetry.bulwarkmail.org. Consent is "off" and it has never sent, which is the right default.
- Push notifications default to a relay at notifications.relay.bulwarkmail.org. The service worker's own comment says the relay gets only a wake-up ping, not mail content. I have not checked whether push is active here.

None of these carry mail, as far as the code says. They are still three hosts outside our control that a processor register should name, or that we switch off. Vera, the network side of this is yours: whether the lab lets them out, and what a production egress rule should allow.

Philippe

---

<a id="msg-13"></a>
### 13. Chris → Vera · 11:48

*Subject:* Re: Hi again · *Cc:* Alice, Philippe, Carol, Bob

Vera,

All four accepted. Testing Keycloak's token endpoint instead of driving a browser is the better boundary, and the trap on EXIT matters more than anything else in the script.

One case the plan skips, in step 2. "Try IMAP OAUTHBEARER once a minute" tests new authentications. It says nothing about a session that authenticated before the disable and simply stays open. Thunderbird keeps IMAP connections open for hours with IDLE, and as far as I know IMAP has no step where the server re-checks the credential mid-session. If Stalwart does not close or re-validate such a session at token expiry, the real bound for a disabled person is not 30 minutes but "until the client disconnects". The same question applies to an app-password session after someone revokes the app password.

So I would add a step 2b: authenticate one IMAP session before the disable, keep it open with NOOP every minute, and record whether and when the server ends it. I do not know what Stalwart does here, and the test should be written so that either answer is recorded rather than one being assumed.

A smaller point on "fails within one attempt after exp": validators usually allow some clock skew. If the first failure comes a minute or two late, that is the leeway, not a defect, and the script should report the measured delay rather than fail on it.

Philippe's point about the browser's local copy belongs in the same README paragraph: the test proves the server stops answering, not that the person loses the mail already on their laptop.

Chris

---

<a id="msg-14"></a>
### 14. Vera → Philippe · 11:49

*Subject:* Re: Hi again · *Cc:* Chris, Alice, Carol, Bob

Philippe,

The lab does not block outbound traffic. The Docker network has internal=false, and CoreDNS forwards every name outside lab.internal to 1.1.1.1 and 9.9.9.9. CoreDNS logs every query, so its log shows which container looked up which outside name. I counted from that log today.

Stalwart (172.28.0.30):
- github.com and release-assets.githubusercontent.com, 44 lookups each. From memory, not checked, Stalwart downloads its web admin bundle and its spam filter rules from GitHub releases. This means the server pulls code and rules from a third party while it runs, separately from the image version problem.
- notifications.relay.bulwarkmail.org, 14 lookups. Stalwart is the one calling the push relay, not the webmail. When a browser registers a JMAP push subscription with that URL, Stalwart sends state-change notices to it. So push is active here. RFC 8620 defines those notices as account and state identifiers, with no mail content. I have not checked whether Stalwart encrypts them. Either way, the relay operator can see that a given account changed and when.
- gmail.com, gmail-smtp-in.l.google.com and the gmail.com MTA-STS records, plus example.com. Someone sent mail to those domains. The mail went to Mailpit, but Stalwart still looked up the recipient domain's MX and MTA-STS records in public DNS. In production that is expected. In the lab it means the public resolvers saw which domains we wrote to.

Bulwark (172.28.0.41):
- version.telemetry.bulwarkmail.org, 4 lookups. This matches what you found.
- icons.duckduckgo.com, 2 lookups. I assume these are sender favicons, which would tell DuckDuckGo which domains our users receive mail from. I have not read the code, so I have not confirmed it.
- telemetry.bulwarkmail.org was never looked up, which matches consent being off.

A DNS log shows lookups. It does not show connections or what was sent. It is the minimum, not the full list.

For a production egress rule I would allow:
- SMTP out on 25, plus DNS, for Stalwart only.
- HTTPS to Keycloak, internally.
- Nothing else, by default.

Each of the four outside hosts above (GitHub, the Bulwark push relay, Bulwark's version check, DuckDuckGo) then needs an owner and a decision. Either it goes on Carol's processor register, or it is switched off and the feature with it. GitHub I would replace with a pinned, reviewed copy of the rules that we serve ourselves. Push without the relay means running our own Web Push endpoint or giving up push notifications.

On your local-store point: a disable ends access to the server. It does not delete what a browser has already cached. That gap is on the device side, not on the server, and the offboarding procedure should say so.

Vera

---

<a id="msg-15"></a>
### 15. Philippe → Chris · 11:49

*Subject:* Re: Hi again · *Cc:* Vera, Alice, Carol, Bob

Chris, Vera,

Step 2b has a webmail twin, and it should go in the same test. Bulwark subscribes to JMAP push through the session's eventSourceUrl (an EventSource, in its client chunk; I found no WebSocket transport). An EventSource is one long-lived HTTP response, authenticated once when it opens, exactly like an IMAP session in IDLE. So the same question applies: does Stalwart end that stream when the bearer token expires, or does a disabled person keep receiving change notifications until the tab is closed? Bulwark would still need a valid token to fetch the messages themselves, so the exposure is smaller than the IMAP case, but it is a second open channel and the bound should be measured, not assumed.

It can be tested without a browser: open the eventSourceUrl with curl and the test user's token before the disable, keep it open, and record when the server closes it. Same record format as 2b.

Agreed on adding the local-copy sentence to the README paragraph.

Philippe

---

<a id="msg-16"></a>
### 16. Vera → Chris · 11:49

*Subject:* Re: Hi again · *Cc:* Alice, Philippe, Carol, Bob

Chris,

I agree with 2b, and the test should record whichever result it gets. I also do not know what Stalwart does with a session that is already authenticated. I have two additions.

1. The webmail has the same problem. Caddy's log shows Bulwark keeping a JMAP EventSource stream open, /jmap/eventsource/?closeafter=no&ping=30. It authenticates once, when the stream opens, and then stays open. So 2b should keep three sessions open across the disable:
   - IMAP, authenticated with a Keycloak token
   - IMAP, authenticated with an app password, then revoke the app password
   - a JMAP EventSource stream

   For each one, send something into the mailbox after the disable and record whether the change still reaches the open session. That shows whether the open session still delivers mail to a disabled user, which is the question Carol is asking.

2. On clock skew, report the measured delay, as you say, but also set a ceiling the script fails on. I suggest 5 minutes past exp. A test that only reports would pass even if tokens stopped expiring at all, and that failure is the one worth catching.

I agree that the README paragraph should cover the local copy too. The test proves the server stops serving the person. It does not prove the mail is gone from their laptop.

Vera

---

<a id="msg-17"></a>
### 17. Chris → Philippe · 11:49

*Subject:* Re: Hi again · *Cc:* Vera, Alice, Carol, Bob

Philippe, Vera,

Agreed, 2c, same record format: opened at, token exp, closed at (or "still open at end of test").

That gives three long-lived channels, and Vera's DNS log suggests a fourth that is not tied to any connection at all. A Web Push subscription is stored server-side and survives the browser closing. If a disabled person had push registered, Stalwart may keep sending state-change notices for their account to the relay after the disable, with no token involved anywhere. Whether a disable in Keycloak removes or suspends that subscription in Stalwart is a question for the same test: register a push subscription for the throwaway user before the disable, point it at a local endpoint instead of the public relay, and record whether notices still arrive after exp.

That is the last case I can see. At this point the script has enough steps that it should be written and reviewed one step at a time, starting with the setup and the EXIT trap, before any assertion. Whoever writes it, I will review it.

Chris

---

<a id="msg-18"></a>
### 18. Vera → Philippe · 11:49

*Subject:* Re: Hi again · *Cc:* Chris, Alice, Carol, Bob

Philippe,

Our mails crossed. The EventSource is now one of the three sessions in 2b, with the record format you describe.

There is a third channel that needs no open connection at all, the JMAP push subscription. Stalwart sends to notifications.relay.bulwarkmail.org (14 lookups in the DNS log), so a subscription is registered here. A PushSubscription is stored on the server and carries its own expiry. Nothing in the protocol ties it to the token that created it. If Stalwart does not delete a disabled account's push subscriptions, it keeps sending state-change notices for that account to the relay until the expiry. I have not checked what expiry Stalwart sets.

For the test, call PushSubscription/get for the test user before the disable and record the expires value. After the disable, deliver a message and check whether the relay is still contacted. The CoreDNS log is enough to show whether Stalwart looks the relay up again. For the README, the offboarding procedure has to delete push subscriptions along with app passwords.

Vera

---

<a id="msg-19"></a>
### 19. Chris → Vera · 11:49

*Subject:* Re: Hi again · *Cc:* Alice, Philippe, Carol, Bob

Vera,

You are right about the ceiling, and my version was wrong. A test that only reports cannot fail on the one regression that matters, tokens that never expire. Five minutes past exp, reported delay below that, failure above it.

Agreed on delivering a message into the mailbox after the disable for each open session. It turns "is the connection still open" into "does a disabled person still receive mail", which is the question as Carol stated it.

Our mails crossed. With Philippe's EventSource and my push-subscription case, the list is: three open sessions (IMAP with token, IMAP with app password then revoked, EventSource), plus the stored push subscription pointed at a local endpoint. Each records opened, exp, last change received, closed. That is the complete specification as far as I can see, and I have nothing further to add until there is a first step of the script to review.

Chris

---

<a id="msg-20"></a>
### 20. Philippe → Vera · 11:49

*Subject:* Re: Hi again · *Cc:* Chris, Alice, Carol, Bob

Vera,

Confirmed on DuckDuckGo, and there is a better source than either of our logs. Bulwark keeps its own register of every outside host it can contact, with the switch for each, served by its admin API (/api/admin/outgoing-hosts). I read it from the route file in the running container.

Favicons. The browser asks our server (/api/favicon), and the server fetches https://icons.duckduckgo.com/ip3/<sender domain>.ico. So DuckDuckGo sees our server's address and every sender domain someone displays, not the user's address. The register names only one off switch, a per-user setting (Settings > Message list). I found no administrator-wide switch for it. The same setting also controls BIMI logo lookups, which go to each sender's own domain. Per-user is the wrong level for something that belongs in a processor register; unless an administrator can lock it, an egress block is the only reliable switch, and the icons then fail to load. I have not checked what the message list shows when they do, or whether that layout holds without them.

Other entries in that register that apply to us:
- Version check: BULWARK_UPDATE_CHECK=off.
- Push relay: our own relay can be configured and locked under Users > Features and settings > Push relays.
- Translation: the Translate plugin defaults to api.mymemory.translated.net, which receives the text being translated, so message content. It is a plugin and I do not know if it is installed here. If it is, that one matters more than the rest; LIBRETRANSLATE_URL points it at our own instance.
- Extension directory (extensions.bulwarkmail.org) and connector (connector.bulwarkmail.org): EXTENSION_DIRECTORY_URL and CONNECTOR_ENABLED=false.

Carol: that register is a vendor statement, not an audit, but it gives the webmail half of your list with a switch per line. Vera's DNS log is the check on it.

Philippe

---

<a id="msg-21"></a>
### 21. Bob → Philippe · 11:55

*Subject:* Re: Hi again · *Cc:* Vera, Chris, Alice, Carol

All,

I stepped away for one coffee and came back to fifteen messages and a test specification. Thank you. That is more review than this lab has had in its whole life (about four hours).

It's my lab, so here are the decisions, one per item.

1. scripts/55-offboarding.sh. Accepted as specified: a throwaway user created in Keycloak and Stalwart and removed again by an EXIT trap; step 1 against Keycloak's token endpoint; step 2 measured against exp, failing beyond 5 minutes; three sessions held open across the disable (IMAP with token, IMAP with app password then revoked, EventSource), each sent a new message after the disable; and a push subscription pointed at a local endpoint. I write it. Chris reviews one step at a time, starting with the setup and the trap. Vera reviews the timing. Philippe reviews 2c.

2. Pinned versions. Stalwart, Bulwark, Mailpit and the CLI go to digests this afternoon. "v1" and "latest" leave the compose file. Upgrades become a commit with a changelog link, not a restart.

3. Outbound traffic. Default deny for production, as Vera proposes. In the lab I will turn off what Philippe found has a switch: BULWARK_UPDATE_CHECK=off, CONNECTOR_ENABLED=false, and an empty extension directory. Push either goes through our own relay or is switched off. Favicons will stay off until someone finds an administrator-wide switch; per-user is not a control. I don't know yet whether the Translate plugin is installed. I will check before I claim anything, and if it is installed it goes first.

4. The gmail.com and example.com lookups were me, testing delivery to Mailpit this morning. In the lab the public resolvers learned nothing worse than that I write to example.com.

5. README. The "Account management" paragraph gets rewritten when 55 exists, not before: two-step offboarding (Keycloak and Stalwart app passwords plus push subscriptions), the measured bounds from the test, and the fact that a browser's local copy outlives all of it.

Carol, for your list: Vera's and Philippe's licence findings agree once you use Bulwark's package.json rather than the image label. That makes two AGPL components, both unmodified, and both will stay that way. Theming goes through configuration and fixes go upstream. The Enterprise quote is still open, and I will ask Stalwart Labs for it in writing. On continuity: Vera, would you be my second on this? Nobody else in the thread has a terminal they trust at 03:00.

Bob

---

<a id="msg-22"></a>
### 22. Philippe → Bob · 11:55

*Subject:* Re: Hi again · *Cc:* Vera, Chris, Alice, Carol

Bob,

I take 2c. What I will check in review: the stream is opened with the throwaway user's token before the disable, not after; "closed" means the server ended the response, distinguished from curl timing out on our side; and a message delivered after the disable is recorded as received or not received on the stream, with the time.

One item from my first mail is missing from your list: accessibility. The webmail is the only part staff will touch all day, and nobody has tested it. I will do one keyboard-only pass and one VoiceOver pass through login via Keycloak, read, reply, attach, search, plus 320px width and 200% zoom, and write up any WCAG 2.1 AA failures by success criterion. If there are serious ones, they belong in the business case next to the Enterprise quote. I will also check what the message list looks like with favicons off, since that is now the lab default.

I need a throwaway account for that, not alice. Tell me which one to use when it exists.

Philippe

---

<a id="msg-23"></a>
### 23. Chris → Bob · 11:56

*Subject:* Re: Hi again · *Cc:* Philippe, Vera, Alice, Carol

Bob,

Accepted. So that the first review is short, here is what I will check in the setup and the trap, before any assertion exists.

1. The trap removes only what was actually created. Setup can fail after the Keycloak user exists and before the Stalwart account does; the trap must not fail on the missing half, and it must not stop at the first error, because a half-cleaned run is the state we are trying to prevent. Record each created object in a variable as it is created, and let the trap act on those.

2. The order of cleanup. App passwords and push subscriptions first, then the Stalwart account, then the Keycloak user. If cleanup is interrupted, what is left behind should be the least capable thing.

3. The trap fires on Ctrl-C as well as on a failed command. A 30-minute test will be interrupted by hand at least once; check that EXIT is reached after INT under set -e on the bash version the lab uses, and test it by interrupting a run, not by reading the code.

4. The throwaway user's name is unique per run, so a run that crashed before its trap does not collide with the next one, and a leftover is recognisable by name.

5. No secret reaches the terminal or a log: the app password and tokens go into variables, not into echo.

Send me the setup and trap on their own, with one run that fails on purpose halfway through setup. If that run leaves nothing behind in either Keycloak or Stalwart, the next step is the first assertion.

Philippe's throwaway account for accessibility can come from the same setup function, kept but not disabled.

Chris

---

<a id="msg-24"></a>
### 24. Vera → Bob · 11:57

*Subject:* Re: Hi again · *Cc:* Philippe, Chris, Alice, Carol

Bob,

Yes, with three conditions. Without them, my name in the continuity plan is just a second person nobody has prepared, which is what Carol warned about.

1. The time is written down and approved. It needs a number of hours per month on a budget line, signed by whoever owns the budget, as Carol asked. Without it the work gets done in spare time, and spare time is the first thing to go.

2. I run the procedures before I am named. The repository has no backup or restore procedure today. I searched it for backup, restore, snapshot and pg_dump and found nothing. All state lives in five Docker volumes:
   - stalwart-data and stalwart-etc
   - postgres-data, which holds Keycloak's database
   - bulwark-data
   - caddy-data

   Before anyone writes my name in a continuity plan, I want to have done each of these myself, from a runbook you have written and not from memory:
   - restored all five volumes onto an empty host
   - replaced an expired certificate
   - recovered from a failed Stalwart upgrade

   Any step I have to ask you about goes into the runbook.

3. I have my own access, with my own credentials. I do not use yours, and nothing that only you can unlock. For the lab that means my own Keycloak admin account and my own Stalwart administrator. In production it also means the alerting goes to both of us.

I accept the timing review for 55. I will check three things:
- That exp is read from the token itself, not computed from the realm setting.
- That the time of the disable is recorded from Keycloak's response, not from when the script sent the request.
- That the 5-minute ceiling is measured on one clock. Keycloak and the lab containers run on the same host, so any clock difference should be close to zero. If it is not, the test should report the difference rather than absorb it.

Carol, the condition you set is not met until all three of these are done. Agreeing to it in an e-mail does not meet it.

Vera

---

<a id="msg-25"></a>
### 25. Carol → Vera · 11:57

*Subject:* Re: Hi again · *Cc:* Bob, Philippe, Chris, Alice

Vera, Bob,

I approve all three conditions as written. For the record, here is what that means on my side.

1. Time on a budget line. Approved in principle. The budget line is mine to sign, so I need a number. Bob, give me an estimate of the hours per month for Vera's share. It should cover running the procedures, upgrades and her review of 55, and keep the one-off runbook work separate from the recurring time. I will sign it once I have it. Until then nobody counts Vera's hours as available, and that includes me in the business case.

2. Procedures run before naming. Approved, and I will add a condition of my own. The continuity plan will name Vera only after she has sent me the date of each of the three exercises (the restore of all five volumes, the certificate replacement and the failed-upgrade recovery) and the runbook version she used. An e-mail saying "done" is enough evidence for me. A runbook that exists only in Bob's head is not.

3. Separate access. Approved. In production I will also want the list of people with administrator access to Keycloak and Stalwart reviewed twice a year, and the first review is the moment Vera's account is created.

Vera is right that agreeing by e-mail does not meet the condition. So the status for the business case is "conditionally staffed", with these three items listed as open. Nothing changes that to "staffed" until they are closed.

Philippe, your accessibility pass goes in the business case as its own line next to the Enterprise quote. If the free webmail needs paid remediation to reach WCAG 2.1 AA, that is part of its cost.

Carol

---

<a id="msg-26"></a>
### 26. Philippe → Carol · 11:58

*Subject:* Re: Hi again · *Cc:* Vera, Bob, Chris, Alice

Carol,

Agreed. One thing that line should state so the number is not read as a one-off: remediation sent upstream is accepted on the maintainers' schedule, or not at all. For each failure I find, the line will list one of three outcomes. Fixed upstream in a released version. Waiting upstream, with a date we can live with. Or refused, in which case the choice is to live with the failure or to fork, and a fork brings back both the AGPL source obligation and maintenance we would carry with every upgrade. The third kind is the expensive one, and I cannot price it until I have the test results.

The pass itself comes first. Until I have run it, I do not know whether there is anything to remediate.

Philippe

---

<a id="msg-27"></a>
### 27. Bob → Carol · 12:03

*Subject:* Re: Hi again · *Cc:* Vera, Philippe, Chris, Alice

Carol,

Here is the estimate you asked for. It splits one-off work from recurring work, and Vera's hours from mine. These are estimates, not measurements. I explain at the end how we turn them into measurements.

**Vera, one-off (about 24 hours, spread over the next two months)**

| Item | Hours |
|---|---:|
| Reading the stack and the runbook, setting up her own access | 5 h |
| Restoring all five volumes onto an empty host | 8 h |
| Replacing an expired certificate | 2 h |
| Recovering from a failed Stalwart upgrade | 6 h |
| Timing review of 55-offboarding.sh | 3 h |


**Vera, recurring (about 8 hours per month)**

| Item | Hours |
|---|---:|
| Reviewing and rehearsing the monthly upgrade (now that versions are pinned, every upgrade is a deliberate change) | 3 h |
| Quarterly restore drill, averaged per month | 2 h |
| Incidents and questions as second on call, averaged per month | 2 h |
| Runbook upkeep and the twice-yearly access review | 1 h |


**Mine, one-off (about 32 hours, so you can see the whole cost)**

| Item | Hours |
|---|---:|
| Writing the runbook Vera works from | 16 h |
| 55-offboarding.sh, written and reviewed one step at a time | 12 h |
| Pinning versions and switching off outbound traffic in the lab | 4 h |


Two things are not included. Being reachable after hours is on-call duty, and how that is paid is decided by the collective agreement, not by this table. It is your call, and it is the larger cost. Also not included: any accessibility remediation Philippe's pass turns up, and the Enterprise licence.

How to check the numbers: Vera and I log actual hours against this budget line for the first three months. If the recurring figure is more than 25 percent off in either direction, I send you a corrected estimate instead of letting it drift. The restore exercise is the least certain item. If it takes much longer than 8 hours, that tells us the runbook is wrong, which is exactly what the exercise is for.

Proposed budget line: 24 hours once and 8 hours per month for Vera, reviewed after three months.

Bob

---

<a id="msg-28"></a>
### 28. Carol → Bob · 12:03

*Subject:* Re: Hi again · *Cc:* Vera, Philippe, Chris, Alice

Bob, Vera,

Budget approved: 24 hours once and 8 hours per month for Vera, charged to the mail platform line, starting this month. I will review it after three months against the hours you log.

The terms I attach:

1. Separate codes. Log the one-off hours and the recurring hours under separate codes. If they are mixed, the three-month review compares nothing.

2. Overrun. The one-off 24 hours is a ceiling, not a target. If the restore exercise runs over, tell me when it happens, not at the review. As you say, an overrun means the runbook is wrong, and I would rather pay to fix the runbook than pay for the same overrun every quarter.

3. Bob's 32 hours. I have recorded them as a cost of the business case. They come from your existing allocation, not from a new budget line. If that allocation cannot absorb them, tell me now.

4. On call. Noted that it is outside this line. I will take it to HR under the collective agreement and come back with the rate before anyone is put on a rota. Until then nobody is on call for this service, and the continuity plan says so.

That closes condition 1 of Vera's three. Conditions 2 and 3 remain open, so the status in the business case stays "conditionally staffed", now with one item closed and two open.

Carol

---

<a id="msg-29"></a>
### 29. Vera → Bob · 12:03

*Subject:* Re: Hi again · *Cc:* Carol, Philippe, Chris, Alice

Bob, Carol,

I accept the estimate for my share. It is missing one line, and that line comes before my restore exercise.

The restore assumes there is a backup to restore from, and there is none. The repository has no backup procedure, and nothing in it copies the five volumes anywhere. That needs its own line, on Bob's side. It has to cover:
- how the volumes are backed up
- where the copies are kept, which has to be outside the host they back up
- how long they are kept
- who is alerted when a backup fails

Stalwart and PostgreSQL each need a consistent copy. Copying the files of a running database is not a backup. I have not estimated this line, so I will not put a number on it.

It also bears on Carol's deletion requirement. Mail deleted on request still exists in every backup until that backup expires. How long backups are kept therefore sets the real deletion bound, and that should be decided together with the backup procedure, not added afterwards.

Vera

---

<a id="msg-30"></a>
### 30. Bob → Vera · 12:04

*Subject:* Re: Hi again · *Cc:* Carol, Philippe, Chris, Alice

Vera, Carol,

Correct, the lab has no backups. That is deliberate: it is a lab, and everything in it can be rebuilt from the repository in about ten minutes, apart from some test mail nobody will miss.

Production will have backups before it takes a single real mailbox. I am making that a go-live gate, not a follow-up. Carol, please list it in the business case as a condition of production next to Vera's three, with the same rule: it counts as done when it has been demonstrated, not when it has been promised.

It will cover the four points Vera listed:
- Consistent copies. pg_dump for Keycloak's database, and Stalwart's own export or a copy taken while it is stopped, never the files of a running server. Caddy and Bulwark data are copied as files.
- Kept off the host, in a separate location with separate credentials, so that losing the host or its admin account does not also lose the backups.
- A retention period that Carol and I decide together, because, as Vera says, it is also the real bound on deletion. I will bring a proposal with the backup design rather than pick a number now.
- Alerting on a failed or missing backup, sent to both of us.

The procedure goes into the runbook, and Vera's restore exercise uses these backups, not a copy I make by hand the day before. That fixes the order: backups first, then the restore exercise, then her name in the continuity plan.

Carol, I have not estimated my hours for this yet, and I will not guess. You will get a number together with the design, before I start the work.

Bob

---

<a id="msg-31"></a>
### 31. Vera → Bob · 12:05

*Subject:* Re: Hi again · *Cc:* Carol, Philippe, Chris, Alice

Bob,

I agree with the design. Some state is not in any of the five volumes, and the backup has to cover it too. I checked the working directory. These files are excluded from git by .gitignore, so they exist only on the host:

- .env, which holds the administrator passwords for Keycloak, PostgreSQL and Stalwart.
- certs/, including rootCA-key.pem, the private key of the CA that every client has been told to trust.
- app-passwords.txt.
- dns/zones/lab.internal.zone. This one can be regenerated from Stalwart.

"Rebuilt from the repository in ten minutes" is therefore true only if we accept a new CA. In the lab, a new CA means repeating 00-prereqs.sh on every machine that trusts the old one. In production the certificate will presumably come from a different CA, but the passwords and the signing keys still live outside the volumes. They belong in a secret store with its own backup, not in a file that happens to be copied along with everything else. A pg_dump of the database is useless if the password to it existed only on the host we lost.

For the restore exercise, the empty host gets the backups and the secret store, and nothing else.

Vera

---

<a id="msg-32"></a>
### 32. Bob → Vera · 12:05

*Subject:* Re: Hi again · *Cc:* Carol, Philippe, Chris, Alice

Vera,

Agreed. My "ten minutes" assumed we would accept a new CA, and I did not say so. Here is how production handles each item you listed, and one item you did not list.

The item not on your list: the DKIM signing keys live in stalwart-data. So the volume backup itself contains secrets. That means backups are encrypted before they leave the host, and the encryption key is one more secret for the store. Without that key the backups are useless, and if it sits next to the backups it protects nothing.

Production, item by item:

- Administrator and database passwords (.env today). They go into the secret store the platform team already runs, which has its own backup and its own access control. We do not build a new one. The containers read the passwords from the store at start, and no .env file exists on the host.
- CA and certificates (certs/ today). Production uses certificates from a proper CA with automatic renewal. Nothing like rootCA-key.pem exists outside the lab. The lab CA stays a lab artefact, and the README will say that rotating it means re-running 00-prereqs.sh on every client.
- app-passwords.txt. It does not exist in production. App passwords belong to the user who created them, and Stalwart stores only their hash. The file is a lab convenience so that I can impersonate the whole cast, and that is exactly what it should never be anywhere else.
- Zone file. Generated from Stalwart, and the runbook says so. It is not backed up.

One recovery key unlocks the secret store and the backups. It is kept offline, in two copies held by you and me. Either of us can restore without the other, and nobody else can do it alone. Carol, that pair is also the continuity answer to "only Bob can unlock it".

For the restore exercise: the empty host gets the backups and access to the secret store, nothing else. If you have to ask me for anything beyond that, it is a defect in the runbook, and the exercise has done its job.

Bob

---

<a id="msg-33"></a>
### 33. Vera → Bob · 12:06

*Subject:* Re: Hi again · *Cc:* Carol, Philippe, Chris, Alice

Bob,

Approved as written. That includes the encrypted backups, the platform team's secret store, and the recovery key kept offline in two copies.

There is one point I will check in the restore exercise. I will restore with my copy of the recovery key, not yours. That is the only way to show that my copy works and that I can restore without you. Afterwards, the twice-yearly access review should confirm that both copies still exist where the runbook says they are.

Vera

---

<a id="msg-34"></a>
### 34. Carol → Vera · 12:07

*Subject:* Re: Hi again · *Cc:* Bob, Philippe, Chris, Alice

All,

I accept the conclusions of this thread as the technical basis for the business case. Here they are in one place, so that nobody has to reconstruct them from thirty messages:

1. Licensing. Stalwart and Bulwark are AGPL and stay unmodified. Theming is done through configuration, and fixes are sent upstream. All versions are pinned before production. I will check Stalwart's licence terms against its LICENSE file myself, because so far they rest on memory.
2. Offboarding. A Keycloak disable alone does not lock anyone out. 55-offboarding.sh measures the gaps: tokens, app passwords, open sessions, push subscriptions. The offboarding procedure closes them. The README states that a browser's local copy outlives all of it.
3. Outbound traffic. Default deny in production. Every outside host either has an owner on the processor register or is switched off.
4. Backups and secrets. Production gets encrypted, off-host backups. A retention period that also bounds deletion, to be decided by Bob and me. Secrets in the platform team's secret store. One offline recovery key, in two copies held by Bob and Vera.
5. Continuity. Vera is Bob's second, with the budget approved. She is named in the plan once she has run the three exercises with her own key and her own access.
6. Accessibility. Philippe's pass is a line in the business case, with an outcome recorded for each failure.

Still open, and not for the technical team: the Enterprise quote from Stalwart Labs, the on-call rate from HR, and Bob's hours for the backup design. Production remains gated on backups demonstrated and on Vera's conditions 2 and 3.

I suggest the work starts in the next sprint. In the order the thread arrived at:
- Pin the versions and lock down outbound traffic in the lab (Bob).
- 55-offboarding.sh, starting with setup and trap (Bob writes it; Chris, Vera and Philippe review).
- The accessibility pass (Philippe), on a throwaway account from the same setup.
- The backup and secret-store design with an hours estimate (Bob, for my approval before work begins).
- The runbook, then Vera's exercises.

Alice, if you agree, please put these in the sprint planning. Unless someone objects, this thread is closed.

Carol

---

<a id="msg-35"></a>
### 35. Alice → Carol · 12:08

*Subject:* Re: Hi again

Carol, & others,

I will add the work to the plan for next sprint. Thanks for your time and effort so far, and good work everyone!

/Alice

