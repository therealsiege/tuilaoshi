# App Review notes: 推老师 TuiLaoShi 1.0.0

Two parts: the notes to paste into App Store Connect (App Review Information,
Notes), and the sign-in problem Clint has to settle before submitting.

## Notes for the reviewer (paste into App Store Connect)

TuiLaoShi is a Mandarin study game set on the road of Journey to the West. No
account is needed: every feature except sync works offline, straight after
install.

What to try without signing in:
1. Camp opens first. It shows the cards due today, the streak and the next stage of the journey.
2. Study (复习): pick Review, tap Show answer, then rate the card (Again, Hard, Good, Easy). Recognition, Production (type the Chinese), Listening and Sentence modes are below it.
3. Read (读): open a reading and tap any word for its pinyin and meaning.
4. The road (上路): walk with the direction pad. A blocked step says so in a banner.
5. Camp papers: the letter, the gazette, the diary and My progress.
6. Settings (gear, top right): voices, sound cues, tone colours, study pace and battle difficulty. Shadowing (跟读) asks for the microphone and speech recognition only when you open it.

Signing in (optional): Settings, Account and sync, Sign in with Apple. Signing
in syncs study progress with the TuiLaoShi game on a Mac and lets you pair one
(Pair a Mac shows a code for the Mac's terminal). There are no purchases and
no other accounts.

[Pick one of the endings below, per Clint's decision.]

- Option A ending: "Any Apple Account can sign in."
- Option B ending: "The game is in a closed beta: sign-in is limited to an allow-list. Your review account, ACCOUNT, has been added." (Only possible if Apple gives the account in advance; see below.)

## The sign-in problem

The server admits only Apple accounts on its allow-list (`ALLOW_LIST` in
Vercel: Apple user ids, or verified emails). A reviewer's Apple Account is
not on it, so tapping Sign in with Apple ends in "this Apple ID (...) is not
on the allow list". Reviewers test every button; a sign-in that fails is a
likely rejection under Guideline 2.1 (App Completeness). And Sign in with
Apple can't use a demo username and password, so the usual "demo account" box
in App Store Connect does not help.

The allow-list exists to keep the Claude relay (paid per call) private. The
iPhone app in 1.0 never calls Claude.

### Options

**A. Open sign-in and sync to any Apple account; keep Claude allow-listed (proposed).**
`/api/auth/apple` issues a session to anyone; `/api/claude` checks the
allow-list itself (the Apple user id from the session, which is the `sub`).
Sync costs little, and `SYNC_MAX_BYTES` already caps each upload. A reviewer
can sign in, sync and make a pairing code. Changes: a few lines in web/, a
test, and an `ALLOW_LIST` check moved to the Claude route. Once anyone can make an
account, App Review also requires deleting it in the app (Guideline
5.1.1(v)); see "Account deletion" below.

**B. Put the reviewer on the allow-list.**
Apple does not say which Apple Account a reviewer will use, and the `sub` is
only known after the first attempt (the server logs it when it refuses). Clint
could watch the Vercel log during review and add the id, but the first try
fails, which is what gets apps rejected. Not reliable.

**C. A demo mode.**
A "Try sync without an account" path, or a server-side review account behind a
code given in the review notes. More work than A, and a second way in to keep
safe, for a feature the reviewer can also skip. Only worth it if the allow-list
must stay on sign-in itself.

**D. Ship 1.0 without sign-in on the iPhone.**
Hide the Account and sync section until the server is open. Simplest for
review, but the iPhone can't sync with the Mac in 1.0.

Proposal: A. If the closed beta has to stay closed, D for 1.0 and TestFlight
for the beta testers.

### Account deletion

With A, accounts can be made in the app, so the app must also offer to delete
one (Guideline 5.1.1(v)). The server has no delete route today. The fix: a
`DELETE /api/account` that removes the user's `device_exports`, `pair_codes`
and Claude counters, and a Delete account button in Settings, Account and sync.
The privacy policy promises deletion on request through the support page in
the meantime.
