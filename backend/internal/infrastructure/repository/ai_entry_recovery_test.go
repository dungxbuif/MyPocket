package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type partialOCRExtractor struct{}

func (partialOCRExtractor) Configured() bool                     { return true }
func (partialOCRExtractor) OCRConfigured() bool                  { return true }
func (partialOCRExtractor) ValidateFiles([]entity.AIImage) error { return nil }
func (partialOCRExtractor) Extract(context.Context, entity.AIExtractInput) (entity.AIExtractOutput, error) {
	return entity.AIExtractOutput{SourceText: "first receipt amount 42000", AttachmentTexts: []string{"first receipt amount 42000"}}, errors.New("second image OCR failed")
}

type recoveryAttachmentStorage struct{}

func (recoveryAttachmentStorage) Key(_, process, attachment, filename string) (string, error) {
	return "test-recovery/" + process + "/" + attachment + "/" + filename, nil
}
func (recoveryAttachmentStorage) Put(context.Context, string, string, []byte) error { return nil }
func (recoveryAttachmentStorage) Delete(context.Context, string) error              { return nil }
func (recoveryAttachmentStorage) SignedGet(context.Context, string) (string, error) { return "", nil }

func TestAIEntryPartialOCRPersistsSuccessfulAndFailedAttachments(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("id = ?", owner).Delete(&entity.User{}) })
	entries := NewAIEntryPostgresRepository(db)
	service := &usecase.AIEntryService{Entries: entries, Wallets: NewWalletPostgresRepository(db), Categories: NewCategoryPostgresRepository(db), Extractor: partialOCRExtractor{}, Storage: recoveryAttachmentStorage{}}
	process := uuid.NewString()
	encoded := base64.StdEncoding.EncodeToString([]byte("%PDF-1.7 receipt %%EOF"))
	_, err = service.Process(context.Background(), owner, usecase.AIEntryMessageInput{
		RequestID: process, Text: "review receipts", Timezone: "UTC",
		Images: []entity.AIImage{{Name: "first.pdf", MIMEType: "application/pdf", Base64: encoded}, {Name: "second.pdf", MIMEType: "application/pdf", Base64: encoded}},
	})
	if !errors.Is(err, usecase.ErrAIProvider) {
		t.Fatalf("expected failed extraction, got %v", err)
	}
	var attachments []entity.AIEntryAttachment
	if err = db.Where("owner_id = ? AND session_id = ?", owner, process).Order("filename").Find(&attachments).Error; err != nil {
		t.Fatal(err)
	}
	if len(attachments) != 2 || attachments[0].Filename != "first.pdf" || attachments[0].OCRStatus != "completed" || attachments[0].OCRText != "first receipt amount 42000" {
		t.Fatalf("successful OCR lost filename alignment/evidence: %+v", attachments)
	}
	if attachments[1].Filename != "second.pdf" || attachments[1].OCRStatus != "failed" || attachments[1].OCRText != "" {
		t.Fatalf("unprocessed OCR persisted as completed: %+v", attachments[1])
	}
	session, err := entries.Session(context.Background(), owner, process)
	if err != nil || session.Processing || session.Error == "" || len(session.Proposals) != 0 {
		t.Fatalf("failed process must terminate without proposals: %+v err=%v", session, err)
	}
}

func TestAIEntryOCRSparseAlignmentKeepsEmptySlotsFailed(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner, process := uuid.NewString(), uuid.NewString()
	if err = db.Create(&entity.User{ID: owner, Email: owner + "@test.invalid", GoogleSubject: owner}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Where("id = ?", owner).Delete(&entity.User{}) })
	entries := NewAIEntryPostgresRepository(db)
	ctx := context.Background()
	if err = entries.CreateProcess(ctx, owner, process); err != nil {
		t.Fatal(err)
	}
	files := []entity.AIEntryAttachment{
		{ID: "a-" + uuid.NewString(), OwnerID: owner, ProcessID: process, ObjectKey: "test/a", Filename: "failed.pdf", MIMEType: "application/pdf", SizeBytes: 32, SHA256: "a", OCRStatus: "pending", DeleteAfter: time.Now().Add(time.Hour)},
		{ID: "z-" + uuid.NewString(), OwnerID: owner, ProcessID: process, ObjectKey: "test/z", Filename: "completed.pdf", MIMEType: "application/pdf", SizeBytes: 32, SHA256: "z", OCRStatus: "pending", DeleteAfter: time.Now().Add(time.Hour)},
	}
	if err = entries.CreateAttachments(ctx, files); err != nil {
		t.Fatal(err)
	}
	if err = entries.UpdateAttachmentOCR(ctx, owner, process, []string{"", "completed evidence"}, false); err != nil {
		t.Fatal(err)
	}
	var actual []entity.AIEntryAttachment
	if err = db.Where("owner_id = ? AND session_id = ?", owner, process).Order("id").Find(&actual).Error; err != nil {
		t.Fatal(err)
	}
	if len(actual) != 2 || actual[0].OCRStatus != "failed" || actual[0].OCRText != "" || actual[1].OCRStatus != "completed" || actual[1].OCRText != "completed evidence" {
		t.Fatalf("sparse aligned OCR status/text corrupted: %+v", actual)
	}
}
