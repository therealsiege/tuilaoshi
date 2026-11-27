import { afterAll, describe, expect, it } from "vitest";
import { PgStore } from "@/lib/pg-store";
import { MemoryStore, type Store } from "@/lib/store";

// The same contract for both stores. PgStore runs only when
// TEST_DATABASE_URL names a throwaway database with db/*.sql applied.
const pgUrl = process.env.TEST_DATABASE_URL;
const pg = pgUrl ? PgStore.fromUrl(pgUrl) : undefined;
afterAll(async () => {
  await pg?.close();
});

const stores: [string, () => Store][] = [["memory", () => new MemoryStore()]];
if (pg) stores.push(["postgres", () => pg]);

describe.each(stores)("%s store", (_, make) => {
  const run = Math.random().toString(36).slice(2); // keeps Postgres runs apart
  const now = new Date("2026-11-27T12:00:00Z");
  const later = new Date("2026-11-27T12:11:00Z");

  it("pairs once, within the time, from another device", async () => {
    const s = make();
    const code = run.slice(0, 6);
    expect(await s.createPairing(code, `u-${run}`, "iphone", new Date(now.getTime() + 600_000), now)).toBe(true);
    expect(await s.createPairing(code, "someone-else", "x", new Date(now.getTime() + 600_000), now)).toBe(false);
    expect(await s.claimPairing(code, "iphone", now)).toEqual({ ok: false, reason: "same-device" });
    expect(await s.claimPairing(code, "mac", now)).toEqual({ ok: true, user: `u-${run}` });
    expect(await s.claimPairing(code, "mac", now)).toEqual({ ok: false, reason: "used" });
    expect(await s.claimPairing(`${code}x`, "mac", now)).toEqual({ ok: false, reason: "missing" });
    const code2 = `${run.slice(0, 5)}y`;
    await s.createPairing(code2, `u-${run}`, "iphone", new Date(now.getTime() + 600_000), now);
    expect(await s.claimPairing(code2, "mac", later)).toEqual({ ok: false, reason: "expired" });
    expect(await s.createPairing(code2, "next", "iphone", new Date(later.getTime() + 600_000), later)).toBe(true);
  });

  it("keeps the latest export per device", async () => {
    const s = make();
    const user = `u-${run}`;
    await s.putExport(user, "iphone", { "journey.yaml": "a" }, now);
    await s.putExport(user, "iphone", { "journey.yaml": "b", "saves/x.yaml": "认真" }, later);
    await s.putExport(user, "mac", {}, now);
    expect(await s.listExports(user)).toEqual([
      { device: "iphone", updated: later.toISOString(), files: { "journey.yaml": "b", "saves/x.yaml": "认真" } },
      { device: "mac", updated: now.toISOString(), files: {} },
    ]);
    expect(await s.listExports(`other-${run}`)).toEqual([]);
  });

  it("counts up to a cap", async () => {
    const s = make();
    expect(await s.bump(`k-${run}`, "2026-11-27", 2)).toBe(1);
    expect(await s.bump(`k-${run}`, "2026-11-27", 2)).toBe(2);
    expect(await s.bump(`k-${run}`, "2026-11-27", 2)).toBeNull();
    expect(await s.bump(`k-${run}`, "2026-11-28", 2)).toBe(1);
    expect(await s.bump(`z-${run}`, "2026-11-27", 0)).toBeNull();
  });
});
