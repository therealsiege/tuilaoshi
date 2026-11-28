import { jwtVerify, SignJWT } from "jose";
import { checkDevice, config } from "./config";
import { getStore } from "./deps";
import { HttpError } from "./http";

// Our own tokens: HS256, naming the user (Apple's sub), the device, and the
// user's token version (ver). Deleting the account moves the version on, so
// every token issued before it is refused; one from a later sign-in works.

const ISSUER = "tuilaoshi";
const AUDIENCE = "tuilaoshi-device";
export const TOKEN_DAYS = 365;

export interface Session {
  user: string;
  device: string;
}

export async function issueToken(s: Session, now = new Date()): Promise<{ token: string; expiresAt: string }> {
  const iat = Math.floor(now.getTime() / 1000);
  const exp = iat + TOKEN_DAYS * 86400;
  const ver = await getStore().tokenVersion(s.user);
  const token = await new SignJWT({ dev: s.device, ver })
    .setProtectedHeader({ alg: "HS256", typ: "JWT" })
    .setIssuer(ISSUER)
    .setAudience(AUDIENCE)
    .setSubject(s.user)
    .setIssuedAt(iat)
    .setExpirationTime(exp)
    .sign(config.sessionSecret());
  return { token, expiresAt: new Date(exp * 1000).toISOString() };
}

// authenticate reads the Bearer token and returns who and which device. A
// token whose version is behind the user's (the account was deleted) is
// refused like an expired one.
export async function authenticate(req: Request): Promise<Session> {
  const header = req.headers.get("authorization") ?? "";
  const m = /^Bearer\s+(\S+)$/i.exec(header);
  if (!m) throw new HttpError(401, "missing Bearer token");
  const secret = config.sessionSecret();
  let session: Session;
  let ver: unknown;
  try {
    const { payload } = await jwtVerify(m[1], secret, { issuer: ISSUER, audience: AUDIENCE, algorithms: ["HS256"] });
    if (typeof payload.sub !== "string" || !payload.sub) throw new Error("no sub");
    session = { user: payload.sub, device: checkDevice(payload.dev) };
    ver = payload.ver ?? 0; // tokens issued before versions carry none
  } catch {
    throw new HttpError(401, "invalid or expired token; sign in again");
  }
  if (ver !== (await getStore().tokenVersion(session.user))) {
    throw new HttpError(401, "this account was deleted; sign in again");
  }
  return session;
}
