-- TuiLaoShi server schema. Safe to run repeatedly.

CREATE TABLE IF NOT EXISTS pair_codes (
  code        text PRIMARY KEY,
  user_id     text NOT NULL,
  device      text NOT NULL,
  expires_at  timestamptz NOT NULL,
  used_at     timestamptz
);
CREATE INDEX IF NOT EXISTS pair_codes_expires_at ON pair_codes (expires_at);

-- One exported progress folder per device: relative path -> file contents.
CREATE TABLE IF NOT EXISTS device_exports (
  user_id     text NOT NULL,
  device      text NOT NULL,
  files       jsonb NOT NULL,
  updated_at  timestamptz NOT NULL,
  PRIMARY KEY (user_id, device)
);

-- Counters for the Claude daily cap and pairing-claim attempts.
CREATE TABLE IF NOT EXISTS counters (
  key     text NOT NULL,
  bucket  text NOT NULL,
  count   integer NOT NULL,
  PRIMARY KEY (key, bucket)
);
