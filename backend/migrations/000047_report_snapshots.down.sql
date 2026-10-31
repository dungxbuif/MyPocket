-- Disposable fixtures only. Never roll this back on a production database.
DROP TRIGGER IF EXISTS report_revision_transactions ON transactions;
DROP TRIGGER IF EXISTS report_revision_wallets ON wallets;
DROP TRIGGER IF EXISTS report_revision_categories ON categories;
DROP TRIGGER IF EXISTS report_revision_category_wallets ON category_wallets;
DROP TRIGGER IF EXISTS report_revision_budgets ON budgets;
DROP TRIGGER IF EXISTS report_revision_budget_series ON budget_series;
DROP TRIGGER IF EXISTS report_revision_budget_series_configs ON budget_series_configs;
DROP TRIGGER IF EXISTS report_revision_jars ON jars;
DROP TRIGGER IF EXISTS report_revision_jar_months ON jar_months;
DROP TRIGGER IF EXISTS report_revision_jar_month_configs ON jar_month_configs;
DROP TRIGGER IF EXISTS report_revision_user ON "user";
DROP FUNCTION IF EXISTS bump_report_revision();
DROP TABLE IF EXISTS report_snapshots;
DROP TABLE IF EXISTS report_revisions;
