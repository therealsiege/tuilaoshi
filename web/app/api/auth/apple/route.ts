import { allowed, verifyAppleToken } from "@/lib/apple";
import { checkDevice } from "@/lib/config";
import { HttpError, onlyKeys, readJson, route } from "@/lib/http";
import { issueToken } from "@/lib/session";

export const runtime = "nodejs";

// POST {identityToken, device} -> {token, user, device, expiresAt}
export const POST = route(async (req) => {
  const body = await readJson(req, 16_384);
  onlyKeys(body, ["identityToken", "device"]);
  const { identityToken } = body;
  if (typeof identityToken !== "string" || !identityToken || identityToken.length > 8192) {
    throw new HttpError(400, "identityToken must be Apple's identity token string");
  }
  const device = checkDevice(body.device);
  const claims = await verifyAppleToken(identityToken);
  if (!allowed(claims)) throw new HttpError(403, "this Apple ID is not on the allow list");
  const user = claims.sub as string;
  const { token, expiresAt } = await issueToken({ user, device });
  return Response.json({ token, user, device, expiresAt });
});
