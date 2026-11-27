import { HttpError } from "./http";

// Settings come from the environment on each call, so tests can change them.

function intEnv(name: string, fallback: number): number {
  const raw = process.env[name]?.trim();
  if (!raw) return fallback;
  const n = Number(raw);
  if (!Number.isInteger(n) || n < 0) throw new HttpError(500, `server misconfigured: ${name} must be a whole number`);
  return n;
}

export const config = {
  sessionSecret(): Uint8Array {
    const s = process.env.SESSION_SECRET ?? "";
    if (s.length < 32) throw new HttpError(500, "server misconfigured: SESSION_SECRET must be at least 32 characters");
    return new TextEncoder().encode(s);
  },
  appleBundleId: () => process.env.APPLE_BUNDLE_ID?.trim() || "com.therealsiege.tuilaoshi",
  allowList: (): string[] =>
    (process.env.ALLOW_LIST ?? "")
      .split(",")
      .map((s) => s.trim().toLowerCase())
      .filter(Boolean),
  anthropicKey(): string {
    const k = process.env.ANTHROPIC_API_KEY?.trim();
    if (!k) throw new HttpError(500, "server misconfigured: ANTHROPIC_API_KEY is not set");
    return k;
  },
  claudeModel: () => process.env.CLAUDE_MODEL?.trim() || "claude-sonnet-5-5",
  claudeDailyCap: () => intEnv("CLAUDE_DAILY_CAP", 200),
  syncMaxBytes: () => intEnv("SYNC_MAX_BYTES", 4_000_000),
};

// Device names become folder names in the Go sync root (Store.Sync), so
// they are one plain path segment.
const DEVICE = /^[A-Za-z0-9_][A-Za-z0-9._-]{0,63}$/;

export function checkDevice(v: unknown, field = "device"): string {
  if (typeof v !== "string" || !DEVICE.test(v)) {
    throw new HttpError(400, `${field} must be 1-64 letters, digits, ".", "_" or "-", not starting with "." or "-"`);
  }
  return v;
}
