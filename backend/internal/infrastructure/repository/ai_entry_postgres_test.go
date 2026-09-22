package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	port "github.com/mypocket/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"sync"
	"testing"
	"time"
)

func TestAIEntryPersistsReviewAndApprovesExactlyOnce(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	w := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "AI test", Type: "basic", Currency: "VND"}
	u := entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner}
	if err = db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("id = ?", w.ID).Delete(&entity.Wallet{})
		db.Where("id = ?", owner).Delete(&entity.User{})
	})
	if err = db.Create(&w).Error; err != nil {
		t.Fatal(err)
	}
	r := NewAIEntryPostgresRepository(db)
	ctx := context.Background()
	s, err := r.CreateSession(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.Session(ctx, "other", s.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("owner leak: %v", err)
	}
	token, started, err := r.BeginMessage(ctx, owner, s.ID, "request-1", "hash-1", "ăn 35k")
	if err != nil || !started {
		t.Fatalf("begin: %v %v", started, err)
	}
	if _, _, err = r.BeginMessage(ctx, owner, s.ID, "request-2", "hash-2", "another"); !errors.Is(err, port.ErrAIBusy) {
		t.Fatalf("concurrent message: %v", err)
	}
	draft := entity.AIEntryDraft{Type: "expense", Amount: 35000, WalletID: w.ID, OccurredAt: "2026-09-21T08:00:00+07:00", IncludedInReports: true}
	out := entity.AIExtractOutput{Reply: "review", Drafts: []entity.AIExtractDraft{{AIEntryDraft: draft}, {AIEntryDraft: draft}}}
	if err = r.FinishMessage(ctx, owner, s.ID, token, out); err != nil {
		t.Fatal(err)
	}
	attachment := entity.AIEntryAttachment{ID: uuid.NewString(), OwnerID: owner, ProcessID: s.ID, ObjectKey: "mypocket/test/private.pdf", Filename: "receipt.pdf", MIMEType: "application/pdf", SizeBytes: 64, SHA256: "test-hash", OCRStatus: "completed", DeleteAfter: time.Now().Add(24 * time.Hour)}
	if err = r.CreateAttachments(ctx, []entity.AIEntryAttachment{attachment}); err != nil {
		t.Fatal(err)
	}
	if _, err = r.AttachmentForTransaction(ctx, owner, "not-yet-approved", attachment.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("unlinked evidence must not be downloadable: %v", err)
	}
	if err = r.FinishMessage(ctx, owner, s.ID, token, out); !errors.Is(err, port.ErrAIConflict) {
		t.Fatalf("completion replay: %v", err)
	}
	if _, started, err = r.BeginMessage(ctx, owner, s.ID, "request-1", "hash-1", "ăn 35k"); err != nil || started {
		t.Fatalf("message duplicated: %v %v", started, err)
	}
	if _, _, err = r.BeginMessage(ctx, owner, s.ID, "request-1", "changed", "other"); !errors.Is(err, port.ErrAIConflict) {
		t.Fatalf("changed key: %v", err)
	}
	s, err = r.Session(ctx, owner, s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Proposals) != 2 || len(s.Messages) != 0 || s.Processing {
		t.Fatalf("wrong one-shot process state: %+v", s)
	}
	var count int64
	db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count)
	if count != 0 {
		t.Fatal("proposal changed ledger")
	}
	p := s.Proposals[0]
	draft.Amount = 40000
	pEdited, err := r.EditProposal(ctx, owner, p.ID, p.Version, draft)
	if err != nil {
		t.Fatal(err)
	}
	// A non-system ownerless template is not in the user's visible catalog.
	template := entity.Category{ID: uuid.NewString(), Name: "hidden template", Kind: "expense", IconKey: "tag"}
	if err = db.Create(&template).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("owner_id = ?", owner).Delete(&entity.Transaction{})
		db.Where("id = ?", template.ID).Delete(&entity.Category{})
	})
	badDraft := draft
	badDraft.CategoryID = &template.ID
	blocked, err := r.EditProposal(ctx, owner, s.Proposals[1].ID, 1, badDraft)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.DecideProposal(ctx, owner, blocked.ID, blocked.Version, true); !errors.Is(err, port.ErrAIInvalid) {
		t.Fatalf("invisible template approved: %v", err)
	}
	if _, err = r.DecideProposal(ctx, owner, p.ID, p.Version, true); !errors.Is(err, port.ErrAIConflict) {
		t.Fatalf("stale approve: %v", err)
	}
	if _, err = r.DecideProposal(ctx, "other", p.ID, pEdited.Version, true); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("approve owner leak: %v", err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	ids := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, e := r.DecideProposal(ctx, owner, p.ID, pEdited.Version, true)
			if e != nil {
				errs <- e
				return
			}
			if len(result.AttachmentIDs) != 1 || result.AttachmentIDs[0] != attachment.ID {
				errs <- errors.New("approved result omitted its linked receipt ID")
				return
			}
			ids <- *result.TransactionID
		}()
	}
	wg.Wait()
	close(errs)
	close(ids)
	for e := range errs {
		t.Error(e)
	}
	var savedID string
	for id := range ids {
		if savedID != "" && savedID != id {
			t.Fatal("duplicate transaction IDs")
		}
		savedID = id
	}
	var rows []entity.Transaction
	db.Where("owner_id = ?", owner).Find(&rows)
	if len(rows) != 1 || rows[0].Amount != 40000 || !rows[0].OccurredAt.Equal(time.Date(2026, 9, 21, 1, 0, 0, 0, time.UTC)) {
		t.Fatalf("wrong ledger: %+v", rows)
	}
	var links []entity.AIEntryAttachmentLink
	if err = db.Where("attachment_id = ?", attachment.ID).Find(&links).Error; err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].TransactionID != savedID {
		t.Fatalf("attachment should link only once to the approved transaction: %+v", links)
	}
	if _, err = r.AttachmentForTransaction(ctx, owner, savedID, attachment.ID); err != nil {
		t.Fatalf("owner should download linked receipt: %v", err)
	}
	if _, err = r.AttachmentForTransaction(ctx, "other", savedID, attachment.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cross-owner attachment download must fail: %v", err)
	}
	cleanupProcess := uuid.NewString()
	if err = r.CreateProcess(ctx, owner, cleanupProcess); err != nil {
		t.Fatal(err)
	}
	expired := entity.AIEntryAttachment{ID: uuid.NewString(), OwnerID: owner, ProcessID: cleanupProcess, ObjectKey: "mypocket/test/expired.pdf", Filename: "expired.pdf", MIMEType: "application/pdf", SizeBytes: 64, SHA256: "expired-hash", OCRStatus: "failed", DeleteAfter: time.Now().Add(-time.Hour)}
	if err = r.CreateAttachments(ctx, []entity.AIEntryAttachment{expired}); err != nil {
		t.Fatal(err)
	}
	claimed, err := r.ClaimExpiredAttachments(ctx, time.Now(), time.Now().Add(-10*time.Minute), 10)
	if err != nil || len(claimed) != 1 || claimed[0].ID != expired.ID {
		t.Fatalf("cleanup must claim only expired unlinked evidence: %+v %v", claimed, err)
	}
	if err = r.SetAttachmentDeleteStatus(ctx, expired.ID, "delete_failed"); err != nil {
		t.Fatal(err)
	}
	if _, err = r.DecideProposal(ctx, owner, s.Proposals[1].ID, blocked.Version, false); err != nil {
		t.Fatal(err)
	}
	if _, err = r.DecideProposal(ctx, owner, s.Proposals[1].ID, 2, true); !errors.Is(err, port.ErrAIConflict) {
		t.Fatalf("rejected approved: %v", err)
	}
	db.Where("id = ?", savedID).Delete(&entity.Transaction{})
	if _, err = r.DecideProposal(ctx, owner, p.ID, pEdited.Version, true); err != nil {
		t.Fatal(err)
	}
	db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count)
	if count != 0 {
		t.Fatal("retry resurrected deleted ledger")
	}
	// Opening a new session cannot bypass the per-owner provider budget.
	for i := 0; i < 19; i++ {
		if err = db.Create(&entity.AIEntryRequest{SessionID: s.ID, RequestID: uuid.NewString(), Hash: "budget", Token: uuid.NewString()}).Error; err != nil {
			t.Fatal(err)
		}
	}
	next, err := r.CreateSession(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = r.BeginMessage(ctx, owner, next.ID, "over-budget", "hash", "hello"); !errors.Is(err, port.ErrAIRateLimited) {
		t.Fatalf("owner budget bypass: %v", err)
	}
}

func TestAIEntryOCRTextMapsByStableAttachmentID(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner := uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("id = ?", owner).Delete(&entity.User{}) })
	process := uuid.NewString()
	r := NewAIEntryPostgresRepository(db)
	if err = r.CreateProcess(ctx, owner, process); err != nil {
		t.Fatal(err)
	}
	created := time.Now().Add(-time.Hour)
	files := []entity.AIEntryAttachment{
		{ID: "a-attachment", OwnerID: owner, ProcessID: process, ObjectKey: "private/a", Filename: "a.pdf", MIMEType: "application/pdf", SizeBytes: 32, SHA256: "a", OCRStatus: "pending", DeleteAfter: time.Now().Add(24 * time.Hour), CreatedAt: created.Add(time.Minute)},
		{ID: "z-attachment", OwnerID: owner, ProcessID: process, ObjectKey: "private/z", Filename: "z.pdf", MIMEType: "application/pdf", SizeBytes: 32, SHA256: "z", OCRStatus: "pending", DeleteAfter: time.Now().Add(24 * time.Hour), CreatedAt: created},
	}
	if err = r.CreateAttachments(ctx, files); err != nil {
		t.Fatal(err)
	}
	if err = r.UpdateAttachmentOCR(ctx, owner, process, []string{"OCR for a", "OCR for z"}, true); err != nil {
		t.Fatal(err)
	}
	var gotA, gotZ entity.AIEntryAttachment
	if err = db.Where("id = ?", "a-attachment").First(&gotA).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Where("id = ?", "z-attachment").First(&gotZ).Error; err != nil {
		t.Fatal(err)
	}
	if gotA.OCRText != "OCR for a" || gotZ.OCRText != "OCR for z" {
		t.Fatalf("OCR text mapped by timestamp instead of stable file order: a=%q z=%q", gotA.OCRText, gotZ.OCRText)
	}
}
