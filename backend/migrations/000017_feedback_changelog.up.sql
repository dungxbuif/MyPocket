CREATE TABLE changelogs (
  id text PRIMARY KEY,
  version text NOT NULL UNIQUE,
  title text NOT NULL,
  description text NOT NULL,
  published_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feedback (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('bug', 'feature', 'improvement')),
  title text NOT NULL,
  description text NOT NULL,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'triaged', 'in_progress', 'fixed', 'rejected')),
  fixed_at timestamptz,
  changelog_id text REFERENCES changelogs(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX feedback_user_created_idx ON feedback(user_id, created_at DESC);
CREATE INDEX feedback_status_created_idx ON feedback(status, created_at ASC);
CREATE INDEX feedback_changelog_idx ON feedback(changelog_id);
