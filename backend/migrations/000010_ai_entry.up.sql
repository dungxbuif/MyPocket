CREATE TABLE ai_entry_sessions (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 request_token text NOT NULL DEFAULT '',
 processing_until timestamptz,
 error text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(id, owner_id)
);
CREATE INDEX ai_entry_sessions_owner_recent ON ai_entry_sessions(owner_id, updated_at DESC);
CREATE TABLE ai_entry_messages (
 id text PRIMARY KEY,
 session_id text NOT NULL REFERENCES ai_entry_sessions(id) ON DELETE CASCADE,
 role text NOT NULL CHECK(role IN ('user','assistant')),
 content text NOT NULL,
 source_text text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ai_entry_messages_session ON ai_entry_messages(session_id,created_at);
CREATE TABLE ai_entry_proposals (
 id text PRIMARY KEY,
 session_id text NOT NULL,
 owner_id text NOT NULL,
 version integer NOT NULL DEFAULT 1 CHECK(version>0),
 status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected')),
 draft jsonb NOT NULL,
 questions jsonb NOT NULL DEFAULT '[]',
 transaction_id text UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(session_id, owner_id) REFERENCES ai_entry_sessions(id, owner_id) ON DELETE CASCADE,
 CHECK((status='approved') = (transaction_id IS NOT NULL))
);
-- transaction_id intentionally remains a receipt after the ledger row is deleted;
-- replaying approval must never resurrect a deleted transaction.
CREATE INDEX ai_entry_proposals_session ON ai_entry_proposals(session_id,created_at);
CREATE TABLE ai_entry_requests (
 session_id text NOT NULL REFERENCES ai_entry_sessions(id) ON DELETE CASCADE,
 request_id text NOT NULL,
 hash text NOT NULL,
 token text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(session_id,request_id)
);
