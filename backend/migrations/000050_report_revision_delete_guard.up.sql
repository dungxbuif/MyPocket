-- Report revisions are owner-scoped. Deleting the owner must not try to
-- recreate a revision row after the user FK has been removed.
CREATE OR REPLACE FUNCTION bump_report_revision() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    affected_owner text;
BEGIN
    IF TG_OP = 'DELETE' AND TG_TABLE_NAME = 'user' THEN
        RETURN OLD;
    END IF;

    IF TG_OP = 'DELETE' THEN
        IF TG_TABLE_NAME = 'category_wallets' THEN
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
    ELSIF affected_owner IS NOT NULL AND EXISTS (SELECT 1 FROM "user" WHERE id = affected_owner) THEN
        INSERT INTO report_revisions(owner_id, revision, updated_at)
        VALUES (affected_owner, 2, now())
        ON CONFLICT (owner_id) DO UPDATE SET revision = report_revisions.revision + 1, updated_at = now();
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$;
