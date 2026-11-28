import { describe, expect, it } from "vitest";
import { DELETE } from "@/app/api/account/route";
import { GET } from "@/app/api/sync/route";
import { claudeKey } from "@/lib/store";
import { jsonRequest, setupTestEnv, tokenFor } from "./helpers";

const env = setupTestEnv();

const del = (token?: string) => DELETE(jsonRequest("/api/account", "DELETE", undefined, token));
const sync = (token: string) => GET(jsonRequest("/api/sync", "GET", undefined, token));
const now = new Date("2026-11-28T12:00:00Z");

async function seed(user: string) {
  const s = env.store();
  await s.putExport(user, "iphone", { "journey.yaml": "gold: 1\n" }, now);
  await s.putExport(user, "mac", { "journey.yaml": "gold: 2\n" }, now);
  await s.createPairing(`${user}-code`, user, "iphone", new Date(now.getTime() + 600_000), now);
  await s.bump(claudeKey(user), "2026-11-28", 10);
}

describe("DELETE /api/account", () => {
  it("removes the user's exports, pairing codes and Claude usage", async () => {
    await seed("user-1");
    const res = await del(await tokenFor("user-1", "iphone"));
    expect(res.status).toBe(204);
    expect(await res.text()).toBe("");
    const s = env.store();
    expect(await s.listExports("user-1")).toEqual([]);
    expect(await s.claimPairing("user-1-code", "mac", now)).toEqual({ ok: false, reason: "missing" });
    expect(await s.bump(claudeKey("user-1"), "2026-11-28", 1)).toBe(1); // the count starts again
  });

  it("refuses every earlier token, from any device, with 401", async () => {
    const phone = await tokenFor("user-1", "iphone");
    const mac = await tokenFor("user-1", "mac");
    expect((await del(phone)).status).toBe(204);
    for (const t of [phone, mac]) {
      const res = await sync(t);
      expect(res.status).toBe(401);
      expect((await res.json()).error).toMatch(/deleted/);
    }
    // Signing in again afterwards makes a fresh, empty account.
    const fresh = await sync(await tokenFor("user-1", "iphone"));
    expect(fresh.status).toBe(200);
    expect((await fresh.json()).devices).toEqual([]);
  });

  it("leaves another user's data alone", async () => {
    await seed("user-1");
    await seed("user-2");
    const other = await tokenFor("user-2", "mac");
    expect((await del(await tokenFor("user-1", "iphone"))).status).toBe(204);
    const s = env.store();
    expect(await s.listExports("user-2")).toHaveLength(2);
    expect(await s.bump(claudeKey("user-2"), "2026-11-28", 10)).toBe(2);
    expect(await s.claimPairing("user-2-code", "mac", now)).toEqual({ ok: true, user: "user-2" });
    expect((await sync(other)).status).toBe(200);
  });

  it("is safe to call twice", async () => {
    await seed("user-1");
    const t = await tokenFor("user-1", "iphone");
    expect((await del(t)).status).toBe(204);
    expect((await del(t)).status).toBe(401); // the token went with the account
    expect((await del(await tokenFor("user-1", "iphone"))).status).toBe(204);
    expect(await env.store().listExports("user-1")).toEqual([]);
  });

  it("needs a token", async () => {
    expect((await del()).status).toBe(401);
  });
});
