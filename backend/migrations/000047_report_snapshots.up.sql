-- Durable owner-scoped report revisions and immutable, bounded report snapshots.
-- Migration numbers 000043-000046 are owned by other in-flight work.
CREATE TABLE report_revisions (
    owner_id text PRIMARY KEY REFERENCES "user"(id) ON DELETE CASCADE,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO report_revisions(owner_id, revision, updated_at)
SELECT id, 1, now() FROM "user"
ON CONFLICT (owner_id) DO NOTHING;

CREATE TABLE report_snapshots (
    id text PRIMARY KEY,
    owner_id text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    month text NOT NULL CHECK (month ~ '^[0-9]{4}-[0-9]{2}$'),
    wallet_id text NOT NULL DEFAULT '',
    granularity text NOT NULL CHECK (granularity IN ('week', 'day')),
    timezone text NOT NULL,
    as_of_day date NOT NULL,
    data_revision bigint NOT NULL CHECK (data_revision > 0),
    rule_version text NOT NULL,
    as_of timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    payload jsonb NOT NULL,
    CHECK (expires_at > created_at)
);
CREATE INDEX report_snapshots_lookup_idx ON report_snapshots(owner_id, month, wallet_id, granularity, timezone, as_of_day, data_revision, rule_version, created_at DESC);
CREATE INDEX report_snapshots_expiry_idx ON report_snapshots(expires_at, created_at);

CREATE OR REPLACE FUNCTION bump_report_revision() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    affected_owner text;
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF TG_TABLE_NAME = 'user' THEN
            affected_owner := OLD.id;
        ELSIF TG_TABLE_NAME = 'category_wallets' THEN
            SELECT owner_id INTO affected_owner FROM wallets WHERE id = OLD.wallet_id;
        ELSIF TG_TABLE_NAME = 'budget_series_configs' THEN
            SELECT owner_id INTO affected_owner FROM budget_series WHERE id = OLD.series_id;
        ELSE
            affected_owner := OLD.owner_id;
        END IF;
    ELSE
        IF TG_TABLE_NAME = 'user' THEN
            affected_owner := NEW.id;
        ELSIF TG_TABLE_NAME = 'category_wallets' THEN
            SELECT owner_id INTO affected_owner FROM wallets WHERE id = NEW.wallet_id;
        ELSIF TG_TABLE_NAME = 'budget_series_configs' THEN
            SELECT owner_id INTO affected_owner FROM budget_series WHERE id = NEW.series_id;
        ELSE
            affected_owner := NEW.owner_id;
        END IF;
    END IF;

    IF TG_TABLE_NAME = 'categories' AND affected_owner IS NULL THEN
        UPDATE report_revisions SET revision = revision + 1, updated_at = now();
    ELSIF affected_owner IS NOT NULL THEN
        INSERT INTO report_revisions(owner_id, revision, updated_at)
        VALUES (affected_owner, 2, now())
        ON CONFLICT (owner_id) DO UPDATE SET revision = report_revisions.revision + 1, updated_at = now();
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER report_revision_transactions AFTER INSERT OR UPDATE OR DELETE ON transactions FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_wallets AFTER INSERT OR UPDATE OR DELETE ON wallets FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_categories AFTER INSERT OR UPDATE OR DELETE ON categories FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_category_wallets AFTER INSERT OR UPDATE OR DELETE ON category_wallets FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_budgets AFTER INSERT OR UPDATE OR DELETE ON budgets FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_budget_series AFTER INSERT OR UPDATE OR DELETE ON budget_series FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_budget_series_configs AFTER INSERT OR UPDATE OR DELETE ON budget_series_configs FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_jars AFTER INSERT OR UPDATE OR DELETE ON jars FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_jar_months AFTER INSERT OR UPDATE OR DELETE ON jar_months FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_jar_month_configs AFTER INSERT OR UPDATE OR DELETE ON jar_month_configs FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
CREATE TRIGGER report_revision_user AFTER INSERT OR UPDATE OR DELETE ON "user" FOR EACH ROW EXECUTE FUNCTION bump_report_revision();
