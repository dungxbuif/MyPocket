package repository

import "testing"

func TestMergeJarAssignmentEditKeepsServerFactsAndDropsUnknownRows(t *testing.T) {
	original := []any{
		map[string]any{"transaction_id": "a", "jar_id": "j1", "status": "proposed", "expected_version": float64(4), "note": "Phở"},
		map[string]any{"transaction_id": "b", "jar_id": "j1", "status": "proposed", "expected_version": float64(2)},
	}
	edited := []any{
		map[string]any{"transaction_id": "a", "jar_id": "j2", "expected_version": float64(99), "status": "applied"}, // facts are not client-owned
		map[string]any{"transaction_id": "zzz", "jar_id": "j1"},                                                     // unknown row dropped; b removed by omission
		map[string]any{"transaction_id": "a", "jar_id": "j1"},                                                       // duplicate dropped
	}
	merged := mergeJarAssignmentEdit(original, edited)
	if len(merged) != 1 {
		t.Fatalf("merged=%v", merged)
	}
	row := merged[0].(map[string]any)
	if row["jar_id"] != "j2" || row["expected_version"] != float64(4) || row["status"] != "proposed" || row["note"] != "Phở" || row["edited"] != true {
		t.Fatalf("row=%v", row)
	}
}
