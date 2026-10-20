package migrations_test

import (
	"embed"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Keep this test next to the SQL so a release cannot silently omit a newly
// added migration from the image or reorder the forward history.
//
//go:embed *.sql
var migrationFiles embed.FS

var migrationNamePattern = regexp.MustCompile(`^(\d{6})_[a-z0-9_]+\.(up|down)\.sql$`)

func TestMigrationHistoryIsOrderedAndHasCurrentRolloutFiles(t *testing.T) {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	versions := make(map[int]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := migrationNamePattern.FindStringSubmatch(entry.Name())
		if match == nil || match[2] != "up" {
			continue
		}
		version, parseErr := strconv.Atoi(match[1])
		if parseErr != nil {
			t.Fatalf("parse migration %q: %v", entry.Name(), parseErr)
		}
		if previous, exists := versions[version]; exists {
			t.Fatalf("duplicate migration version %06d: %s and %s", version, previous, entry.Name())
		}
		versions[version] = entry.Name()
	}

	ordered := make([]int, 0, len(versions))
	for version := range versions {
		ordered = append(ordered, version)
	}
	sort.Ints(ordered)
	for index := 1; index < len(ordered); index++ {
		if ordered[index] <= ordered[index-1] {
			t.Fatalf("migration versions are not strictly increasing: %v", ordered)
		}
	}

	for _, version := range []int{35, 36, 37, 38, 39, 40, 41} {
		if _, ok := versions[version]; !ok {
			t.Fatalf("current release migration %06d is missing", version)
		}
	}
	if got := ordered[len(ordered)-1]; got != 41 {
		t.Fatalf("latest migration=%06d, want 000041", got)
	}

	for _, version := range []int{36, 37, 38, 39, 40, 41} {
		name := fmt.Sprintf("%06d_", version)
		foundDown := false
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), name) && strings.HasSuffix(entry.Name(), ".down.sql") {
				foundDown = true
				break
			}
		}
		if !foundDown {
			t.Fatalf("migration %06d has no disposable-fixture down file", version)
		}
	}
}

func TestAdvisorRuntimeMigrationAddsDeadlineAndDuplicateGuard(t *testing.T) {
	contents, err := migrationFiles.ReadFile("000041_advisor_runtime.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToUpper(string(contents))
	for _, required := range []string{
		"ADD COLUMN DEADLINE_AT TIMESTAMPTZ",
		"DEADLINE_AT = CREATED_AT + INTERVAL '240 SECONDS'",
		"ADVISOR_RUN_RUNTIME_REAPER_IDX",
		"HAVING COUNT(*) > 1",
		"ADVISOR_ASSISTANT_COMPLETION_UNIQUE",
	} {
		if !strings.Contains(sql, strings.ToUpper(required)) {
			t.Errorf("000041_advisor_runtime.up.sql missing runtime safety contract %q", required)
		}
	}
}

func TestCurrentRolloutMigrationsAreAdditiveAndOwnerSafe(t *testing.T) {
	for _, file := range []string{"000036_push_tokens.up.sql", "000037_audit_outbox.up.sql", "000039_push_deliveries.up.sql"} {
		contents, err := migrationFiles.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		sql := strings.ToUpper(string(contents))
		for _, forbidden := range []string{"DROP TABLE", "DROP DATABASE", "TRUNCATE", "ALTER TABLE \"USER\""} {
			if strings.Contains(sql, forbidden) {
				t.Fatalf("%s contains destructive or owner-schema rewrite %q", file, forbidden)
			}
		}
	}

	push, _ := migrationFiles.ReadFile("000036_push_tokens.up.sql")
	for _, required := range []string{
		"REFERENCES \"user\"(id) ON DELETE CASCADE",
		"UNIQUE INDEX push_tokens_owner_device_idx",
		"UNIQUE INDEX push_tokens_provider_address_idx",
		"CHECK (platform = 'ios')",
	} {
		if !strings.Contains(strings.ToUpper(string(push)), strings.ToUpper(required)) {
			t.Errorf("000036_push_tokens.up.sql missing safety contract %q", required)
		}
	}

	outbox, _ := migrationFiles.ReadFile("000037_audit_outbox.up.sql")
	for _, required := range []string{
		"UNIQUE(stream,event_id)",
		"WHERE delivered_at IS NULL",
		"CHECK (attempts >= 0)",
	} {
		if !strings.Contains(strings.ToUpper(string(outbox)), strings.ToUpper(required)) {
			t.Errorf("000037_audit_outbox.up.sql missing delivery contract %q", required)
		}
	}

	deliveries, _ := migrationFiles.ReadFile("000039_push_deliveries.up.sql")
	for _, required := range []string{
		"REFERENCES \"user\"(id) ON DELETE CASCADE",
		"UNIQUE INDEX push_deliveries_dedupe_idx",
		"CHECK (attempts >= 0)",
		"CHECK (status IN ('pending', 'retry', 'sent', 'revoked', 'failed'))",
	} {
		if !strings.Contains(strings.ToUpper(string(deliveries)), strings.ToUpper(required)) {
			t.Errorf("000039_push_deliveries.up.sql missing safety contract %q", required)
		}
	}
}

func TestRowVersionMigrationPreservesCASInvariant(t *testing.T) {
	contents, err := migrationFiles.ReadFile("000040_row_versions.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToUpper(string(contents))
	for _, table := range []string{"TRANSACTIONS", "CATEGORIES"} {
		if !strings.Contains(sql, "ALTER TABLE "+table) {
			t.Fatalf("000040_row_versions.up.sql does not alter %s", strings.ToLower(table))
		}
	}
	for _, required := range []string{
		"ADD COLUMN VERSION BIGINT NOT NULL DEFAULT 1",
		"ADD CONSTRAINT TRANSACTIONS_VERSION_CHECK CHECK (VERSION >= 1)",
		"ADD CONSTRAINT CATEGORIES_VERSION_CHECK CHECK (VERSION >= 1)",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("000040_row_versions.up.sql missing CAS invariant %q", required)
		}
	}
}
