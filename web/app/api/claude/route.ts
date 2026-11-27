import Anthropic from "@anthropic-ai/sdk";
import { checkClaudeRequest, sse, upstreamError, utcDay } from "@/lib/claude";
import { config } from "@/lib/config";
import { getStore } from "@/lib/deps";
import { jsonError, readJson, route } from "@/lib/http";
import { authenticate } from "@/lib/session";

export const runtime = "nodejs";
export const maxDuration = 300;

// POST (Bearer) {system?, messages, max_tokens} -> text/event-stream of
// Claude's own stream events (message_start ... message_stop), passed
// through as "event: <type>\ndata: <json>"; a failure mid-reply is an
// "error" event. Prompts and keys are never logged.
export const POST = route(async (req) => {
  const { user } = await authenticate(req);
  const params = checkClaudeRequest(await readJson(req, 2_000_000));
  const apiKey = config.anthropicKey();
  const cap = config.claudeDailyCap();
  const now = new Date();
  const used = await getStore().bump(`claude:${user}`, utcDay(now), cap);
  if (used === null) {
    const midnight = Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() + 1);
    return jsonError(
      429,
      `daily Claude limit reached: ${cap} calls a day; it resets at midnight UTC`,
      { "Retry-After": String(Math.ceil((midnight - now.getTime()) / 1000)) },
    );
  }

  const client = new Anthropic({ apiKey });
  let stream: AsyncIterable<Anthropic.RawMessageStreamEvent> & { controller: AbortController };
  try {
    stream = await client.messages.create(
      { model: config.claudeModel(), ...params, stream: true },
      { signal: req.signal },
    );
  } catch (err) {
    throw upstreamError(err);
  }

  let open = true; // false once the caller hangs up
  const body = new ReadableStream<Uint8Array>({
    async start(controller) {
      try {
        for await (const event of stream) {
          if (!open) break;
          controller.enqueue(sse(event.type, event));
        }
      } catch (err) {
        if (open) {
          const e = upstreamError(err);
          controller.enqueue(sse("error", { type: "error", error: { type: "api_error", message: e.message } }));
        }
      } finally {
        if (open) controller.close();
      }
    },
    cancel() {
      open = false;
      stream.controller.abort();
    },
  });
  return new Response(body, {
    headers: {
      "Content-Type": "text/event-stream; charset=utf-8",
      "Cache-Control": "no-cache, no-transform",
      "X-Accel-Buffering": "no",
      "X-Claude-Calls-Left": String(cap - used),
    },
  });
});
