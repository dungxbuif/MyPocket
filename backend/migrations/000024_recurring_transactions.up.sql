CREATE TABLE IF NOT EXISTS recurring_schedules (
  id text PRIMARY KEY,
  owner_id text NOT NULL,
  name text NOT NULL,
  wallet_id text NOT NULL,
  category_id text,
  type text NOT NULL,
  amount bigint NOT NULL CHECK (amount > 0),
  note text,
  frequency text NOT NULL,
  interval integer NOT NULL DEFAULT 1 CHECK (interval > 0),
  next_run_at timestamptz NOT NULL,
  ends_at timestamptz,
  anchor_day integer NOT NULL CHECK (anchor_day BETWEEN 1 AND 31),
  anchor_month integer NOT NULL CHECK (anchor_month BETWEEN 1 AND 12),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT recurring_schedules_type_check CHECK (type IN ('income', 'expense')),
  CONSTRAINT recurring_schedules_frequency_check CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
  FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS recurring_schedules_owner_idx ON recurring_schedules(owner_id);
CREATE INDEX IF NOT EXISTS recurring_schedules_due_idx ON recurring_schedules(active, next_run_at);

CREATE TABLE IF NOT EXISTS recurring_occurrences (
  id text PRIMARY KEY,
  schedule_id text NOT NULL,
  due_at timestamptz NOT NULL,
  transaction_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT recurring_occurrence_due_unique UNIQUE (schedule_id, due_at),
  FOREIGN KEY (schedule_id) REFERENCES recurring_schedules(id) ON DELETE CASCADE,
  FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS recurring_occurrences_transaction_idx ON recurring_occurrences(transaction_id);
