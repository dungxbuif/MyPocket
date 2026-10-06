ALTER TABLE wallets
  ADD COLUMN IF NOT EXISTS goal_notifications_enabled boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS notifications (
  id text PRIMARY KEY,
  owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  kind text NOT NULL,
  title text NOT NULL,
  body text NOT NULL,
  dedupe_key text NOT NULL,
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notifications_owner_created_idx
  ON notifications(owner_id, created_at DESC, id DESC);

CREATE UNIQUE INDEX IF NOT EXISTS notifications_owner_dedupe_idx
  ON notifications(owner_id, dedupe_key);
