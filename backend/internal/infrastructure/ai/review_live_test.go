package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func reviewClient(t *testing.T) (*Client, config.Config) {
	t.Helper()
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("cannot load local provider configuration")
	}
	cfg := config.Load()
	return NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel}), cfg
}

func TestLiveReviewFirstOverlap(t *testing.T) {
	if os.Getenv("AI_REVIEW_LIVE") != "1" {
		t.Skip("opt-in provider evaluation")
	}
	c, _ := reviewClient(t)
	out, err := c.Extract(context.Background(), Input{
		Text:     "Ảnh 1 - Thẻ tín dụng\n18/09/2026 PHUC LONG mã GD B201 -55,000 VND\n19/09/2026 PHUC LONG mã GD B202 -55,000 VND\nẢnh 2 - Thẻ tín dụng (ảnh cuộn chồng lặp)\n19/09/2026 PHUC LONG mã GD B202 -55,000 VND\nẢnh 3 - Tài khoản thanh toán\nWINMART mua hàng -83,000 VND (ngày bị cắt)\nSố dư khả dụng 9,000,000 VND",
		Timezone: "Asia/Ho_Chi_Minh", Now: time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC),
		Wallets: []entity.Wallet{{ID: "cash", Name: "Tiền mặt", Type: "basic", Currency: "VND"}},
	})
	if err != nil {
		t.Fatalf("extraction failed: %v", err)
	}
	if len(out.Drafts) != 3 {
		t.Fatalf("expected 3 distinct drafts including incomplete row, got %d (reply withheld)", len(out.Drafts))
	}
	var coffee, shop int
	dates := make(map[string]bool)
	for _, d := range out.Drafts {
		if d.WalletID != "cash" {
			t.Fatal("draft did not prefill sole wallet")
		}
		switch d.Amount {
		case 55000:
			coffee++
			parsed, err := time.Parse(time.RFC3339, d.OccurredAt)
			if err != nil {
				t.Fatal("dated purchase lost its date")
			}
			dates[parsed.Format("2006-01-02")] = true
		case 83000:
			shop++
			if d.OccurredAt != "" || len(d.Questions) == 0 {
				t.Fatal("missing date must stay blank with a review question")
			}
		default:
			t.Fatal("invented or balance-derived amount")
		}
	}
	if coffee != 2 || shop != 1 {
		t.Fatal("distinct purchases lost or duplicate retained")
	}
	if !dates["2026-09-18"] || !dates["2026-09-19"] {
		t.Fatal("separate purchase dates not preserved")
	}
	t.Log("live overlap: 3 drafts, repeated reference collapsed, separate purchases and incomplete row preserved")
}

// Explicitly enabled replay reuses stored OCR; only SELECTs and one model request.
// Neither source text nor model reply/draft values are printed.
func TestLiveReviewStoredOCR(t *testing.T) {
	id := os.Getenv("AI_REVIEW_PROCESS_ID")
	if id == "" {
		t.Skip("requires explicit stored process ID")
	}
	c, cfg := reviewClient(t)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot open local database")
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	var process entity.AIEntrySession
	if err := db.Where("id = ?", id).First(&process).Error; err != nil {
		t.Fatal("process unavailable")
	}
	var account entity.User
	if err := db.Where("id = ?", process.OwnerID).First(&account).Error; err != nil {
		t.Fatal("account unavailable")
	}
	var attachments []entity.AIEntryAttachment
	if err := db.Where("session_id = ? AND owner_id = ? AND ocr_status = 'completed'", id, process.OwnerID).Order("id").Find(&attachments).Error; err != nil || len(attachments) == 0 {
		t.Fatal("completed OCR unavailable")
	}
	var wallets []entity.Wallet
	if err := db.Where("owner_id = ?", process.OwnerID).Order("created_at, id").Find(&wallets).Error; err != nil || len(wallets) == 0 {
		t.Fatal("wallet catalog unavailable")
	}
	var categories []entity.Category
	if err := db.Where("owner_id = ? OR is_system = true", process.OwnerID).Order("kind, parent_id NULLS FIRST, name").Find(&categories).Error; err != nil {
		t.Fatal("category catalog unavailable")
	}
	for i := range categories {
		if err := db.Table("category_wallets").Joins("JOIN wallets ON wallets.id = category_wallets.wallet_id").Where("category_id = ? AND wallets.owner_id = ?", categories[i].ID, process.OwnerID).Pluck("wallet_id", &categories[i].WalletIDs).Error; err != nil {
			t.Fatal("category scope unavailable")
		}
	}
	var texts []string
	for _, attachment := range attachments {
		texts = append(texts, attachment.OCRText)
	}
	out, err := c.Extract(context.Background(), Input{Text: strings.Join(texts, "\n\n"), Timezone: account.Timezone, Now: process.CreatedAt, Wallets: wallets, Categories: categories})
	if err != nil {
		t.Fatalf("stored OCR extraction failed: %v", err)
	}
	if len(out.Drafts) == 0 {
		t.Fatal("stored OCR still produced zero drafts; reply withheld")
	}
	for _, d := range out.Drafts {
		owned := false
		for _, w := range wallets {
			if d.WalletID == w.ID {
				owned = true
			}
		}
		if !owned {
			t.Fatal("draft lacks owner-wallet prefill")
		}
	}
	t.Logf("stored OCR replay: %d attachments -> %d drafts with owner-wallet prefill; no ledger writes", len(attachments), len(out.Drafts))
}
