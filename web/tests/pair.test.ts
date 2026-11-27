import { afterEach, describe, expect, it, vi } from "vitest";
import { POST as pair } from "@/app/api/pair/route";
import { POST as claim } from "@/app/api/pair/claim/route";
import { authenticate } from "@/lib/session";
import { jsonRequest, tokenFor, setupTestEnv } from "./helpers";

setupTestEnv();
afterEach(() => vi.useRealTimers());

async function newCode(device = "iphone"): Promise<string> {
  const res = await pair(jsonRequest("/api/pair", "POST", undefined, await tokenFor("user-1", device)));
  expect(res.status).toBe(200);
  const body = await res.json();
  expect(body.code).toMatch(/^\d{6}$/);
  expect(new Date(body.expiresAt).getTime() - Date.now()).toBeGreaterThan(9 * 60_000);
  return body.code;
}

const claimCode = (code: string, device = "mac", ip = "203.0.113.1") =>
  claim(jsonRequest("/api/pair/claim", "POST", { code, device }, undefined, { "x-forwarded-for": ip }));

describe("pairing", () => {
  it("needs a signed-in device to make a code", async () => {
    expect((await pair(jsonRequest("/api/pair", "POST"))).status).toBe(401);
    expect((await pair(jsonRequest("/api/pair", "POST", undefined, "garbage"))).status).toBe(401);
  });

  it("gives the claiming device a token for the same user", async () => {
    const res = await claimCode(await newCode());
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toMatchObject({ user: "user-1", device: "mac" });
    const req = new Request("http://localhost/", { headers: { authorization: `Bearer ${body.token}` } });
    expect(await authenticate(req)).toEqual({ user: "user-1", device: "mac" });
  });

  it("refuses a code after ten minutes", async () => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-11-27T12:00:00Z"));
    const code = await newCode();
    vi.setSystemTime(new Date("2026-11-27T12:10:01Z"));
    const res = await claimCode(code);
    expect(res.status).toBe(410);
    expect((await res.json()).error).toMatch(/expired/);
  });

  it("refuses a code used once already", async () => {
    const code = await newCode();
    expect((await claimCode(code)).status).toBe(200);
    const again = await claimCode(code, "mac2");
    expect(again.status).toBe(410);
    expect((await again.json()).error).toMatch(/already been used/);
  });

  it("refuses unknown codes, bad input and the code's own device name", async () => {
    expect((await claimCode("12345")).status).toBe(400);
    expect((await claimCode("123456", "bad/name")).status).toBe(400);
    const code = await newCode();
    expect((await claimCode(code === "000000" ? "000001" : "000000")).status).toBe(404);
    expect((await claimCode(code, "iphone")).status).toBe(409);
    expect((await claimCode(code, "mac")).status).toBe(200); // the 409 didn't use it up
  });

  it("limits claim attempts per address", async () => {
    for (let i = 0; i < 20; i++) expect((await claimCode("999999", "mac", "198.51.100.7")).status).toBe(404);
    expect((await claimCode("999999", "mac", "198.51.100.7")).status).toBe(429);
    expect((await claimCode("999999", "mac", "198.51.100.8")).status).toBe(404);
  });
});
