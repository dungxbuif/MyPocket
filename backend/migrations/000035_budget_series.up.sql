CREATE TABLE budget_series (
 id text PRIMARY KEY,
 owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
 cadence text NOT NULL CHECK(cadence IN ('weekly','monthly','quarterly','yearly')),
 start_date date NOT NULL,
 next_start_date date NOT NULL,
 active boolean NOT NULL DEFAULT true,
 version bigint NOT NULL DEFAULT 1 CHECK(version > 0),
 generation_error text,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX budget_series_owner_idx ON budget_series(owner_id,active);
CREATE TABLE budget_series_configs (
 id text PRIMARY KEY,
 series_id text NOT NULL REFERENCES budget_series(id) ON DELETE CASCADE,
 effective_start_date date NOT NULL,
 name text NOT NULL,
 limit_amount bigint NOT NULL CHECK(limit_amount>0 AND limit_amount<=9007199254740991),
 wallet_id text,
 category_id text,
 carryover_enabled boolean NOT NULL DEFAULT false,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(series_id,effective_start_date)
);
-- Scope IDs in revisions are immutable snapshots, deliberately not SET NULL FKs.
-- Stop before existing budget-scope cascades; never widen a deleted scope.
-- Superseded revisions cannot stop a series after their last period generated.
CREATE OR REPLACE FUNCTION stop_budget_series_for_deleted_scope() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 UPDATE budget_series AS s SET active=false,version=version+1,
 generation_error='scope_deleted',updated_at=now()
 WHERE s.active=true AND EXISTS (
  SELECT 1 FROM budget_series_configs AS cfg
  WHERE cfg.series_id=s.id
    AND ((TG_TABLE_NAME='wallets' AND cfg.wallet_id=OLD.id)
      OR (TG_TABLE_NAME='categories' AND cfg.category_id=OLD.id))
    AND (cfg.effective_start_date>=s.next_start_date
      OR cfg.effective_start_date=(
       SELECT max(current_cfg.effective_start_date)
       FROM budget_series_configs AS current_cfg
       WHERE current_cfg.series_id=s.id
         AND current_cfg.effective_start_date<=s.next_start_date
      ))
 );
 RETURN OLD;
END;
$$;
CREATE TRIGGER budget_series_wallet_delete BEFORE DELETE ON wallets FOR EACH ROW EXECUTE FUNCTION stop_budget_series_for_deleted_scope();
CREATE TRIGGER budget_series_category_delete BEFORE DELETE ON categories FOR EACH ROW EXECUTE FUNCTION stop_budget_series_for_deleted_scope();
ALTER TABLE budgets ADD COLUMN series_id text REFERENCES budget_series(id) ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED;
ALTER TABLE budgets ADD COLUMN carryover_enabled boolean NOT NULL DEFAULT false;
CREATE UNIQUE INDEX budgets_series_period_unique ON budgets(series_id,start_date) WHERE series_id IS NOT NULL;
