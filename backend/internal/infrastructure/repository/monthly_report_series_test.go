package repository

import (
	"context"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

func TestMonthlyWeeklySeriesStopsAtPartialFinalDay(t *testing.T) {
	db := openReportsPostgresTestDB(t)
	owner := insertReportsTestUser(t, db, "Asia/Ho_Chi_Minh")
	wallet := budgetScopeWallet(t, db, owner)
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(2026, 11, 1, 12, 0, 0, 0, loc)
	second := first.AddDate(0, 0, 1)
	budgetScopeExpense(t, db, owner, wallet, nil, 100, first)
	budgetScopeExpense(t, db, owner, wallet, nil, 200, second)
	repo := NewMonthlyReportPostgresRepository(db)
	for _, asOf := range []time.Time{first.Add(time.Hour), second.Add(time.Hour), time.Date(2026, 11, 3, 0, 0, 0, 0, loc)} {
		scope, err := repo.resolveScope(context.Background(), db, owner, entity.MonthlyReportScope{Month: "2026-11", WalletID: wallet, Granularity: "week"}, asOf)
		if err != nil {
			t.Fatal(err)
		}
		points, err := repo.series(context.Background(), db, owner, scope)
		if err != nil {
			t.Fatal(err)
		}
		wantCount, wantExpense := 2, int64(300)
		if asOf.Before(second) {
			wantCount, wantExpense = 1, 100
		}
		if len(points) != wantCount {
			t.Fatalf("asOf=%v buckets=%d want=%d", asOf, len(points), wantCount)
		}
		var total int64
		for i, p := range points {
			total += p.Expense
			if !p.EndExclusive.After(p.From) {
				t.Fatalf("empty bucket: %+v", p)
			}
			if i > 0 && !p.From.Equal(points[i-1].EndExclusive) {
				t.Fatalf("buckets not contiguous: %+v", points)
			}
		}
		if total != wantExpense || !points[len(points)-1].EndExclusive.Equal(asOf) {
			t.Fatalf("asOf=%v total=%d points=%+v", asOf, total, points)
		}
	}
}
