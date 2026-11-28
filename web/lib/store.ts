// Storage behind the routes: Postgres in production (lib/pg-store.ts),
// in memory for tests and local runs without DATABASE_URL.

export type ExportFiles = Record<string, string>;

export interface DeviceExport {
  device: string;
  updated: string; // RFC 3339
  files: ExportFiles;
}

export type ClaimResult =
  | { ok: true; user: string }
  | { ok: false; reason: "missing" | "expired" | "used" | "same-device" };

export interface Store {
  // createPairing saves a code unless a live one has it; false means retry.
  createPairing(code: string, user: string, device: string, expiresAt: Date, now: Date): Promise<boolean>;
  // claimPairing uses a code once, for a device other than the one that made it.
  claimPairing(code: string, device: string, now: Date): Promise<ClaimResult>;
  putExport(user: string, device: string, files: ExportFiles, updated: Date): Promise<void>;
  listExports(user: string): Promise<DeviceExport[]>;
  // bump adds one to a counter unless it has reached cap; it returns the
  // new count, or null when the cap was already reached.
  bump(key: string, bucket: string, cap: number): Promise<number | null>;
  // deleteAccount removes everything kept for user (exports, pairing codes,
  // Claude counters) and moves its token version on, so every token issued
  // before stops working. Deleting twice is harmless.
  deleteAccount(user: string): Promise<void>;
  // tokenVersion is the version a user's tokens must carry: 0 until the
  // account is first deleted.
  tokenVersion(user: string): Promise<number>;
}

// claudeKey names a user's Claude call counter (see app/api/claude).
export const claudeKey = (user: string) => `claude:${user}`;

interface Pairing {
  user: string;
  device: string;
  expiresAt: Date;
  used: boolean;
}

export class MemoryStore implements Store {
  private pairings = new Map<string, Pairing>();
  private exports = new Map<string, Map<string, DeviceExport>>();
  private counters = new Map<string, number>();
  private versions = new Map<string, number>();

  async createPairing(code: string, user: string, device: string, expiresAt: Date, now: Date) {
    const old = this.pairings.get(code);
    if (old && !old.used && old.expiresAt > now) return false;
    this.pairings.set(code, { user, device, expiresAt, used: false });
    return true;
  }

  async claimPairing(code: string, device: string, now: Date): Promise<ClaimResult> {
    const p = this.pairings.get(code);
    if (!p) return { ok: false, reason: "missing" };
    if (p.used) return { ok: false, reason: "used" };
    if (p.expiresAt <= now) return { ok: false, reason: "expired" };
    if (p.device === device) return { ok: false, reason: "same-device" };
    p.used = true;
    return { ok: true, user: p.user };
  }

  async putExport(user: string, device: string, files: ExportFiles, updated: Date) {
    let mine = this.exports.get(user);
    if (!mine) this.exports.set(user, (mine = new Map()));
    mine.set(device, { device, updated: updated.toISOString(), files: { ...files } });
  }

  async listExports(user: string) {
    return [...(this.exports.get(user)?.values() ?? [])]
      .map((e) => ({ ...e, files: { ...e.files } }))
      .sort((a, b) => a.device.localeCompare(b.device));
  }

  async bump(key: string, bucket: string, cap: number) {
    const k = `${key}\u0000${bucket}`;
    const n = this.counters.get(k) ?? 0;
    if (n >= cap) return null;
    this.counters.set(k, n + 1);
    return n + 1;
  }

  async deleteAccount(user: string) {
    this.exports.delete(user);
    for (const [code, p] of this.pairings) if (p.user === user) this.pairings.delete(code);
    const prefix = `${claudeKey(user)}\u0000`;
    for (const k of this.counters.keys()) if (k.startsWith(prefix)) this.counters.delete(k);
    this.versions.set(user, (this.versions.get(user) ?? 0) + 1);
  }

  async tokenVersion(user: string) {
    return this.versions.get(user) ?? 0;
  }
}
