import { exportJWK, generateKeyPair, createLocalJWKSet, SignJWT, jwtVerify, type CryptoKey } from "jose";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";
import { POST } from "@/app/api/auth/apple/route";
import { setDepsForTests } from "@/lib/deps";
import { authenticate } from "@/lib/session";
import { jsonRequest, SECRET, setupTestEnv } from "./helpers";

const BUNDLE = "com.therealsiege.tuilaoshi";
let applePrivate: CryptoKey;
let jwks: ReturnType<typeof createLocalJWKSet>;

beforeAll(async () => {
  const { publicKey, privateKey } = await generateKeyPair("RS256");
  applePrivate = privateKey;
  const jwk = { ...(await exportJWK(publicKey)), kid: "test-key", alg: "RS256", use: "sig" };
  jwks = createLocalJWKSet({ keys: [jwk] });
});

const env = setupTestEnv({ ALLOW_LIST: "001234.allowed, Friend@Example.com" });
beforeEach(() => setDepsForTests({ store: env.store(), appleKeys: jwks }));

async function appleToken(claims: Record<string, unknown>, opts: { aud?: string; exp?: string | number; key?: CryptoKey } = {}) {
  return new SignJWT(claims)
    .setProtectedHeader({ alg: "RS256", kid: "test-key" })
    .setIssuer("https://appleid.apple.com")
    .setAudience(opts.aud ?? BUNDLE)
    .setIssuedAt()
    .setExpirationTime(opts.exp ?? "10m")
    .sign(opts.key ?? applePrivate);
}

const signIn = (identityToken: unknown, device: unknown = "iphone") =>
  POST(jsonRequest("/api/auth/apple", "POST", { identityToken, device }));

describe("POST /api/auth/apple", () => {
  it("admits an allow-listed sub and returns our session token", async () => {
    const res = await signIn(await appleToken({ sub: "001234.allowed" }));
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toMatchObject({ user: "001234.allowed", device: "iphone" });
    const { payload, protectedHeader } = await jwtVerify(body.token, new TextEncoder().encode(SECRET));
    expect(protectedHeader.alg).toBe("HS256");
    expect(payload).toMatchObject({ sub: "001234.allowed", dev: "iphone" });
    expect((payload.exp as number) - (payload.iat as number)).toBeGreaterThan(300 * 86400);
    const req = new Request("http://localhost/", { headers: { authorization: `Bearer ${body.token}` } });
    expect(await authenticate(req)).toEqual({ user: "001234.allowed", device: "iphone" });
  });

  it("admits a verified allow-listed email, case-insensitively", async () => {
    const res = await signIn(await appleToken({ sub: "999.other", email: "friend@example.com", email_verified: "true" }));
    expect(res.status).toBe(200);
    expect((await res.json()).user).toBe("999.other");
  });

  it("refuses someone not on the allow list", async () => {
    const res = await signIn(await appleToken({ sub: "555.stranger", email: "stranger@example.com", email_verified: true }));
    expect(res.status).toBe(403);
    expect((await res.json()).error).toMatch(/555\.stranger.*allow list/);
  });

  it("refuses an unverified email even if listed", async () => {
    const res = await signIn(await appleToken({ sub: "999.other", email: "friend@example.com", email_verified: false }));
    expect(res.status).toBe(403);
  });

  it("refuses a token for another app", async () => {
    const res = await signIn(await appleToken({ sub: "001234.allowed" }, { aud: "com.example.other" }));
    expect(res.status).toBe(401);
    expect((await res.json()).error).toMatch(/audience/);
  });

  it("refuses an expired token", async () => {
    const res = await signIn(await appleToken({ sub: "001234.allowed" }, { exp: Math.floor(Date.now() / 1000) - 60 }));
    expect(res.status).toBe(401);
    expect((await res.json()).error).toMatch(/expired/);
  });

  it("refuses a token signed by another key", async () => {
    const { privateKey } = await generateKeyPair("RS256");
    const res = await signIn(await appleToken({ sub: "001234.allowed" }, { key: privateKey }));
    expect(res.status).toBe(401);
  });

  it("validates the body", async () => {
    expect((await signIn("")).status).toBe(400);
    expect((await signIn(await appleToken({ sub: "001234.allowed" }), "../mac")).status).toBe(400);
    expect((await POST(jsonRequest("/api/auth/apple", "POST", "not json"))).status).toBe(400);
  });
});
