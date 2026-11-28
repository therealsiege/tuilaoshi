-- Deleting an account (DELETE /api/account) moves its version on; a token
-- is refused unless it carries the current one (0 when there is no row).
-- The row outlives the deletion so the old tokens stay refused.
CREATE TABLE IF NOT EXISTS token_versions (
  user_id  text PRIMARY KEY,
  version  integer NOT NULL
);
