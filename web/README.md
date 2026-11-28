# TuiLaoShi server

A Next.js app (App Router, route handlers under `app/api`) that signs people in, syncs progress between their devices, and forwards the game's Claude calls so the API key stays on the server.

## Routes

| Route | Request | Response |
|---|---|---|
| `POST /api/auth/apple` | `{identityToken, device}` | `{token, user, device, expiresAt}` |
| `POST /api/pair` (Bearer) | none | `{code, expiresAt}`: 6 digits, 10 minutes, one use |
| `POST /api/pair/claim` | `{code, device}` | `{token, user, device, expiresAt}` |
| `PUT /api/sync/:device` (Bearer) | `{files: {path: contents}}` | `{device, updated, files}` |
| `GET /api/sync` (Bearer) | none | `{devices: [{device, updated, files}]}` |
| `POST /api/claude` (Bearer) | `{system?, messages, max_tokens}` | `text/event-stream` |

Errors are `{error}` with the status code. Tokens last a year.

**Sync payload.** `files` is one device's export folder as written by `Store.Export` (internal/wiki/backup.go): `schedules.yaml`, `updated.yaml`, `reviews.csv` and each doc under its own name (`journey.yaml`, `saves/<id>.yaml`, ...). Paths are relative, `.yaml` or the root `reviews.csv`, with no `..` or hidden parts. To sync, a device writes each other device's `files` to `<root>/<device>/<path>` and runs `Store.Sync(root, name)`, which merges them and exports its own folder to `<root>/<name>`. It then uploads that folder with `PUT /api/sync/<name>`.

**Claude stream.** Each Anthropic stream event is passed through as `event: <type>` / `data: <json>`. A failure partway through ends the stream with an `error` event. The `X-Claude-Calls-Left` header counts what's left today. When the cap is reached the route answers 429 with `Retry-After`.

## Run

```bash
npm install
cp .env.example .env.local   # fill in values; leave DATABASE_URL empty for an in-memory store
npm run dev
```

## Test

```bash
npm test          # vitest: routes against the in-memory store, Apple keys and the Anthropic SDK mocked
npm run lint
npm run build
```

To run the store tests against Postgres as well, give them a throwaway database that `npm run migrate` has prepared: `TEST_DATABASE_URL=postgres://... npm test`.

## Deploy (Vercel + Neon)

1. Create a Vercel project with root directory `web/`, and add a Neon Postgres database. It sets `DATABASE_URL`; use the pooled URL.
2. Set the variables in `.env.example` in the project's environment settings. `SESSION_SECRET` should be 32+ random characters, for example from `openssl rand -base64 48`.
3. Migrations run on every Vercel build (`vercel-build`), so the database is ready when the server is; they are safe to repeat. To run them by hand: `DATABASE_URL=... npm run migrate`.
4. Deploy: `vercel deploy --prod` from `web/` (the project is 1putt/tuilaoshi).

Vercel caps request bodies at 4.5 MB, so `SYNC_MAX_BYTES` defaults to 4,000,000.
