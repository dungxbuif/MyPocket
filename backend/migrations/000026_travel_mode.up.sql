CREATE TABLE IF NOT EXISTS travel_events (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  name text NOT NULL,
  context text,
  starts_on date,
  ends_on date,
  active boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT travel_events_name_check CHECK (char_length(trim(name)) > 0),
  CONSTRAINT travel_events_date_check CHECK (ends_on IS NULL OR starts_on IS NULL OR ends_on >= starts_on),
  FOREIGN KEY (owner_id) REFERENCES "user"(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS travel_events_owner_idx ON travel_events(owner_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS travel_events_one_active_idx ON travel_events(owner_id) WHERE active = true;

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS travel_event_id text;
CREATE INDEX IF NOT EXISTS transactions_travel_event_idx ON transactions(travel_event_id);
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'transactions_travel_event_fk'
  ) THEN
    ALTER TABLE transactions
      ADD CONSTRAINT transactions_travel_event_fk
      FOREIGN KEY (travel_event_id) REFERENCES travel_events(id) ON DELETE SET NULL;
  END IF;
END $$;
