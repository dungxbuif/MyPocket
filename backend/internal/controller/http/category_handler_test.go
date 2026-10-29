package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	categoryrepo "github.com/mypocket/backend/internal/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type categoryStatusRepositoryStub struct {
	categoryrepo.CategoryRepository
	ownerID  string
	category entity.Category
}

func (s *categoryStatusRepositoryStub) SetActive(ownerID, id string, active bool) (*entity.Category, error) {
	if ownerID != s.ownerID || id != s.category.ID || s.category.IsSystem {
		return nil, gorm.ErrRecordNotFound
	}
	s.category.IsActive = active
	return &s.category, nil
}

func TestSetCategoryStatusReturnsPersistedActiveState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &categoryStatusRepositoryStub{
		ownerID:  "owner",
		category: entity.Category{ID: "personal", Name: "Cafe", Kind: "expense", IsActive: true},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner"); c.Next() })
	router.PATCH("/categories/:id/status", NewCategoryHandler(usecase.NewCategoryInteractor(repository)).SetCategoryStatus)

	request := httptest.NewRequest(http.MethodPatch, "/categories/personal/status", strings.NewReader(`{"is_active":false}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload struct {
		Data entity.Category `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.IsActive {
		t.Fatalf("response active = true, want false: %s", response.Body.String())
	}
	if repository.category.IsActive {
		t.Fatal("repository did not persist archived state")
	}
}

func TestSetCategoryStatusRequiresExplicitState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner"); c.Next() })
	router.PATCH("/categories/:id/status", NewCategoryHandler(usecase.NewCategoryInteractor(&categoryStatusRepositoryStub{})).SetCategoryStatus)
	request := httptest.NewRequest(http.MethodPatch, "/categories/personal/status", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", response.Code, response.Body.String())
	}
}

func TestCategoryArchiveHTTPFlowWithPostgresAndTransactionAPI(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	ownerID := uuid.NewString()
	user := entity.User{ID: ownerID, Email: ownerID + "@category-archive.test", GoogleSubject: ownerID, Timezone: "Asia/Ho_Chi_Minh"}
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: ownerID, Name: "archive flow wallet", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND}
	category := entity.Category{ID: uuid.NewString(), OwnerID: &ownerID, Kind: entity.TransactionTypeExpense, Name: "archive flow category", IconKey: "tag", IsActive: true}
	unscopedCategory := entity.Category{ID: uuid.NewString(), OwnerID: &ownerID, Kind: entity.TransactionTypeExpense, Name: "archive flow unscoped category", IconKey: "tag", IsActive: true}
	categoryWallet := entity.CategoryWallet{CategoryID: category.ID, WalletID: wallet.ID}
	historical := entity.Transaction{ID: uuid.NewString(), OwnerID: ownerID, WalletID: wallet.ID, CategoryID: &category.ID, Type: entity.TransactionTypeExpense, Amount: 1200, OccurredAt: time.Now().UTC()}
	startDate, err := entity.ParseCalendarDate(time.Now().Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	endDate, err := entity.ParseCalendarDate(time.Now().AddDate(0, 0, 30).Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	budget := entity.Budget{ID: uuid.NewString(), OwnerID: ownerID, Name: "archive flow budget", LimitAmount: 10000, WalletID: &wallet.ID, CategoryID: &category.ID, StartDate: startDate, EndDate: endDate}
	dueAt := time.Now().UTC().Add(-time.Minute)
	schedule := entity.RecurringSchedule{ID: uuid.NewString(), OwnerID: ownerID, Name: "archive flow schedule", WalletID: wallet.ID, CategoryID: &category.ID, Type: entity.TransactionTypeExpense, Amount: 300, Frequency: "daily", Interval: 1, NextRunAt: dueAt, AnchorDay: dueAt.Day(), AnchorMonth: int(dueAt.Month()), Active: true}
	t.Cleanup(func() {
		db.Where("schedule_id = ?", schedule.ID).Delete(&entity.RecurringOccurrence{})
		db.Where("id = ?", schedule.ID).Delete(&entity.RecurringSchedule{})
		db.Where("id = ?", budget.ID).Delete(&entity.Budget{})
		db.Where("owner_id = ?", ownerID).Delete(&entity.Transaction{})
		db.Where("category_id = ?", category.ID).Delete(&entity.CategoryWallet{})
		db.Where("id = ?", category.ID).Delete(&entity.Category{})
		db.Where("id = ?", unscopedCategory.ID).Delete(&entity.Category{})
		db.Where("id = ?", wallet.ID).Delete(&entity.Wallet{})
		db.Where("id = ?", ownerID).Delete(&entity.User{})
	})
	for _, item := range []any{&user, &wallet, &category, &unscopedCategory, &historical, &budget, &schedule} {
		if err := db.Create(item).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&categoryWallet).Error; err != nil {
		t.Fatal(err)
	}

	categories := postgresrepo.NewCategoryPostgresRepository(db)
	transactions := postgresrepo.NewTransactionPostgresRepository(db)
	wallets := postgresrepo.NewWalletPostgresRepository(db)
	categoryHandler := NewCategoryHandler(usecase.NewCategoryInteractor(categories))
	transactionHandler := NewTransactionHandler(transactions, wallets, categories)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(contextUserIDKey, ownerID); c.Next() })
	router.PATCH("/categories/:id/status", categoryHandler.SetCategoryStatus)
	router.POST("/transactions", transactionHandler.CreateTransaction)

	request := httptest.NewRequest(http.MethodPatch, "/categories/"+category.ID+"/status", strings.NewReader(`{"is_active":false}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"is_active":false`) {
		t.Fatalf("archive status response = %d %s", response.Code, response.Body.String())
	}
	var archivePayload struct {
		Data struct {
			WalletIDs []string `json:"wallet_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &archivePayload); err != nil {
		t.Fatal(err)
	}
	if len(archivePayload.Data.WalletIDs) != 1 || archivePayload.Data.WalletIDs[0] != wallet.ID {
		t.Fatalf("archive response wallet_ids = %v, want owner-scoped wallet %s: %s", archivePayload.Data.WalletIDs, wallet.ID, response.Body.String())
	}

	visible, err := categories.ListVisible(ownerID)
	if err != nil {
		t.Fatal(err)
	}
	var foundArchived bool
	for _, item := range visible {
		if item.ID == category.ID {
			foundArchived = !item.IsActive
		}
	}
	if !foundArchived {
		t.Fatal("archived category must remain visible for historical resolution")
	}

	transactionBody := `{"wallet_id":"` + wallet.ID + `","category_id":"` + category.ID + `","type":"expense","amount":500,"occurred_at":"2026-10-08T09:00:00Z"}`
	request = httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(transactionBody))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"category_inactive"`) {
		t.Fatalf("archived category transaction response = %d %s", response.Code, response.Body.String())
	}

	var historyCount int64
	if err := db.Model(&entity.Transaction{}).Where("id = ? AND category_id = ?", historical.ID, category.ID).Count(&historyCount).Error; err != nil {
		t.Fatal(err)
	}
	if historyCount != 1 {
		t.Fatalf("archive changed historical transactions: count=%d", historyCount)
	}
	var persistedBudget entity.Budget
	if err := db.Where("id = ?", budget.ID).First(&persistedBudget).Error; err != nil || persistedBudget.CategoryID == nil || *persistedBudget.CategoryID != category.ID {
		t.Fatalf("archive changed the existing budget category reference: budget=%+v err=%v", persistedBudget, err)
	}
	var persistedSchedule entity.RecurringSchedule
	if err := db.Where("id = ?", schedule.ID).First(&persistedSchedule).Error; err != nil || persistedSchedule.CategoryID == nil || *persistedSchedule.CategoryID != category.ID {
		t.Fatalf("archive changed the recurring schedule category reference: schedule=%+v err=%v", persistedSchedule, err)
	}
	createdBySchedule, err := postgresrepo.NewRecurringPostgresRepository(db).RunDue(ownerID, time.Now().UTC())
	if err != nil || createdBySchedule != 1 {
		t.Fatalf("existing recurring schedule should continue with its saved category while inactive: created=%d err=%v", createdBySchedule, err)
	}

	request = httptest.NewRequest(http.MethodPatch, "/categories/"+category.ID+"/status", strings.NewReader(`{"is_active":true}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"is_active":true`) {
		t.Fatalf("restore status response = %d %s", response.Code, response.Body.String())
	}
	var restorePayload struct {
		Data struct {
			WalletIDs []string `json:"wallet_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &restorePayload); err != nil {
		t.Fatal(err)
	}
	if len(restorePayload.Data.WalletIDs) != 1 || restorePayload.Data.WalletIDs[0] != wallet.ID {
		t.Fatalf("restore response wallet_ids = %v, want owner-scoped wallet %s: %s", restorePayload.Data.WalletIDs, wallet.ID, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(transactionBody))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("restored category transaction response = %d %s", response.Code, response.Body.String())
	}
	if err := db.Model(&entity.Transaction{}).Where("owner_id = ? AND category_id = ?", ownerID, category.ID).Count(&historyCount).Error; err != nil {
		t.Fatal(err)
	}
	if historyCount != 3 {
		t.Fatalf("expected historical, recurring and newly created transactions, count=%d", historyCount)
	}

	request = httptest.NewRequest(http.MethodPatch, "/categories/"+unscopedCategory.ID+"/status", strings.NewReader(`{"is_active":false}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	var unscopedPayload struct {
		Data struct {
			WalletIDs []string `json:"wallet_ids"`
		} `json:"data"`
	}
	if response.Code != http.StatusOK {
		t.Fatalf("unscoped archive response = %d %s", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &unscopedPayload); err != nil {
		t.Fatal(err)
	}
	if unscopedPayload.Data.WalletIDs == nil || len(unscopedPayload.Data.WalletIDs) != 0 {
		t.Fatalf("unscoped archive response wallet_ids = %v, want empty array: %s", unscopedPayload.Data.WalletIDs, response.Body.String())
	}
}
