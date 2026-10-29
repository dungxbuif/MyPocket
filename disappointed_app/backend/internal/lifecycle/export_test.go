package lifecycle

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestCSVExportIsStableAndMachineReadable(t *testing.T) {
	snapshot := Snapshot{Rows: []ExportRow{{Dataset: "transactions", ID: "tx-1", Type: "expense", AmountVND: 12000, OccurredAt: time.Date(2026, 9, 11, 3, 0, 0, 0, time.UTC), Note: "cà phê"}}}
	var first, second bytes.Buffer
	if err := (CSVExportFormatter{}).Write(context.Background(), snapshot, &first); err != nil {
		t.Fatal(err)
	}
	if err := (CSVExportFormatter{}).Write(context.Background(), snapshot, &second); err != nil {
		t.Fatal(err)
	}
	if first.String() != second.String() {
		t.Fatal("export is not deterministic")
	}
	if !strings.Contains(first.String(), "amount_vnd") || !strings.Contains(first.String(), "2026-09-11T03:00:00Z") {
		t.Fatalf("unexpected CSV: %s", first.String())
	}
}
