import { randomInt } from "node:crypto";
import { getStore } from "@/lib/deps";
import { HttpError, route } from "@/lib/http";
import { authenticate } from "@/lib/session";

export const runtime = "nodejs";

const PAIR_MINUTES = 10;

// POST (Bearer) -> {code, expiresAt}: a 6-digit code another device can claim once.
export const POST = route(async (req) => {
  const { user, device } = await authenticate(req);
  const now = new Date();
  const expiresAt = new Date(now.getTime() + PAIR_MINUTES * 60_000);
  const store = getStore();
  for (let i = 0; i < 8; i++) {
    const code = String(randomInt(0, 1_000_000)).padStart(6, "0");
    if (await store.createPairing(code, user, device, expiresAt, now)) {
      return Response.json({ code, expiresAt: expiresAt.toISOString() });
    }
  }
  throw new HttpError(503, "could not make a pairing code; try again");
});
