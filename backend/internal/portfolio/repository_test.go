package portfolio_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"mypocket/internal/finance"
	"mypocket/internal/platform/db"
	"mypocket/internal/portfolio"
)

func TestRepositoryIsolatesPositionsByUserAndKeepsWalletsUnchanged(t *testing.T) {
	conn := migratedPortfolioPostgres(t)
	repo := portfolio.NewRepository(conn)
	financeRepo := finance.NewRepository(conn)
	userA := createPortfolioUser(t, conn, "asset-owner-a@example.com")
	userB := createPortfolioUser(t, conn, "asset-owner-b@example.com")
	wallet := createPortfolioWallet(t, financeRepo, userA)

	position := createGoldPosition(t, repo, userA)
	if _, err := repo.GetPosition(context.Background(), userB, position.ID); !errors.Is(err, portfolio.ErrForbidden) {
		t.Fatalf("expected forbidden cross-user read, got %v", err)
	}
	if _, err := repo.AddTrade(context.Background(), userA, position.ID, portfolio.AddTradeInput{
		Side:         portfolio.TradeBuy,
		Quantity:     "2",
		UnitPriceVND: 70_000_000,
		FeeVND:       20_000,
		OccurredAt:   fixedPortfolioTime(),
		BaseVersion:  position.Version,
	}); err != nil {
		t.Fatalf("add gold buy: %v", err)
	}
	if _, err := repo.AddPrice(context.Background(), userA, position.ID, portfolio.AddPriceInput{
		UnitPriceVND: 72_000_000,
		PricedAt:     fixedPortfolioTime().Add(time.Hour),
		Source:       "manual",
	}); err != nil {
		t.Fatalf("add manual price: %v", err)
	}
	assertPortfolioWalletBalance(t, conn, wallet.ID, 0, 1)
}

func TestRepositoryReplaysMovingAverageLedgerAndManualPriceHistory(t *testing.T) {
	conn := migratedPortfolioPostgres(t)
	repo := portfolio.NewRepository(conn)
	userID := createPortfolioUser(t, conn, "asset-ledger@example.com")
	position := createGoldPosition(t, repo, userID)
	base := fixedPortfolioTime()

	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeBuy, "2", 70_000_000, 20_000, base)
	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeBuy, "1.5", 72_000_000, 10_000, base.Add(time.Hour))
	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeSell, "1.2", 75_000_000, 15_000, base.Add(2*time.Hour))

	if position.Summary.Quantity != "2.3" {
		t.Fatalf("quantity = %s, want 2.3", position.Summary.Quantity)
	}
	if position.Summary.CostBasisVND != 162_991_143 || position.Summary.RealizedPNLVND != 4_946_143 {
		t.Fatalf("summary = %#v", position.Summary)
	}
	position = addPortfolioPrice(t, repo, userID, position.ID, position.Version, 74_000_000, base.Add(3*time.Hour), "manual", "")
	if position.Summary.MarketValueVND == nil || *position.Summary.MarketValueVND != 170_200_000 {
		t.Fatalf("market value = %#v, want 170200000", position.Summary.MarketValueVND)
	}
	if position.Summary.UnrealizedPNLVND == nil || *position.Summary.UnrealizedPNLVND != 7_208_857 {
		t.Fatalf("unrealized pnl = %#v, want 7208857", position.Summary.UnrealizedPNLVND)
	}
	position = addPortfolioPrice(t, repo, userID, position.ID, position.Version, 76_000_000, base.Add(4*time.Hour), "manual", "")
	if len(position.PriceHistory) != 2 || position.LatestPrice == nil || position.LatestPrice.UnitPriceVND != 76_000_000 {
		t.Fatalf("expected append-only latest price history, got latest=%#v history=%#v", position.LatestPrice, position.PriceHistory)
	}
}

func TestRepositoryRejectsOversellAndRecomputesAfterHistoricalCorrection(t *testing.T) {
	conn := migratedPortfolioPostgres(t)
	repo := portfolio.NewRepository(conn)
	userID := createPortfolioUser(t, conn, "asset-correction@example.com")
	position := createGoldPosition(t, repo, userID)
	base := fixedPortfolioTime()

	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeBuy, "1", 100, 0, base)
	if _, err := repo.AddTrade(context.Background(), userID, position.ID, portfolio.AddTradeInput{
		Side:         portfolio.TradeSell,
		Quantity:     "2",
		UnitPriceVND: 100,
		OccurredAt:   base.Add(time.Hour),
		BaseVersion:  position.Version,
	}); !errors.Is(err, portfolio.ErrOversell) {
		t.Fatalf("expected oversell, got %v", err)
	}

	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeBuy, "1", 200, 0, base.Add(time.Hour))
	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeSell, "1", 300, 0, base.Add(2*time.Hour))
	if position.Summary.CostBasisVND != 150 || position.Summary.RealizedPNLVND != 150 {
		t.Fatalf("before correction summary = %#v", position.Summary)
	}
	firstTrade := position.Trades[0]
	position, err := repo.UpdateTrade(context.Background(), userID, position.ID, firstTrade.ID, portfolio.UpdateTradeInput{
		Side:         portfolio.TradeBuy,
		Quantity:     "2",
		UnitPriceVND: 100,
		OccurredAt:   base,
		BaseVersion:  firstTrade.Version,
	})
	if err != nil {
		t.Fatalf("correct historical trade: %v", err)
	}
	if position.Summary.Quantity != "2" || position.Summary.CostBasisVND != 267 || position.Summary.RealizedPNLVND != 167 {
		t.Fatalf("after correction summary = %#v", position.Summary)
	}
}

func TestRepositoryArchivesPositionOutsideActiveSummary(t *testing.T) {
	conn := migratedPortfolioPostgres(t)
	repo := portfolio.NewRepository(conn)
	userID := createPortfolioUser(t, conn, "asset-archive@example.com")
	position := createGoldPosition(t, repo, userID)
	position = addPortfolioTrade(t, repo, userID, position.ID, position.Version, portfolio.TradeBuy, "1", 10_000, 0, fixedPortfolioTime())
	position = addPortfolioPrice(t, repo, userID, position.ID, position.Version, 11_000, fixedPortfolioTime(), "manual", "")
	before, err := repo.Summary(context.Background(), userID)
	if err != nil {
		t.Fatalf("portfolio summary before archive: %v", err)
	}
	if before.InvestmentMarketValueVND != 11_000 {
		t.Fatalf("summary before archive = %#v", before)
	}
	if err := repo.ArchivePosition(context.Background(), userID, position.ID, position.Version); err != nil {
		t.Fatalf("archive position: %v", err)
	}
	after, err := repo.Summary(context.Background(), userID)
	if err != nil {
		t.Fatalf("portfolio summary after archive: %v", err)
	}
	if after.InvestmentMarketValueVND != 0 || after.PositionCount != 0 {
		t.Fatalf("summary after archive = %#v", after)
	}
	loaded, err := repo.GetPosition(context.Background(), userID, position.ID)
	if err != nil {
		t.Fatalf("archived detail retained: %v", err)
	}
	if loaded.ArchivedAt == nil || len(loaded.Trades) != 1 || len(loaded.PriceHistory) != 1 {
		t.Fatalf("archive should retain history: %#v", loaded)
	}
}

func migratedPortfolioPostgres(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("MYPOCKET_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MYPOCKET_TEST_DATABASE_URL is not set; portfolio repository integration proof skipped")
	}
	conn, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(context.Background(), `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := db.Migrate(context.Background(), conn, os.DirFS("../../migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}

func createPortfolioUser(t *testing.T, conn *sql.DB, email string) string {
	t.Helper()
	var userID string
	err := conn.QueryRowContext(context.Background(), `
		INSERT INTO users (google_subject, email, email_verified)
		VALUES ($1, $2, true)
		RETURNING id::text
	`, "subject-"+email, email).Scan(&userID)
	if err != nil {
		t.Fatalf("create portfolio user: %v", err)
	}
	return userID
}

func createPortfolioWallet(t *testing.T, repo *finance.Repository, userID string) finance.Wallet {
	t.Helper()
	wallet, err := repo.CreateWallet(context.Background(), userID, finance.CreateWalletInput{
		Name: "Tiền mặt",
		Type: finance.WalletCash,
	})
	if err != nil {
		t.Fatalf("create portfolio wallet: %v", err)
	}
	return wallet
}

func createGoldPosition(t *testing.T, repo *portfolio.Repository, userID string) portfolio.Position {
	t.Helper()
	position, err := repo.CreatePosition(context.Background(), userID, portfolio.CreatePositionInput{
		Type: portfolio.AssetGold,
		Name: "SJC 9999",
		Unit: "tael",
	})
	if err != nil {
		t.Fatalf("create gold position: %v", err)
	}
	return position
}

func addPortfolioTrade(t *testing.T, repo *portfolio.Repository, userID string, assetID string, baseVersion int64, side portfolio.TradeSide, quantity string, price int64, fee int64, occurredAt time.Time) portfolio.Position {
	t.Helper()
	position, err := repo.AddTrade(context.Background(), userID, assetID, portfolio.AddTradeInput{
		Side:         side,
		Quantity:     quantity,
		UnitPriceVND: price,
		FeeVND:       fee,
		OccurredAt:   occurredAt,
		BaseVersion:  baseVersion,
	})
	if err != nil {
		t.Fatalf("add portfolio trade: %v", err)
	}
	return position
}

func addPortfolioPrice(t *testing.T, repo *portfolio.Repository, userID string, assetID string, baseVersion int64, price int64, pricedAt time.Time, source string, quoteID string) portfolio.Position {
	t.Helper()
	position, err := repo.AddPrice(context.Background(), userID, assetID, portfolio.AddPriceInput{
		UnitPriceVND:    price,
		PricedAt:        pricedAt,
		Source:          source,
		ProviderQuoteID: quoteID,
		BaseVersion:     baseVersion,
	})
	if err != nil {
		t.Fatalf("add portfolio price: %v", err)
	}
	return position
}

func fixedPortfolioTime() time.Time {
	return time.Date(2026, 8, 31, 9, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
}

func assertPortfolioWalletBalance(t *testing.T, conn *sql.DB, walletID string, wantBalanceVND int64, wantVersion int64) {
	t.Helper()
	var balanceVND int64
	var version int64
	err := conn.QueryRowContext(context.Background(), `
		SELECT balance_vnd, version
		FROM wallets
		WHERE id = $1
	`, walletID).Scan(&balanceVND, &version)
	if err != nil {
		t.Fatalf("load wallet balance: %v", err)
	}
	if balanceVND != wantBalanceVND || version != wantVersion {
		t.Fatalf("wallet balance/version = %d/%d, want %d/%d", balanceVND, version, wantBalanceVND, wantVersion)
	}
}
