CREATE TABLE jars (
  id text PRIMARY KEY,
  owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(owner_id, id)
);

CREATE TABLE jar_months (
  owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  month date NOT NULL CHECK (extract(day FROM month) = 1),
  PRIMARY KEY(owner_id, month)
);

CREATE TABLE jar_month_configs (
  owner_id text NOT NULL,
  month date NOT NULL,
  jar_id text NOT NULL,
  name text NOT NULL CHECK (length(trim(name)) > 0),
  allocation_mode text NOT NULL CHECK (allocation_mode IN ('none', 'fixed', 'percent')),
  allocation_amount bigint,
  allocation_percent_bps integer,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY(owner_id, month, jar_id),
  FOREIGN KEY(owner_id, month) REFERENCES jar_months(owner_id, month) ON DELETE CASCADE,
  FOREIGN KEY(owner_id, jar_id) REFERENCES jars(owner_id, id) ON DELETE CASCADE,
  CHECK (
    (allocation_mode = 'none' AND allocation_amount IS NULL AND allocation_percent_bps IS NULL) OR
    (allocation_mode = 'fixed' AND allocation_amount IS NOT NULL AND allocation_amount > 0 AND allocation_percent_bps IS NULL) OR
    (allocation_mode = 'percent' AND allocation_amount IS NULL AND allocation_percent_bps IS NOT NULL AND allocation_percent_bps BETWEEN 0 AND 10000)
  )
);
CREATE INDEX jar_month_configs_owner_month_active ON jar_month_configs(owner_id, month, active);

ALTER TABLE transactions ADD COLUMN jar_id text;
ALTER TABLE transactions ADD CONSTRAINT transactions_owner_jar_fk
  FOREIGN KEY(owner_id, jar_id) REFERENCES jars(owner_id, id);
CREATE INDEX transactions_owner_jar_idx ON transactions(owner_id, jar_id);
