package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	postgresrepo "github.com/mypocket/backend/internal/infrastructure/repository"
	transactionrepo "github.com/mypocket/backend/internal/repository"
)

func TestOfflineCreateTravelContextFreezesIntentInsteadOfCurrentActiveTrip(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	travel := postgresrepo.NewTravelPostgresRepository(f.db)
	firstTrip := &entity.TravelEvent{ID: uuid.NewString(), OwnerID: f.owner, Name: "Original trip", Active: true}
	secondTrip := &entity.TravelEvent{ID: uuid.NewString(), OwnerID: f.owner, Name: "Later trip", Active: true}
	if err := travel.Create(f.owner, firstTrip); err != nil {
		t.Fatal(err)
	}
	// The app would prepare an explicit original trip or explicit no-trip here.
	if err := travel.Create(f.owner, secondTrip); err != nil {
		t.Fatal(err)
	}
	r := f.router(f.owner)
	base := fmt.Sprintf(`{"wallet_id":%q,"category_id":%q,"type":"expense","amount":12500,"occurred_at":"2026-09-01T03:00:00Z","included_in_reports":false`, f.wallet, f.category)
	for _, tc := range []struct {
		name, suffix string
		want         *string
	}{
		{"explicit none", `,"auto_link_travel":false}`, nil},
		{"explicit old trip", fmt.Sprintf(`,"auto_link_travel":false,"travel_event_id":%q}`, firstTrip.ID), &firstTrip.ID},
		{"legacy omitted", `}`, &secondTrip.ID},
		{"legacy null", `,"auto_link_travel":null}`, &secondTrip.ID},
		{"online explicit true", `,"auto_link_travel":true}`, &secondTrip.ID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key := uuid.NewString()
			response := offlineHTTPCall(r, "/transactions", base+tc.suffix, key)
			if response.Code != 201 {
				t.Fatalf("create status%d", response.Code)
			}
			var row entity.Transaction
			if err := json.Unmarshal(decodeOfflineHTTP(t, response).Data, &row); err != nil {
				t.Fatal(err)
			}
			if (row.TravelEventID == nil) != (tc.want == nil) || tc.want != nil && *row.TravelEventID != *tc.want {
				t.Fatal("delayed create inherited the wrong trip")
			}
			stored, err := postgresrepo.NewTransactionPostgresRepository(f.db).Find(f.owner, row.ID)
			if err != nil || !stored.OccurredAt.Equal(time.Date(2026, 9, 1, 3, 0, 0, 0, time.UTC)) {
				t.Fatal("fixed instant was not persisted")
			}
			if (stored.TravelEventID == nil) != (tc.want == nil) || tc.want != nil && *stored.TravelEventID != *tc.want {
				t.Fatal("response and persisted trip disagree")
			}
			if retry := offlineHTTPCall(r, "/transactions", base+tc.suffix, key); retry.Code != 200 || !decodeOfflineHTTP(t, retry).Meta.Mutation.Replayed {
				t.Fatal("stable policy replay failed")
			}
		})
	}
}

func TestOfflineCreateTravelPolicyConflictAndDeletedTripReplay(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	trip := &entity.TravelEvent{ID: uuid.NewString(), OwnerID: f.owner, Name: "Frozen", Active: false}
	if err := postgresrepo.NewTravelPostgresRepository(f.db).Create(f.owner, trip); err != nil {
		t.Fatal(err)
	}
	r := f.router(f.owner)
	key := uuid.NewString()
	body := fmt.Sprintf(`{"wallet_id":%q,"category_id":%q,"type":"expense","amount":100,"occurred_at":"2026-09-01T03:00:00Z","auto_link_travel":false,"travel_event_id":%q}`, f.wallet, f.category, trip.ID)
	created := offlineHTTPCall(r, "/transactions", body, key)
	if created.Code != 201 {
		t.Fatalf("create status%d", created.Code)
	}
	var first entity.Transaction
	if err := json.Unmarshal(decodeOfflineHTTP(t, created).Data, &first); err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(body, `"auto_link_travel":false`, `"auto_link_travel":true`, 1)
	if result := offlineHTTPCall(r, "/transactions", changed, key); result.Code != 409 || !strings.Contains(result.Body.String(), "IDEMPOTENCY_KEY_REUSED") {
		t.Fatal("changed policy did not conflict")
	}
	if err := f.db.Where("owner_id=? AND id=?", f.owner, trip.ID).Delete(&entity.TravelEvent{}).Error; err != nil {
		t.Fatal(err)
	}
	replayed := offlineHTTPCall(r, "/transactions", body, key)
	if replayed.Code != 200 || !decodeOfflineHTTP(t, replayed).Meta.Mutation.Replayed {
		t.Fatal("deleted-trip receipt replay must precede validation")
	}
	var result entity.Transaction
	if err := json.Unmarshal(decodeOfflineHTTP(t, replayed).Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.ID != first.ID || result.TravelEventID == nil || *result.TravelEventID != trip.ID {
		t.Fatal("receipt result changed original frozen intent")
	}
	if failed := offlineHTTPCall(r, "/transactions", body, uuid.NewString()); failed.Code != 404 {
		t.Fatal("new intent referenced deleted trip")
	}
	foreign := newOfflineHTTPFixture(t)
	foreignTrip := &entity.TravelEvent{ID: uuid.NewString(), OwnerID: foreign.owner, Name: "Foreign"}
	if err := postgresrepo.NewTravelPostgresRepository(foreign.db).Create(foreign.owner, foreignTrip); err != nil {
		t.Fatal(err)
	}
	if denied := offlineHTTPCall(r, "/transactions", strings.Replace(body, trip.ID, foreignTrip.ID, 1), uuid.NewString()); denied.Code != 404 {
		t.Fatal("foreign trip accepted")
	}
	var rows, receipts int64
	f.db.Model(&entity.Transaction{}).Where("owner_id=?", f.owner).Count(&rows)
	f.db.Model(&entity.TransactionMutation{}).Where("owner_id=?", f.owner).Count(&receipts)
	if rows != 1 || receipts != 1 {
		t.Fatal("rejected/replayed requests created additional financial writes")
	}
}

func TestOfflineCreateTravelAdditionPreservesPreexistingReceiptHash(t *testing.T) {
	f := newOfflineHTTPFixture(t)
	key := uuid.NewString()
	// Literal pre-change JSON field sequence. Do not derive this with the new DTO/hash helper.
	canonical := fmt.Sprintf(`{"wallet_id":%q,"category_id":%q,"jar_id":null,"type":"expense","amount":12500,"occurred_at":"2026-09-01T03:00:00Z","note":null,"included_in_reports":false,"travel_event_id":null}`, f.wallet, f.category)
	sum := sha256.Sum256([]byte(canonical))
	oldHash := hex.EncodeToString(sum[:])
	row := &entity.Transaction{ID: uuid.NewString(), OwnerID: f.owner, WalletID: f.wallet, CategoryID: &f.category, Type: "expense", Amount: 12500, OccurredAt: time.Date(2026, 9, 1, 3, 0, 0, 0, time.UTC), IncludedInReports: false}
	mutator := postgresrepo.NewTransactionPostgresRepository(f.db).(transactionrepo.TransactionCreateMutationRepository)
	if _, _, err := mutator.CreateIdempotent(row, key, oldHash); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{canonical, strings.TrimSuffix(canonical, "}") + `,"auto_link_travel":null}`} {
		response := offlineHTTPCall(f.router(f.owner), "/transactions", body, key)
		if response.Code != 200 || !decodeOfflineHTTP(t, response).Meta.Mutation.Replayed {
			t.Fatalf("legacy receipt hash changed:%d", response.Code)
		}
	}
}
