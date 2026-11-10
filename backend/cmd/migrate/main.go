package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mypocket/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	commandUp                          = "up"
	commandVersion                     = "version"
	commandForce                       = "force"
	commandRepairAssistantDelegation54 = "repair-assistant-delegation-54"
	environmentMigrationsPath          = "MIGRATIONS_PATH"
)

type migrationRepairEngine interface {
	Version() (uint, bool, error)
	Force(int) error
}

func main() {
	command := commandUp
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		log.Fatalf("resolve migrations directory: %v", err)
	}
	configuration := config.Load()
	engine, err := migrate.New("file://"+migrationsPath, configuration.DatabaseURL)
	if err != nil {
		log.Fatalf("open migration engine: %v", err)
	}

	switch command {
	case commandUp:
		err = engine.Up()
	case commandVersion:
		version, dirty, versionErr := engine.Version()
		if errors.Is(versionErr, migrate.ErrNilVersion) {
			fmt.Println("version: none")
			return
		}
		if versionErr != nil {
			log.Fatalf("read migration version: %v", versionErr)
		}
		fmt.Printf("version: %d dirty: %t\n", version, dirty)
		return
	case commandForce:
		if len(os.Args) < 3 {
			log.Fatalf("force requires a migration version")
		}
		version, parseErr := strconv.Atoi(os.Args[2])
		if parseErr != nil {
			log.Fatalf("parse migration version: %v", parseErr)
		}
		err = engine.Force(version)
	case commandRepairAssistantDelegation54:
		err = repairAssistantDelegation54(engine, func() (bool, error) {
			database, openErr := gorm.Open(postgres.Open(configuration.DatabaseURL), &gorm.Config{})
			if openErr != nil {
				return false, fmt.Errorf("open database for migration repair: %w", openErr)
			}
			var exists bool
			if queryErr := database.Raw(
				"SELECT to_regclass('public.assistant_delegations') IS NOT NULL",
			).Scan(&exists).Error; queryErr != nil {
				return false, fmt.Errorf("inspect assistant_delegations before migration repair: %w", queryErr)
			}
			return exists, nil
		})
	default:
		log.Fatalf("unsupported migration command %q; use up, version, or the documented recovery command", command)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration %s failed: %v", command, err)
	}
	log.Printf("migration %s completed", command)
}

func repairAssistantDelegation54(
	engine migrationRepairEngine,
	tableExists func() (bool, error),
) error {
	version, dirty, err := engine.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read migration version before assistant delegation repair: %w", err)
	}
	if !dirty {
		return nil
	}
	if version != 54 {
		return fmt.Errorf("refuse migration repair: dirty version is %d, expected 54", version)
	}
	exists, err := tableExists()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("refuse migration repair: public.assistant_delegations already exists")
	}
	if err := engine.Force(53); err != nil {
		return fmt.Errorf("reset rolled-back migration 54 to clean version 53: %w", err)
	}
	return nil
}

func resolveMigrationsPath() (string, error) {
	path := os.Getenv(environmentMigrationsPath)
	if path == "" {
		path = "migrations"
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absPath), nil
}
