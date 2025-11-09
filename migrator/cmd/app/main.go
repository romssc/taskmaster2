package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := m(); err != nil {
		log.Fatalln(err)
	}
}

func m() error {
	driver, migrationsPath, dbPath, direction, err := loadEnvs()
	if err != nil {
		return err
	}
	migrator, err := migrate.New(migrationsPath, strings.Join([]string{driver, dbPath}, "://"))
	if err != nil {
		return fmt.Errorf("migrator: failed to create migrator: %v", err)
	}
	switch direction {
	case "0":
		if err := migrator.Up(); err != nil {
			return fmt.Errorf("migrator: failed to migrate up: %v", err)
		}
	case "1":
		if err := migrator.Down(); err != nil {
			return fmt.Errorf("migrator: failed to migrate down: %v", err)
		}
	}
	return nil
}

func loadEnvs() (string, string, string, string, error) {
	driver := os.Getenv("DRIVER")
	if driver == "" {
		return "", "", "", "", errors.New("migrator: driver not specified")
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		return "", "", "", "", errors.New("migrator: migrations path not specified")
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = ".tasks.db"
	}
	direction := os.Getenv("DIRECTION")
	switch {
	case direction != "" && direction != "0" && direction != "1":
		return "", "", "", "", errors.New("migrator: invalid direction value")
	case direction == "":
		direction = "1"
	}
	return driver, migrationsPath, dbPath, direction, nil
}
