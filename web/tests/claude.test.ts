import { beforeEach, describe, expect, it, vi } from "vitest";
import { jsonRequest, tokenFor, setupTestEnv } from "./helpers";

// The SDK is mocked: messages.create returns whatever the test sets up.
const create = vi.fn();
vi.mock("@anthropic-ai/sdk", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@anthropic-ai/sdk")>();
  class FakeAnthropic extends actual.default {
    constructor(opts: ConstructorParameters<typeof actual.default>[0]) {
      super(opts);
      (this as unknown as { messages: unknown }).messages = { create };
    }
  }
  return { ...actual, default: FakeAnthropic };
});

const { POST } = await import("@/app/api/claude/route");
const { default: Anthropic } = await import("@anthropic-ai/sdk");

setupTestEnv({ ANTHROPIC_API_KEY: "sk-test-not-real", CLAUDE_DAILY_CAP: "2" });

const events = [
  { type: "message_start", message: { id: "msg_1", role: "assistant", content: [], model: "claude-sonnet-5-5" } },
  { type: "content_block_start", index: 0, content_block: { type: "text", text: "" } },
  { type: "content_block_delta", index: 0, delta: { type: "text_delta", text: "你好" } },
  { type: "content_block_delta", index: 0, delta: { type: "text_delta", text: "！" } },
  { type: "content_block_stop", index: 0 },
  { type: "message_delta", delta: { stop_reason: "end_turn" }, usage: { output_tokens: 3 } },
  { type: "message_stop" },
];

function fakeStream(items: unknown[], failAfter?: Error) {
  return {
    controller: new AbortController(),
    async *[Symbol.asyncIterator]() {
      for (const e of items) yield e;
      if (failAfter) throw failAfter;
    },
  };
}

beforeEach(() => {
  create.mockReset();
  create.mockImplementation(async () => fakeStream(events));
});

const good = { system: "You are a tutor.", messages: [{ role: "user", content: "你好" }], max_tokens: 1024 };
const call = async (body: unknown, user = "user-1") =>
  POST(jsonRequest("/api/claude", "POST", body, await tokenFor(user, "iphone")));

function parseSSE(text: string) {
  return text
    .trim()
    .split("\n\n")
    .map((chunk) => {
      const [ev, data] = chunk.split("\n");
      return { event: ev.replace(/^event: /, ""), data: JSON.parse(data.replace(/^data: /, "")) };
    });
}

describe("POST /api/claude", () => {
  it("streams Claude's events through as server-sent events", async () => {
    const res = await call(good);
    expect(res.status).toBe(200);
    expect(res.headers.get("content-type")).toMatch(/^text\/event-stream/);
    expect(res.headers.get("x-claude-calls-left")).toBe("1");
    const got = parseSSE(await res.text());
    expect(got.map((g) => g.event)).toEqual(events.map((e) => e.type));
    expect(got.map((g) => g.data)).toEqual(events);
    expect(create).toHaveBeenCalledTimes(1);
    expect(create.mock.calls[0][0]).toEqual({ model: "claude-sonnet-5-5", ...good, stream: true });
  });

  it("uses CLAUDE_MODEL when set", async () => {
    process.env.CLAUDE_MODEL = "claude-haiku-5-5";
    await (await call(good)).text();
    expect(create.mock.calls[0][0].model).toBe("claude-haiku-5-5");
  });

  it("enforces the daily cap per user", async () => {
    expect((await call(good)).status).toBe(200);
    expect((await call(good)).status).toBe(200);
    const res = await call(good);
    expect(res.status).toBe(429);
    expect((await res.json()).error).toMatch(/daily Claude limit reached: 2 calls/);
    expect(Number(res.headers.get("retry-after"))).toBeGreaterThan(0);
    expect(create).toHaveBeenCalledTimes(2);
    expect((await call(good, "user-2")).status).toBe(200);
  });

  it.each([
    ["max_tokens over 4096", { ...good, max_tokens: 4097 }],
    ["max_tokens missing", { messages: good.messages }],
    ["max_tokens not whole", { ...good, max_tokens: 10.5 }],
    ["no messages", { ...good, messages: [] }],
    ["a bad role", { ...good, messages: [{ role: "system", content: "x" }] }],
    ["empty content", { ...good, messages: [{ role: "user", content: "" }] }],
    ["a non-text block", { ...good, messages: [{ role: "user", content: [{ type: "image", source: {} }] }] }],
    ["an assistant first", { ...good, messages: [{ role: "assistant", content: "hi" }] }],
    ["a non-string system", { ...good, system: ["x"] }],
    ["an unknown field", { ...good, model: "claude-opus-5-5" }],
  ])("refuses %s without spending a call", async (_, body) => {
    const res = await call(body);
    expect(res.status).toBe(400);
    expect(typeof (await res.json()).error).toBe("string");
    expect(create).not.toHaveBeenCalled();
  });

  it("accepts text blocks and multi-turn history", async () => {
    const body = {
      max_tokens: 10,
      messages: [
        { role: "user", content: [{ type: "text", text: "a" }] },
        { role: "assistant", content: "b" },
        { role: "user", content: "c" },
      ],
    };
    expect((await call(body)).status).toBe(200);
    expect(create.mock.calls[0][0].messages).toEqual(body.messages);
    expect(create.mock.calls[0][0]).not.toHaveProperty("system");
  });

  it("needs a token and a configured key", async () => {
    expect((await POST(jsonRequest("/api/claude", "POST", good))).status).toBe(401);
    delete process.env.ANTHROPIC_API_KEY;
    const res = await call(good);
    expect(res.status).toBe(500);
    expect((await res.json()).error).toMatch(/ANTHROPIC_API_KEY/);
  });

  it("maps an upstream failure before the stream to JSON", async () => {
    create.mockRejectedValueOnce(new Anthropic.APIConnectionError({ message: "down" }));
    const res = await call(good);
    expect(res.status).toBe(502);
    expect((await res.json()).error).toMatch(/could not reach Claude/);
  });

  it("ends with an error event when the stream fails midway", async () => {
    create.mockImplementationOnce(async () => fakeStream(events.slice(0, 3), new Anthropic.APIConnectionError({ message: "x" })));
    const got = parseSSE(await (await call(good)).text());
    expect(got.at(-1)).toEqual({ event: "error", data: { type: "error", error: { type: "api_error", message: "could not reach Claude; try again" } } });
  });

  it("never logs the prompt or the key", async () => {
    const spies = (["log", "info", "warn", "error", "debug"] as const).map((m) => vi.spyOn(console, m).mockImplementation(() => {}));
    create.mockRejectedValueOnce(new Error("boom"));
    await call({ ...good, messages: [{ role: "user", content: "SECRET-PROMPT" }] });
    await (await call({ ...good, messages: [{ role: "user", content: "SECRET-PROMPT" }] })).text();
    const logged = JSON.stringify(spies.flatMap((s) => s.mock.calls));
    expect(logged).not.toContain("SECRET-PROMPT");
    expect(logged).not.toContain("sk-test-not-real");
    spies.forEach((s) => s.mockRestore());
  });
});
