package migrations_test

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func migrationSQL(t *testing.T, name string) string {
	t.Helper()
	contents, err := migrationFiles.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ToUpper(string(contents))
}

func TestAssetUnitMigrationPreservesTradesAndCorrections(t *testing.T) {
	sql := migrationSQL(t, "000043_asset_units.up.sql")
	for _, required := range []string{
		"ADD COLUMN IF NOT EXISTS UNIT_LABEL TEXT",
		"ADD COLUMN IF NOT EXISTS UNIT_CONFIRMED BOOLEAN NOT NULL DEFAULT FALSE",
		"ADD COLUMN IF NOT EXISTS PRICE_MODE TEXT NOT NULL DEFAULT 'MANUAL'",
		"UNIT_LABEL IS NULL OR CHAR_LENGTH(TRIM(UNIT_LABEL)) BETWEEN 1 AND 64",
		"PRICE_MODE IN ('MANUAL', 'AUTOMATIC')",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("000043_asset_units.up.sql missing additive compatibility contract %q", required)
		}
	}
	for _, forbidden := range []string{
		"UPDATE PORTFOLIO_ASSETS",
		"DELETE FROM PORTFOLIO_ASSETS",
		"DROP TABLE PORTFOLIO_TRADES",
		"DROP TABLE PORTFOLIO_TRADE_REVISIONS",
	} {
		if strings.Contains(sql, forbidden) {
			t.Errorf("000043_asset_units.up.sql must not rewrite ledger history: found %q", forbidden)
		}
	}
}

func TestAccumulatedAssetsMigrationRollbackIsScopedToAddedSchema(t *testing.T) {
	unitDown := migrationSQL(t, "000043_asset_units.down.sql")
	for _, column := range []string{"PRICE_MODE", "UNIT_CONFIRMED", "UNIT_LABEL"} {
		if !strings.Contains(unitDown, "DROP COLUMN IF EXISTS "+column) {
			t.Errorf("unit rollback missing DROP COLUMN for %s", column)
		}
	}
	for _, forbidden := range []string{"DROP TABLE PORTFOLIO_ASSETS", "DROP TABLE PORTFOLIO_TRADES", "DROP TABLE PORTFOLIO_TRADE_REVISIONS"} {
		if strings.Contains(unitDown, forbidden) {
			t.Errorf("unit rollback must not remove ledger tables: found %q", forbidden)
		}
	}

	marketDown := migrationSQL(t, "000044_asset_market.down.sql")
	for _, required := range []string{
		"DROP TABLE IF EXISTS MARKET_QUOTES",
		"DROP TABLE IF EXISTS MARKET_PRODUCTS",
		"DROP COLUMN IF EXISTS LATEST_PRICE_SOURCE",
		"DROP COLUMN IF EXISTS MARKET_PRODUCT_ID",
	} {
		if !strings.Contains(marketDown, required) {
			t.Errorf("market rollback missing scoped cleanup %q", required)
		}
	}
	for _, forbidden := range []string{"DROP TABLE PORTFOLIO_ASSETS", "DROP TABLE PORTFOLIO_TRADES", "DROP TABLE PORTFOLIO_TRADE_REVISIONS"} {
		if strings.Contains(marketDown, forbidden) {
			t.Errorf("market rollback must not remove ledger tables: found %q", forbidden)
		}
	}

	claimDown := migrationSQL(t, "000045_market_poll_claims.down.sql")
	if !strings.Contains(claimDown, "DROP TABLE IF EXISTS MARKET_POLL_CLAIMS") {
		t.Fatal("claim rollback must remove only market_poll_claims")
	}
	if strings.Contains(claimDown, "DROP TABLE PORTFOLIO_") {
		t.Fatal("claim rollback must not remove portfolio tables")
	}

	seedDown := migrationSQL(t, "000049_gold_market_catalog.down.sql")
	if strings.Contains(seedDown, "DELETE FROM PORTFOLIO_ASSETS") {
		t.Fatal("catalog rollback must not delete holdings")
	}
	for _, productID := range []string{"GOLD-BTMC-VRTL", "GOLD-BTMC-RING", "GOLD-BTMH-RING", "GOLD-BTMH-24K"} {
		if !strings.Contains(seedDown, productID) {
			t.Errorf("catalog rollback missing seeded product %s", productID)
		}
	}
}

func TestAccumulatedAssetsMigrationsPreserveLegacyRowsAndRollback(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL TEST_DATABASE_URL")
	}
	if os.Getenv("APP_ENV") == "production" {
		t.Skip("refuses to run accumulated-assets migration tests in production")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	schema := "accumulated_assets_migration_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	err = database.Connection(func(connection *gorm.DB) error {
		quotedSchema := `"` + schema + `"`
		if err := connection.Exec("CREATE SCHEMA " + quotedSchema).Error; err != nil {
			return err
		}
		defer connection.Exec("DROP SCHEMA " + quotedSchema + " CASCADE")
		if err := connection.Exec("SET search_path TO " + quotedSchema).Error; err != nil {
			return err
		}
		for _, statement := range []string{
			`CREATE TABLE "user" (id text PRIMARY KEY)`,
			`CREATE TABLE portfolio_assets (
				id text PRIMARY KEY, owner_id text NOT NULL REFERENCES "user"(id), symbol text NOT NULL, name text NOT NULL,
				latest_price bigint, latest_price_at timestamptz, version bigint NOT NULL DEFAULT 1,
				created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
			)`,
			`CREATE TABLE portfolio_trades (
				id text PRIMARY KEY, owner_id text NOT NULL REFERENCES "user"(id), asset_id text NOT NULL REFERENCES portfolio_assets(id),
				side text NOT NULL, quantity_scaled bigint NOT NULL, unit_price bigint NOT NULL, fee bigint NOT NULL DEFAULT 0,
				occurred_at timestamptz NOT NULL, note text, created_at timestamptz NOT NULL DEFAULT now()
			)`,
			`CREATE TABLE portfolio_trade_revisions (
				id text PRIMARY KEY, owner_id text NOT NULL REFERENCES "user"(id), asset_id text NOT NULL REFERENCES portfolio_assets(id),
				source_trade_id text NOT NULL, revision bigint NOT NULL, operation text NOT NULL, reason text NOT NULL,
				actor_id text NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
			)`,
			`INSERT INTO "user" (id) VALUES ('owner-1')`,
			`INSERT INTO portfolio_assets (id, owner_id, symbol, name, latest_price, latest_price_at, version) VALUES ('asset-1', 'owner-1', 'GOLD', 'Legacy gold', 8000000, '2026-10-01T00:00:00Z', 3)`,
			`INSERT INTO portfolio_trades (id, owner_id, asset_id, side, quantity_scaled, unit_price, fee, occurred_at, note) VALUES ('trade-1', 'owner-1', 'asset-1', 'buy', 150000000, 7000000, 1000, '2026-10-01T00:00:00Z', 'fractional legacy quantity')`,
			`INSERT INTO portfolio_trades (id, owner_id, asset_id, side, quantity_scaled, unit_price, fee, occurred_at, note) VALUES ('trade-2', 'owner-1', 'asset-1', 'sell', 50000000, 8000000, 500, '2026-10-02T00:00:00Z', 'legacy sell')`,
			`INSERT INTO portfolio_trade_revisions (id, owner_id, asset_id, source_trade_id, revision, operation, reason, actor_id) VALUES ('revision-1', 'owner-1', 'asset-1', 'trade-1', 1, 'replace', 'receipt correction', 'owner-1')`,
		} {
			if err := connection.Exec(statement).Error; err != nil {
				return err
			}
		}

		for _, migration := range []string{"000043_asset_units.up.sql", "000044_asset_market.up.sql", "000045_market_poll_claims.up.sql", "000049_gold_market_catalog.up.sql"} {
			contents, readErr := migrationFiles.ReadFile(migration)
			if readErr != nil {
				return readErr
			}
			if err := connection.Exec(string(contents)).Error; err != nil {
				return err
			}
		}

		var legacy struct {
			Symbol        string
			LatestPrice   int64
			Version       int64
			UnitLabel     *string
			UnitConfirmed bool
			PriceMode     string
		}
		if err := connection.Raw("SELECT symbol, latest_price, version, unit_label, unit_confirmed, price_mode FROM portfolio_assets WHERE id = 'asset-1'").Scan(&legacy).Error; err != nil {
			return err
		}
		if legacy.Symbol != "GOLD" || legacy.LatestPrice != 8000000 || legacy.Version != 3 || legacy.UnitLabel != nil || legacy.UnitConfirmed || legacy.PriceMode != "manual" {
			return &migrationAssertionError{message: "legacy asset changed or did not receive safe manual defaults"}
		}
		for table, want := range map[string]int64{"portfolio_trades": 2, "portfolio_trade_revisions": 1, "market_products": 4} {
			var count int64
			if err := connection.Raw("SELECT COUNT(*) FROM " + table).Scan(&count).Error; err != nil {
				return err
			}
			if count != want {
				return &migrationAssertionError{message: table + " count changed during forward migration"}
			}
		}

		for _, migration := range []string{"000049_gold_market_catalog.down.sql", "000045_market_poll_claims.down.sql", "000044_asset_market.down.sql", "000043_asset_units.down.sql"} {
			contents, readErr := migrationFiles.ReadFile(migration)
			if readErr != nil {
				return readErr
			}
			if err := connection.Exec(string(contents)).Error; err != nil {
				return err
			}
		}

		for table, want := range map[string]int64{"portfolio_assets": 1, "portfolio_trades": 2, "portfolio_trade_revisions": 1} {
			var count int64
			if err := connection.Raw("SELECT COUNT(*) FROM " + table).Scan(&count).Error; err != nil {
				return err
			}
			if count != want {
				return &migrationAssertionError{message: table + " count changed during rollback"}
			}
		}
		var droppedColumns int64
		if err := connection.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'portfolio_assets' AND column_name IN ('unit_label', 'unit_confirmed', 'price_mode', 'market_product_id', 'latest_price_source')").Scan(&droppedColumns).Error; err != nil {
			return err
		}
		if droppedColumns != 0 {
			return &migrationAssertionError{message: "rollback left accumulated-assets columns behind"}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type migrationAssertionError struct{ message string }

func (e *migrationAssertionError) Error() string { return e.message }
