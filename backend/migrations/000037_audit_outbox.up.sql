CREATE TABLE audit_outbox (
 id text PRIMARY KEY,
 stream text NOT NULL CHECK (stream IN ('mypocket:audit:feedback','mypocket:audit:advisor')),
 event_id text NOT NULL,
 payload jsonb NOT NULL,
 attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
 available_at timestamptz NOT NULL DEFAULT now(),
 lease_until timestamptz,
 delivered_at timestamptz,
 last_error text,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(stream,event_id)
);
CREATE INDEX audit_outbox_pending_idx ON audit_outbox(available_at,created_at)
 WHERE delivered_at IS NULL;
