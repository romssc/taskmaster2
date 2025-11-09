package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"taskmaster2/service1/internal/domain"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrOpeningConnection  = errors.New("sqlite3: failed to open connection")
	ErrPinging            = errors.New("sqlite3: failed to ping")
	ErrClosingConnection  = errors.New("sqlite3: failed to close connection")
	ErrPreparingStatement = errors.New("sqlite3: failed to prepare a statement")
	ErrExecutingStatement = errors.New("sqlite3: failed to execute a statement")
	ErrAlreadyExists      = errors.New("sqlite3: failed due to record already existing")
	ErrScanningRows       = errors.New("sqlite3: failed to scan rows")
	ErrIteratingRows      = errors.New("sqlite3: failed while iterating rows")
	ErrNoRows             = errors.New("sqlite3: failed due to no records available")
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
		return &Storage{}, fmt.Errorf("%w: %v", ErrOpeningConnection, err)
	}
	if err := db.Ping(); err != nil {
		return &Storage{}, fmt.Errorf("%w: %v", ErrPinging, err)
	}
	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosingConnection, err)
	}
	return nil
}

func (s *Storage) CreateTask(ctx context.Context, task domain.Record) (int, error) {
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO tasks (id, title, created_at, status) VALUES($1, $2, $3, $4);`)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrPreparingStatement, err)
	}
	defer stmt.Close()
	createdAtUnix := task.CreatedAt.Unix()
	_, err = stmt.ExecContext(ctx, task.ID, task.Title, createdAtUnix, task.Status)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, fmt.Errorf("%w: %v", ErrAlreadyExists, err)
		}
		return 0, fmt.Errorf("%w: %v", ErrExecutingStatement, err)
	}
	return task.ID, nil
}

func (s *Storage) GetTasks(ctx context.Context) ([]domain.Record, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT id, title, created_at, status FROM tasks ORDER BY created_at DESC;")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPreparingStatement, err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExecutingStatement, err)
	}
	defer rows.Close()
	var tasks []domain.Record
	for rows.Next() {
		var task domain.Record
		var createdAtUnix int64
		if err := rows.Scan(&task.ID, &task.Title, &createdAtUnix, &task.Status); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrScanningRows, err)
		}
		task.CreatedAt = time.Unix(createdAtUnix, 0)
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScanningRows, err)
	}
	return tasks, nil
}

func (s *Storage) GetTaskByID(ctx context.Context, id int) (domain.Record, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT id, title, created_at, status FROM tasks WHERE id = $1`)
	if err != nil {
		return domain.Record{}, fmt.Errorf("%w: %v", ErrPreparingStatement, err)
	}
	defer stmt.Close()
	var task domain.Record
	var createdAtUnix int64
	err = stmt.QueryRowContext(ctx, id).Scan(&task.ID, &task.Title, &createdAtUnix, &task.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Record{}, fmt.Errorf("%w: %v", ErrNoRows, err)
		}
		return domain.Record{}, fmt.Errorf("%w: %v", ErrExecutingStatement, err)
	}
	task.CreatedAt = time.Unix(createdAtUnix, 0)
	return task, nil
}
