package sqlite3

import (
	"database/sql"
	"fmt"
	"taskmaster2/service1/internal/adapter/storage"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Path string `yaml:"path"`
}

type Storage struct {
	db *sql.DB
}

func New(c Config) (*Storage, error) {
	db, err := sql.Open("sqlite3", c.Path)
	if err != nil {
		return &Storage{}, fmt.Errorf("%w: %v", storage.ErrOpeningConnection, err)
	}
	if err := db.Ping(); err != nil {
		return &Storage{}, fmt.Errorf("%w: %v", storage.ErrPinging, err)
	}
	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%w: %v", storage.ErrClosingConnection, err)
	}
	return nil
}
