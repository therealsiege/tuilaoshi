# TuiLaoShi privacy policy

Effective 28 November 2026.

TuiLaoShi (推老师) is a Mandarin study game for iPhone and Mac, made by Clint Johnson. This policy says what the game and its server (tuilaoshi.vercel.app) keep about you, and why.

## In short

- The game works without an account. Until you sign in, nothing leaves your device.
- If you sign in with Apple, the server keeps your Apple user id and your study progress, so your devices can share it.
- There are no ads, no analytics and no tracking. Nothing is sold or shared for marketing.

## What stays on your device

Your study progress (review schedules, answers, saves and settings) is kept in the app's own storage on your iPhone or Mac. Shadowing uses the microphone and Apple's speech recognition to hear a line you say back; the game sends no recordings to its server.

## What the server keeps when you sign in

- **Your Apple user id.** Sign in with Apple gives the app an identity token. The server checks it and keeps the anonymous user id Apple assigns to you for this app. It uses that id as your account key.
- **Your email address is not kept.** If the token carries a verified email, the server reads it once, to check whether your account is on the list of people allowed to sign in, and then drops it. While the game is in testing, only accounts on that list can sign in.
- **A device label**, "iphone" on every iPhone, and the Mac's computer name on a Mac, so your devices can tell their progress apart.
- **Your study progress**, as each device uploads it when you sync: review schedules, your answers with their times and ratings, and the game's other records (the journey, saves, settings). It is kept so your other devices can download it.
- **Pairing codes.** When you pair a Mac, the server keeps a six-digit code with your user id for ten minutes. It deletes the code a day after it expires.
- **Counts against abuse.** The server counts each account's Claude calls per day, to apply a daily limit, and counts pairing attempts per network address per hour, to stop codes being guessed.

## Claude

Some features of the game (on the Mac today) ask Claude, Anthropic's AI model, to write or grade text. The game sends those requests through its server so that its API key stays private. The server passes your prompt to Anthropic and streams the reply back to you. It does not store or log the prompt or the reply. Anthropic processes them under its commercial API terms. The iPhone app does not send anything to Claude yet.

## Who else handles your data

- **Apple**, for Sign in with Apple and speech recognition.
- **Vercel**, which runs the server and keeps short-lived request logs.
- **Neon**, which hosts the server's database.
- **Anthropic**, for Claude requests, as described above.

Nothing is shared with anyone else, and nothing is sold.

## How long it is kept

Your study progress and account id are kept until you ask for them to be deleted. Pairing codes are deleted a day after they expire. Signing out removes the session from your device; it does not delete what the server holds.

## Your choices

- You can use the whole game without signing in.
- You can sign out at any time in Settings.
- You can ask for everything the server holds about you to be deleted, or for a copy of it, through the support page below. You can also stop the app's access in your iPhone's Settings, under your name, Sign-In & Security, Sign in with Apple.

## Children

The game is not directed at children under 13 and does not knowingly collect their data.

## Changes

If this policy changes, the new version will be posted here with a new effective date.

## Contact

Questions and requests: https://github.com/therealsiege/tuilaoshi/issues
