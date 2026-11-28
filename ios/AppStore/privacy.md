# App Privacy answers: 推老师 TuiLaoShi 1.0.0

The answers for App Store Connect's App Privacy section, worked out from the
code: the iPhone app (ios/App), its Go engine (mobile/, internal/remote) and
the server (web/). They match ios/App/PrivacyInfo.xcprivacy. Clint confirms
them before submitting.

## What reaches the server, and when

Nothing leaves the iPhone until the learner signs in with Apple in Settings.
Without signing in the whole game works offline and nothing is collected.

After signing in, the iPhone calls three routes on https://tuilaoshi.vercel.app:

| Route | Sends | Kept |
|---|---|---|
| `POST /api/auth/apple` | Apple's identity token and the device label `iphone` | Nothing. The server checks the token, reads its `sub` (the Apple user id) and, if present, its verified email, and admits the account only if one of them is on the allow-list. The email is not stored or logged. A refused sign-in logs the Apple user id (not the email) in the Vercel function log. |
| `PUT /api/sync/iphone`, `GET /api/sync` | The study progress export | `device_exports`: the Apple user id, the device label, the export files, the time of the upload. Kept until deleted. |
| `POST /api/pair` | Only the session token | `pair_codes`: a 6-digit code, the Apple user id and the device label, for 10 minutes; the row is deleted a day after it expires. |

The study progress export holds: review schedules (`schedules.yaml`), every
answer with its time and rating (`reviews.csv`), and the game's other docs
(the journey, saves, settings, study bookkeeping). Never anything under the
wiki's `private/` folder; the server refuses those paths.

`POST /api/claude` relays prompts to Anthropic's API and streams the reply
back. It keeps neither and never logs them; it keeps only a count of calls per
Apple user id per day, for the daily cap. **The iPhone app does not call it in
1.0** (the diary and gazette say they wait for the server), so Claude prompts
and replies are not an iPhone data type yet. When the iPhone starts sending
them, add "Other User Content" sharing with Anthropic, below, and update the
manifest.

The Mac game's pairing claim (`POST /api/pair/claim`) counts attempts per IP
address per hour against guessing. The iPhone never calls it.

## Data types

| Apple's data type | Collected? | Linked to the user | Tracking | Purpose |
|---|---|---|---|---|
| Contact Info: Name | No | | | Sign in with Apple asks only for email, not name. |
| Contact Info: Email Address | No | | | Read from the identity token during sign-in to check the allow-list, then dropped. Apple counts data handled only to serve a request in real time, and not kept, as not collected. |
| Contact Info: everything else | No | | | |
| Health and Fitness | No | | | |
| Financial Info | No | | | |
| Location | No | | | |
| Sensitive Info | No | | | |
| Contacts | No | | | |
| User Content: Audio Data | No | | | Shadowing uses the microphone and Apple's speech recognition (on the device where it can). The app sends no audio to its server. Audio Apple's framework sends to Apple is not the developer's collection. |
| User Content: Other User Content | **Yes** | **Yes** | No | **App Functionality**: the study progress export, synced between the learner's devices. |
| User Content: emails, messages, photos, video, gameplay content, customer support | No | | | |
| Browsing History, Search History | No | | | |
| Identifiers: User ID | **Yes** | **Yes** | No | **App Functionality**: the Apple user id is the account key for sync and pairing. |
| Identifiers: Device ID | No | | | The device label is the fixed word `iphone`, the same on every iPhone; it tells the learner's own devices apart, and identifies nothing. |
| Purchases | No | | | |
| Usage Data: Product Interaction | No | | | Review answers are part of the progress export above, the learner's own study record, not analytics. Nothing else about use is recorded. |
| Usage Data: Advertising Data, Other Usage Data | No | | | |
| Diagnostics: Crash Data, Performance Data, Other | No | | | No crash reporter or analytics SDK. The app's crash log stays on the device. |
| Surroundings, Body | No | | | |
| Other Data | No | | | |

So in App Store Connect: "Yes, we collect data", then **User ID** and **Other
User Content**, each: used for App Functionality, linked to the user's
identity, not used for tracking.

## Tracking

None. No advertising, no analytics, no third-party SDK, no data broker. The
manifest says `NSPrivacyTracking` false with no tracking domains.

## Third parties that handle the data

- Vercel runs the server (and keeps its own request logs).
- Neon hosts the Postgres database.
- Anthropic receives Claude prompts relayed from the Mac game, under its
  commercial API terms. None from the iPhone in 1.0.
- Apple, for Sign in with Apple and speech recognition.
