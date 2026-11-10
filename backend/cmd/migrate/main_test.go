package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

type fakeMigrationEngine struct {
	version uint
	dirty   bool
	err     error
	forced  []int
}

func (engine *fakeMigrationEngine) Version() (uint, bool, error) {
	return engine.version, engine.dirty, engine.err
}

func (engine *fakeMigrationEngine) Force(version int) error {
	engine.forced = append(engine.forced, version)
	return engine.err
}

func TestRepairAssistantDelegation54RecoversExactRolledBackFailure(t *testing.T) {
	engine := &fakeMigrationEngine{version: 54, dirty: true}
	err := repairAssistantDelegation54(engine, func() (bool, error) { return false, nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(engine.forced) != 1 || engine.forced[0] != 53 {
		t.Fatalf("forced versions = %v, want [53]", engine.forced)
	}
}

func TestRepairAssistantDelegation54IsNoopForCleanHistory(t *testing.T) {
	for _, version := range []uint{53, 54, 55} {
		engine := &fakeMigrationEngine{version: version}
		if err := repairAssistantDelegation54(engine, func() (bool, error) {
			t.Fatal("clean history must not inspect the migration table")
			return false, nil
		}); err != nil {
			t.Fatalf("version %d: %v", version, err)
		}
		if len(engine.forced) != 0 {
			t.Fatalf("version %d forced unexpectedly: %v", version, engine.forced)
		}
	}
}

func TestRepairAssistantDelegation54RejectsUnsafeDirtyStates(t *testing.T) {
	tests := []struct {
		name        string
		engine      *fakeMigrationEngine
		tableExists bool
	}{
		{name: "different dirty version", engine: &fakeMigrationEngine{version: 53, dirty: true}},
		{name: "partial table", engine: &fakeMigrationEngine{version: 54, dirty: true}, tableExists: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := repairAssistantDelegation54(test.engine, func() (bool, error) {
				return test.tableExists, nil
			})
			if err == nil || !strings.Contains(err.Error(), "refuse") {
				t.Fatalf("error = %v, want fail-closed refusal", err)
			}
			if len(test.engine.forced) != 0 {
				t.Fatalf("forced unexpectedly: %v", test.engine.forced)
			}
		})
	}
}

func TestRepairAssistantDelegation54PropagatesInspectionErrors(t *testing.T) {
	engine := &fakeMigrationEngine{version: 54, dirty: true}
	err := repairAssistantDelegation54(engine, func() (bool, error) {
		return false, errors.New("database unavailable")
	})
	if err == nil || !strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("error = %v, want inspection failure", err)
	}
	if len(engine.forced) != 0 {
		t.Fatalf("forced unexpectedly: %v", engine.forced)
	}
}

func TestRepairAssistantDelegation54TreatsEmptyHistoryAsClean(t *testing.T) {
	engine := &fakeMigrationEngine{err: migrate.ErrNilVersion}
	if err := repairAssistantDelegation54(engine, func() (bool, error) {
		t.Fatal("empty history must not inspect the migration table")
		return false, nil
	}); err != nil {
		t.Fatal(err)
	}
}
