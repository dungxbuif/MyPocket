package ai

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
)

func TestLiveExtractionWhenExplicitlyEnabled(t *testing.T) {
	if os.Getenv("AI_LIVE_TEST") != "1" { t.Skip("set AI_LIVE_TEST=1 to exercise the configured text model") }
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil { t.Fatal(err) }
	cfg := config.Load()
	c := NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel})
	out, err := c.Extract(context.Background(), Input{Text: "Tôi vừa ăn trưa hết 35k bằng Cash.", Timezone: "Asia/Ho_Chi_Minh", Now: time.Date(2026, 9, 21, 10, 0, 0, 0, time.FixedZone("ICT", 7*3600)), Wallets: []entity.Wallet{{ID: "cash", Name: "Cash", Type: entity.WalletTypeBasic, Currency: "VND"}}, Categories: []entity.Category{{ID: "food", Name: "Ăn uống", Kind: entity.TransactionTypeExpense}}})
	if err != nil { t.Fatal(err) }
	if len(out.Drafts) != 1 || out.Drafts[0].Type != entity.TransactionTypeExpense || out.Drafts[0].Amount != 35000 || out.Drafts[0].WalletID != "cash" { t.Fatalf("unexpected live draft: %#v", out.Drafts) }
}
