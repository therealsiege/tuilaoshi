import { describe, expect, it } from "vitest";
import { GET } from "@/app/api/sync/route";
import { PUT } from "@/app/api/sync/[device]/route";
import { isExportPath } from "@/lib/sync";
import { jsonRequest, tokenFor, setupTestEnv } from "./helpers";

setupTestEnv();

// What Store.Export writes (internal/wiki/backup.go).
const phoneExport = {
  "schedules.yaml": "vocabulary/hsk3.md#认真:\n  introduced: 2026-11-01\n  last_reviewed: 2026-11-26\n  review_interval: 3\n",
  "updated.yaml": "journey.yaml: 2026-11-26T08:00:00Z\n",
  "reviews.csv": "card_id,at,rating,interval_was,interval_now\nvocabulary/hsk3.md#认真,2026-11-26T08:00:00Z,2,1,3\n",
  "journey.yaml": "gold: 12\n",
  "saves/20261126-080000.yaml": "name: 小明\n",
};

const put = async (device: string, body: unknown, as = device) =>
  PUT(jsonRequest(`/api/sync/${device}`, "PUT", body, await tokenFor("user-1", as)), {
    params: Promise.resolve({ device }),
  });
const get = async (device: string, user = "user-1") => GET(jsonRequest("/api/sync", "GET", undefined, await tokenFor(user, device)));

describe("sync", () => {
  it("round-trips one device's export to the others", async () => {
    const res = await put("iphone", { files: phoneExport });
    expect(res.status).toBe(200);
    expect(await res.json()).toMatchObject({ device: "iphone", files: 5 });
    await put("mac", { files: { "schedules.yaml": "{}\n" } });

    const mac = await (await get("mac")).json();
    expect(mac.devices).toHaveLength(1);
    expect(mac.devices[0]).toMatchObject({ device: "iphone", files: phoneExport });
    expect(Date.parse(mac.devices[0].updated)).not.toBeNaN();

    const phone = await (await get("iphone")).json();
    expect(phone.devices.map((d: { device: string }) => d.device)).toEqual(["mac"]);
    expect((await (await get("iphone", "user-2")).json()).devices).toEqual([]);
  });

  it("replaces a device's earlier upload", async () => {
    await put("iphone", { files: phoneExport });
    await put("iphone", { files: { "journey.yaml": "gold: 13\n" } });
    const mac = await (await get("mac")).json();
    expect(mac.devices[0].files).toEqual({ "journey.yaml": "gold: 13\n" });
  });

  it("refuses a write to another device's name", async () => {
    const res = await put("iphone", { files: phoneExport }, "mac");
    expect(res.status).toBe(403);
    expect((await (await get("mac")).json()).devices).toEqual([]);
  });

  it("refuses requests without a token", async () => {
    const res = await PUT(jsonRequest("/api/sync/iphone", "PUT", { files: {} }), { params: Promise.resolve({ device: "iphone" }) });
    expect(res.status).toBe(401);
    expect((await GET(jsonRequest("/api/sync", "GET"))).status).toBe(401);
  });

  it.each([
    "../journey.yaml",
    "saves/../../etc.yaml",
    "/etc/passwd.yaml",
    "./journey.yaml",
    ".hidden.yaml",
    "saves/.hidden.yaml",
    "a\\b.yaml",
    "progress.db",
    "notes.md",
    "sub/reviews.csv",
    "",
    "a//b.yaml",
    "C:/x.yaml",
  ])("refuses the path %j", async (path) => {
    const res = await put("iphone", { files: { [path]: "x" } });
    expect(res.status).toBe(400);
    expect((await res.json()).error).toMatch(/path/);
  });

  it("accepts every path an export writes", () => {
    for (const p of Object.keys(phoneExport)) expect(isExportPath(p)).toBe(true);
    expect(isExportPath("gazette/20261126-080000-123.yaml")).toBe(true);
    expect(isExportPath("private/frequency.yaml")).toBe(false);
  });

  it("refuses bad bodies", async () => {
    expect((await put("iphone", { files: { "journey.yaml": 3 } })).status).toBe(400);
    expect((await put("iphone", { files: { "journey.yaml": "a\u0000b" } })).status).toBe(400);
    expect((await put("iphone", { files: [] })).status).toBe(400);
    expect((await put("iphone", { files: {}, extra: 1 })).status).toBe(400);
    expect((await put("..", { files: {} }, "iphone")).status).toBe(400);
  });

  it("refuses an oversize body", async () => {
    process.env.SYNC_MAX_BYTES = "1000";
    const res = await put("iphone", { files: { "journey.yaml": "x".repeat(2000) } });
    expect(res.status).toBe(413);
    expect((await (await get("mac")).json()).devices).toEqual([]);
  });

  it("refuses an oversize body sent without a length", async () => {
    process.env.SYNC_MAX_BYTES = "1000";
    const chunk = new TextEncoder().encode(JSON.stringify({ files: { "journey.yaml": "x".repeat(2000) } }));
    const body = new ReadableStream({ start: (c) => (c.enqueue(chunk), c.close()) });
    const req = new Request("http://localhost/api/sync/iphone", {
      method: "PUT",
      headers: { authorization: `Bearer ${await tokenFor("user-1", "iphone")}` },
      body,
      duplex: "half",
    } as RequestInit);
    expect((await PUT(req, { params: Promise.resolve({ device: "iphone" }) })).status).toBe(413);
  });
});
