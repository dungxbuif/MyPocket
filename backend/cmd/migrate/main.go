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
)

const (
	commandUp                 = "up"
	commandVersion            = "version"
	commandForce              = "force"
	environmentMigrationsPath = "MIGRATIONS_PATH"
)

func main() {
	command := commandUp
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		log.Fatalf("resolve migrations directory: %v", err)
	}
	engine, err := migrate.New("file://"+migrationsPath, config.Load().DatabaseURL)
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
	default:
		log.Fatalf("unsupported migration command %q; use up or version", command)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration %s failed: %v", command, err)
	}
	log.Printf("migration %s completed", command)
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
