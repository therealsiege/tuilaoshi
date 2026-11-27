import { Pool } from "pg";
import type { ClaimResult, DeviceExport, ExportFiles, Store } from "./store";

// PgStore keeps everything in Postgres (schema: db/001_init.sql).
export class PgStore implements Store {
  constructor(private pool: Pool) {}

  static fromUrl(url: string): PgStore {
    return new PgStore(new Pool({ connectionString: url, max: 3 }));
  }

  close(): Promise<void> {
    return this.pool.end();
  }

  async createPairing(code: string, user: string, device: string, expiresAt: Date, now: Date) {
    await this.pool.query(`DELETE FROM pair_codes WHERE expires_at < $1`, [new Date(now.getTime() - 86_400_000)]);
    const r = await this.pool.query(
      `INSERT INTO pair_codes (code, user_id, device, expires_at) VALUES ($1, $2, $3, $4)
       ON CONFLICT (code) DO UPDATE
         SET user_id = excluded.user_id, device = excluded.device, expires_at = excluded.expires_at, used_at = NULL
         WHERE pair_codes.used_at IS NOT NULL OR pair_codes.expires_at <= $5`,
      [code, user, device, expiresAt, now],
    );
    return r.rowCount === 1;
  }

  async claimPairing(code: string, device: string, now: Date): Promise<ClaimResult> {
    const r = await this.pool.query<{ user_id: string }>(
      `UPDATE pair_codes SET used_at = $3
       WHERE code = $1 AND used_at IS NULL AND expires_at > $3 AND device <> $2
       RETURNING user_id`,
      [code, device, now],
    );
    if (r.rows[0]) return { ok: true, user: r.rows[0].user_id };
    const why = await this.pool.query<{ used_at: Date | null; expires_at: Date; device: string }>(
      `SELECT used_at, expires_at, device FROM pair_codes WHERE code = $1`,
      [code],
    );
    const p = why.rows[0];
    if (!p) return { ok: false, reason: "missing" };
    if (p.used_at) return { ok: false, reason: "used" };
    if (p.expires_at <= now) return { ok: false, reason: "expired" };
    return { ok: false, reason: "same-device" };
  }

  async putExport(user: string, device: string, files: ExportFiles, updated: Date) {
    await this.pool.query(
      `INSERT INTO device_exports (user_id, device, files, updated_at) VALUES ($1, $2, $3, $4)
       ON CONFLICT (user_id, device) DO UPDATE SET files = excluded.files, updated_at = excluded.updated_at`,
      [user, device, JSON.stringify(files), updated],
    );
  }

  async listExports(user: string): Promise<DeviceExport[]> {
    const r = await this.pool.query<{ device: string; files: ExportFiles; updated_at: Date }>(
      `SELECT device, files, updated_at FROM device_exports WHERE user_id = $1 ORDER BY device`,
      [user],
    );
    return r.rows.map((row) => ({ device: row.device, files: row.files, updated: row.updated_at.toISOString() }));
  }

  async bump(key: string, bucket: string, cap: number) {
    if (cap <= 0) return null;
    const r = await this.pool.query<{ count: number }>(
      `INSERT INTO counters (key, bucket, count) VALUES ($1, $2, 1)
       ON CONFLICT (key, bucket) DO UPDATE SET count = counters.count + 1 WHERE counters.count < $3
       RETURNING count`,
      [key, bucket, cap],
    );
    return r.rows[0]?.count ?? null;
  }
}
