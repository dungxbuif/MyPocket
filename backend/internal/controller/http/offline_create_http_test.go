package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	dbprovider "github.com/mypocket/backend/internal/infrastructure/db"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	"gorm.io/gorm"
)

type offlineHTTPFixture struct {
	db                                   *gorm.DB
	owner, wallet, destination, category string
}

func newOfflineHTTPFixture(t *testing.T) offlineHTTPFixture {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL")
	}
	database, err := dbprovider.NewPostgres(dsn)
	if err != nil {
		t.Fatal("local DB connection failed")
	}
	f := offlineHTTPFixture{db: database, owner: uuid.NewString(), wallet: uuid.NewString(), destination: uuid.NewString(), category: uuid.NewString()}
	t.Cleanup(func() {
		for _, model := range []any{&entity.Transaction{}, &entity.TransactionMutation{}, &entity.TravelEvent{}} {
			if err := database.Where("owner_id = ?", f.owner).Delete(model).Error; err != nil {
				t.Error(err)
			}
		}
		if err := database.Where("category_id = ?", f.category).Delete(&entity.CategoryWallet{}).Error; err != nil {
			t.Error(err)
		}
		if err := database.Where("id = ? AND owner_id = ?", f.category, f.owner).Delete(&entity.Category{}).Error; err != nil {
			t.Error(err)
		}
		if err := database.Where("owner_id = ?", f.owner).Delete(&entity.Wallet{}).Error; err != nil {
			t.Error(err)
		}
		if err := database.Where("id = ?", f.owner).Delete(&entity.User{}).Error; err != nil {
			t.Error(err)
		}
		connection, _ := database.DB()
		if connection != nil {
			_ = connection.Close()
		}
	})
	for _, item := range []any{
		&entity.User{ID: f.owner, Email: f.owner + "@offline-http.test", GoogleSubject: f.owner, Timezone: "Asia/Ho_Chi_Minh"},
		&entity.Wallet{ID: f.wallet, OwnerID: f.owner, Name: "HTTP source", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND},
		&entity.Wallet{ID: f.destination, OwnerID: f.owner, Name: "HTTP destination", Type: entity.WalletTypeBasic, Currency: entity.WalletCurrencyVND},
		&entity.Category{ID: f.category, OwnerID: &f.owner, Name: "HTTP category", Kind: entity.TransactionTypeExpense, IconKey: "tag", IsActive: true},
		&entity.CategoryWallet{CategoryID: f.category, WalletID: f.wallet},
	} {
		if err := database.Create(item).Error; err != nil {
			t.Fatal(err)
		}
	}
	return f
}

func (f offlineHTTPFixture) router(owner string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if owner != "" {
			c.Set(contextUserIDKey, owner)
		}
		c.Next()
	})
	h := NewTransactionHandler(postgresrepo.NewTransactionPostgresRepository(f.db), postgresrepo.NewWalletPostgresRepository(f.db), postgresrepo.NewCategoryPostgresRepository(f.db))
	h.Travel = postgresrepo.NewTravelPostgresRepository(f.db)
	r.POST("/transactions", h.CreateTransaction)
	r.POST("/transactions/transfer", h.CreateTransfer)
	return r
}

func (f offlineHTTPFixture) body(amount int) string {
	return fmt.Sprintf(`{"wallet_id":%q,"category_id":%q,"type":"expense","amount":%d,"included_in_reports":false,"owner_id":"synthetic-not-owner"}`, f.wallet, f.category, amount)
}

func offlineHTTPCall(r http.Handler, path, body string, keys ...string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	for _, key := range keys {
		request.Header.Add("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	return response
}

type offlineHTTPEnvelope struct {
	Data json.RawMessage `json:"data"`
	Meta struct {
		Mutation struct {
			Operation string `json:"operation"`
			Key       string `json:"idempotency_key"`
			Status    string `json:"status"`
			Replayed  bool   `json:"replayed"`
		} `json:"mutation"`
	} `json:"meta"`
}

func decodeOfflineHTTP(t *testing.T, response *httptest.ResponseRecorder) offlineHTTPEnvelope {
	t.Helper()
	var envelope offlineHTTPEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope
}

func TestOfflineCreateHTTPReplaysBeforeCategoryAndWalletRevalidation(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	r := f.router(f.owner)
	key := uuid.NewString()
	body := f.body(12500)
	first := offlineHTTPCall(r, "/transactions", body, key)
	if first.Code != 201 {
		t.Fatalf("first status%d body%s", first.Code, first.Body.String())
	}
	initial := decodeOfflineHTTP(t, first)
	if initial.Meta.Mutation.Key != key || initial.Meta.Mutation.Operation != "transaction.create" || initial.Meta.Mutation.Status != "applied" || initial.Meta.Mutation.Replayed {
		t.Fatal("missing first-create mutation metadata")
	}
	var firstRow entity.Transaction
	if err := json.Unmarshal(initial.Data, &firstRow); err != nil {
		t.Fatal(err)
	}
	if firstRow.ID == "" || firstRow.OwnerID != f.owner {
		t.Fatal("request body selected financial owner")
	}
	if err := f.db.Model(&entity.Category{}).Where("id = ?", f.category).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	// JSON whitespace does not alter the typed request identity.
	retry := offlineHTTPCall(r, "/transactions", "  "+body+"\n", key)
	if retry.Code != 200 {
		t.Fatalf("archived category replay status%d body%s", retry.Code, retry.Body.String())
	}
	replayed := decodeOfflineHTTP(t, retry)
	var retryRow entity.Transaction
	if err := json.Unmarshal(replayed.Data, &retryRow); err != nil {
		t.Fatal(err)
	}
	if !replayed.Meta.Mutation.Replayed || retryRow.ID != firstRow.ID || !retryRow.OccurredAt.Equal(firstRow.OccurredAt) || retryRow.IncludedInReports {
		t.Fatal("retry changed original transaction/date/report flag")
	}
	changed := offlineHTTPCall(r, "/transactions", f.body(12501), key)
	if changed.Code != 409 || !strings.Contains(changed.Body.String(), "IDEMPOTENCY_KEY_REUSED") {
		t.Fatalf("changed body status%d body%s", changed.Code, changed.Body.String())
	}
	if err := f.db.Where("id = ? AND owner_id = ?", f.wallet, f.owner).Delete(&entity.Wallet{}).Error; err != nil {
		t.Fatal(err)
	}
	retry = offlineHTTPCall(r, "/transactions", body, key)
	if retry.Code != 200 || !decodeOfflineHTTP(t, retry).Meta.Mutation.Replayed {
		t.Fatalf("deleted wallet replay failed:%d", retry.Code)
	}
	var count int64
	if err := f.db.Model(&entity.Transaction{}).Where("owner_id = ?", f.owner).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("receipt replay resurrected deleted financial data")
	}
}

func TestOfflineTransferHTTPConcurrentReplayKeepsExactlyOneExcludedPair(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	r := f.router(f.owner)
	key := uuid.NewString()
	body := fmt.Sprintf(`{"source_wallet_id":%q,"destination_wallet_id":%q,"amount":40000,"occurred_at":"2026-09-01T03:00:00Z"}`, f.wallet, f.destination)
	responses := make(chan *httptest.ResponseRecorder, 8)
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() { defer group.Done(); responses <- offlineHTTPCall(r, "/transactions/transfer", body, key) }()
	}
	group.Wait()
	close(responses)
	applied := 0
	pairID := ""
	for response := range responses {
		if response.Code != 200 && response.Code != 201 {
			t.Fatalf("pair status%d body%s", response.Code, response.Body.String())
		}
		envelope := decodeOfflineHTTP(t, response)
		var rows []entity.Transaction
		if err := json.Unmarshal(envelope.Data, &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 2 || rows[0].TransferID == nil || rows[1].TransferID == nil {
			t.Fatal("missing pair identity")
		}
		if envelope.Meta.Mutation.Operation != "transfer.create" || envelope.Meta.Mutation.Key != key || envelope.Meta.Mutation.Status != "applied" {
			t.Fatalf("transfer receipt metadata mismatch: %+v", envelope.Meta.Mutation)
		}
		if pairID == "" {
			pairID = *rows[0].TransferID
		}
		if *rows[0].TransferID != pairID || *rows[1].TransferID != pairID {
			t.Fatal("retry created different transfer pair")
		}
		for _, row := range rows {
			if row.IncludedInReports || row.JarID != nil || row.Amount != 40000 {
				t.Fatal("pair accounting changed")
			}
		}
		if !envelope.Meta.Mutation.Replayed {
			applied++
		}
	}
	if applied != 1 {
		t.Fatalf("want1 applied result, got%d", applied)
	}
	var count int64
	if err := f.db.Model(&entity.Transaction{}).Where("owner_id = ?", f.owner).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("want2 rows, got%d err%v", count, err)
	}
	if err := f.db.Model(&entity.TransactionMutation{}).Where("owner_id = ? AND operation = ? AND idempotency_key = ?", f.owner, "transfer.create", key).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("want1 transfer receipt, got%d err%v", count, err)
	}
}

func TestOfflineCreateHTTPRejectsExplicitInvalidKeysAndIsolatesOwners(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	r := f.router(f.owner)
	for _, keys := range [][]string{{""}, {"bad\nkey"}, {strings.Repeat("x", 129)}, {"one", "two"}} {
		response := offlineHTTPCall(r, "/transactions", f.body(100), keys...)
		if response.Code != 400 {
			t.Fatalf("invalid key must fail closed, got%d", response.Code)
		}
	}
	key := uuid.NewString()
	first := offlineHTTPCall(r, "/transactions", f.body(100), key)
	if first.Code != 201 {
		t.Fatal("first create failed")
	}
	var original entity.Transaction
	if err := json.Unmarshal(decodeOfflineHTTP(t, first).Data, &original); err != nil {
		t.Fatal(err)
	}
	other := newOfflineHTTPFixture(t)
	foreign := offlineHTTPCall(other.router(other.owner), "/transactions", f.body(100), key)
	if foreign.Code != 404 || strings.Contains(foreign.Body.String(), original.ID) {
		t.Fatalf("foreign replay leaked result:%d", foreign.Code)
	}
	owned := offlineHTTPCall(other.router(other.owner), "/transactions", other.body(100), key)
	if owned.Code != 201 || decodeOfflineHTTP(t, owned).Meta.Mutation.Replayed {
		t.Fatal("same key across owners must remain independent")
	}
	unauth := offlineHTTPCall(f.router(""), "/transactions", f.body(100), key)
	if unauth.Code != 401 {
		t.Fatal("receipt bypassed auth")
	}
}

func TestOfflineCreateHTTPLegacyShapeAndMissingCapabilityFailClosed(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	legacy := offlineHTTPCall(f.router(f.owner), "/transactions", f.body(100))
	if legacy.Code != http.StatusCreated {
		t.Fatalf("legacy online request failed:%d", legacy.Code)
	}
	var envelope struct {
		Data entity.Transaction `json:"data"`
		Meta json.RawMessage    `json:"meta"`
	}
	if err := json.Unmarshal(legacy.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.ID == "" || string(envelope.Meta) != "null" {
		t.Fatal("legacy response shape changed")
	}
	stub := &transactionRepositoryStub{}
	router := transactionTestRouter(NewTransactionHandler(stub, &transactionWalletRepositoryStub{}, &transactionCategoryRepositoryStub{}))
	keyed := offlineHTTPCall(router, "/transactions", `{"amount":100}`, uuid.NewString())
	if keyed.Code != http.StatusServiceUnavailable || stub.created != nil {
		t.Fatal("keyed request fell back to an unprotected financial write")
	}
}
