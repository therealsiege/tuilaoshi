import { checkDevice } from "@/lib/config";
import { getStore } from "@/lib/deps";
import { HttpError, onlyKeys, readJson, route } from "@/lib/http";
import { issueToken } from "@/lib/session";

export const runtime = "nodejs";

// Failed or not, each address gets this many claims an hour, so the
// million codes can't be walked.
const CLAIMS_PER_HOUR = 20;

// POST {code, device} -> {token, user, device, expiresAt}
export const POST = route(async (req) => {
  const body = await readJson(req, 4096);
  onlyKeys(body, ["code", "device"]);
  if (typeof body.code !== "string" || !/^\d{6}$/.test(body.code)) throw new HttpError(400, "code must be 6 digits");
  const device = checkDevice(body.device);
  const store = getStore();
  const now = new Date();
  const ip = req.headers.get("x-forwarded-for")?.split(",")[0]?.trim() || "unknown";
  if ((await store.bump(`claim:${ip}`, now.toISOString().slice(0, 13), CLAIMS_PER_HOUR)) === null) {
    throw new HttpError(429, "too many pairing attempts; wait an hour and try again");
  }
  const r = await store.claimPairing(body.code, device, now);
  if (!r.ok) {
    switch (r.reason) {
      case "missing":
        throw new HttpError(404, "no such pairing code");
      case "expired":
        throw new HttpError(410, "this pairing code has expired; ask for a new one");
      case "used":
        throw new HttpError(410, "this pairing code has already been used");
      case "same-device":
        throw new HttpError(409, "choose a device name different from the device that made the code");
    }
  }
  const { token, expiresAt } = await issueToken({ user: r.user, device });
  return Response.json({ token, user: r.user, device, expiresAt });
});
