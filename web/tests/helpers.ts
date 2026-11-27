import { afterEach, beforeEach } from "vitest";
import { setDepsForTests } from "@/lib/deps";
import { issueToken } from "@/lib/session";
import { MemoryStore } from "@/lib/store";

export const SECRET = "test-secret-that-is-at-least-32-characters-long";

// setupTestEnv gives each test a fresh in-memory store and known settings.
export function setupTestEnv(env: Record<string, string> = {}): { store: () => MemoryStore } {
  let store = new MemoryStore();
  const saved = { ...process.env };
  beforeEach(() => {
    store = new MemoryStore();
    setDepsForTests({ store });
    Object.assign(process.env, { SESSION_SECRET: SECRET, ...env });
  });
  afterEach(() => {
    process.env = { ...saved };
    setDepsForTests({});
  });
  return { store: () => store };
}

export async function tokenFor(user: string, device: string): Promise<string> {
  return (await issueToken({ user, device })).token;
}

export function jsonRequest(url: string, method: string, body?: unknown, token?: string, headers: Record<string, string> = {}): Request {
  return new Request(`http://localhost${url}`, {
    method,
    headers: {
      "content-type": "application/json",
      ...(token ? { authorization: `Bearer ${token}` } : {}),
      ...headers,
    },
    body: body === undefined ? undefined : typeof body === "string" ? body : JSON.stringify(body),
  });
}
