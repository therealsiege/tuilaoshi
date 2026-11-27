import Anthropic from "@anthropic-ai/sdk";
import { HttpError, onlyKeys } from "./http";

// The request body /api/claude accepts, checked field by field.

export const MAX_TOKENS = 4096;
const MAX_MESSAGES = 200;
const MAX_CHARS = 400_000; // across system and every message

export interface ClaudeRequest {
  system?: string;
  messages: Anthropic.MessageParam[];
  max_tokens: number;
}

export function checkClaudeRequest(body: Record<string, unknown>): ClaudeRequest {
  onlyKeys(body, ["system", "messages", "max_tokens"]);
  let chars = 0;
  const { system, messages, max_tokens } = body;
  if (system !== undefined && typeof system !== "string") throw new HttpError(400, "system must be a string");
  chars += system?.length ?? 0;
  if (typeof max_tokens !== "number" || !Number.isInteger(max_tokens) || max_tokens < 1 || max_tokens > MAX_TOKENS) {
    throw new HttpError(400, `max_tokens must be a whole number from 1 to ${MAX_TOKENS}`);
  }
  if (!Array.isArray(messages) || messages.length === 0 || messages.length > MAX_MESSAGES) {
    throw new HttpError(400, `messages must be a list of 1 to ${MAX_MESSAGES} messages`);
  }
  const out: Anthropic.MessageParam[] = messages.map((m: unknown, i) => {
    if (typeof m !== "object" || m === null) throw new HttpError(400, `messages[${i}] must be an object`);
    const { role, content, ...rest } = m as Record<string, unknown>;
    if (Object.keys(rest).length) throw new HttpError(400, `messages[${i}] may hold only role and content`);
    if (role !== "user" && role !== "assistant") throw new HttpError(400, `messages[${i}].role must be "user" or "assistant"`);
    if (typeof content === "string") {
      if (!content) throw new HttpError(400, `messages[${i}].content is empty`);
      chars += content.length;
      return { role, content };
    }
    if (!Array.isArray(content) || content.length === 0) {
      throw new HttpError(400, `messages[${i}].content must be a string or a list of text blocks`);
    }
    const blocks: Anthropic.TextBlockParam[] = content.map((b: unknown, j) => {
      const block = b as Record<string, unknown> | null;
      if (!block || block.type !== "text" || typeof block.text !== "string" || Object.keys(block).length !== 2) {
        throw new HttpError(400, `messages[${i}].content[${j}] must be {type: "text", text}`);
      }
      chars += block.text.length;
      return { type: "text", text: block.text };
    });
    return { role, content: blocks };
  });
  if (out[0].role !== "user") throw new HttpError(400, "the first message must be from the user");
  if (chars > MAX_CHARS) throw new HttpError(413, `the prompt is longer than ${MAX_CHARS} characters`);
  return { ...(system ? { system } : {}), messages: out, max_tokens };
}

// upstreamError turns an SDK error, before any reply streamed, into ours.
export function upstreamError(err: unknown): HttpError {
  if (err instanceof Anthropic.BadRequestError) return new HttpError(400, `Claude refused the request: ${err.message}`);
  if (err instanceof Anthropic.AuthenticationError || err instanceof Anthropic.PermissionDeniedError) {
    return new HttpError(502, "the server's Claude key was refused");
  }
  if (err instanceof Anthropic.RateLimitError) return new HttpError(503, "Claude is busy right now; try again shortly");
  if (err instanceof Anthropic.APIConnectionError) return new HttpError(502, "could not reach Claude; try again");
  if (err instanceof Anthropic.APIError) return new HttpError(502, `Claude returned an error (${err.status ?? "no status"})`);
  return new HttpError(500, "internal error");
}

const enc = new TextEncoder();

export function sse(event: string, data: unknown): Uint8Array {
  return enc.encode(`event: ${event}\ndata: ${JSON.stringify(data)}\n\n`);
}

// A UTC day, the cap's bucket.
export const utcDay = (d: Date) => d.toISOString().slice(0, 10);
