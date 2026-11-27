import { errors, jwtVerify, type JWTPayload } from "jose";
import { config } from "./config";
import { getAppleKeys } from "./deps";
import { HttpError } from "./http";

export const APPLE_ISSUER = "https://appleid.apple.com";

// verifyAppleToken checks a Sign in with Apple identity token and returns
// its claims, or throws a 401 (or 502 when Apple's keys can't be fetched).
export async function verifyAppleToken(token: string): Promise<JWTPayload> {
  try {
    const { payload } = await jwtVerify(token, getAppleKeys(), {
      issuer: APPLE_ISSUER,
      audience: config.appleBundleId(),
      algorithms: ["RS256"],
    });
    if (typeof payload.sub !== "string" || !payload.sub) throw new HttpError(401, "Apple identity token has no subject");
    return payload;
  } catch (err) {
    if (err instanceof HttpError) throw err;
    if (err instanceof errors.JWTExpired) throw new HttpError(401, "Apple identity token has expired; sign in again");
    if (err instanceof errors.JWTClaimValidationFailed) {
      throw new HttpError(401, `Apple identity token has the wrong ${err.claim === "aud" ? "audience" : err.claim}`);
    }
    if (err instanceof errors.JWKSTimeout || err instanceof errors.JWKSInvalid) {
      throw new HttpError(502, "could not fetch Apple's signing keys; try again");
    }
    throw new HttpError(401, "invalid Apple identity token");
  }
}

// allowed admits a user whose Apple sub, or verified email, is listed.
export function allowed(payload: JWTPayload): boolean {
  const list = config.allowList();
  if (typeof payload.sub === "string" && list.includes(payload.sub.toLowerCase())) return true;
  const verified = payload.email_verified === true || payload.email_verified === "true";
  return verified && typeof payload.email === "string" && list.includes(payload.email.toLowerCase());
}
