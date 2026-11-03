package httpapi

import (
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestFilterTransactionsByJar(t *testing.T) {
	jar, other, transfer := "jar-1", "jar-2", "t-1"
	rows := []entity.Transaction{
		{ID: "a", Type: entity.TransactionTypeExpense, JarID: &jar},
		{ID: "b", Type: entity.TransactionTypeExpense, JarID: &other},
		{ID: "c", Type: entity.TransactionTypeExpense},
		{ID: "d", Type: entity.TransactionTypeIncome},
		{ID: "e", Type: entity.TransactionTypeExpense, TransferID: &transfer},
	}
	ids := func(list []entity.Transaction) string {
		out := ""
		for _, row := range list {
			out += row.ID
		}
		return out
	}
	if got := ids(filterTransactionsByJar(rows, "")); got != "abcde" {
		t.Fatalf("no filter must keep every row, got %q", got)
	}
	if got := ids(filterTransactionsByJar(rows, "jar-1")); got != "a" {
		t.Fatalf("jar filter, got %q", got)
	}
	if got := ids(filterTransactionsByJar(rows, "unassigned")); got != "c" {
		t.Fatalf("unassigned selects ordinary expenses without a jar (no income, no transfer legs), got %q", got)
	}
}
