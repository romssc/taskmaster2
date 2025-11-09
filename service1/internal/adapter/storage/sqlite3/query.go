package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"taskmaster2/service1/internal/adapter/storage"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	listid "taskmaster2/service1/internal/usecase/list_id"
	"time"
)

func (s *Storage) CreateTask(ctx context.Context, i create.Input) (create.Output, error) {
	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO tasks (id, title, created_at, status) VALUES($1, $2, $3, $4);`)
	if err != nil {
		return create.Output{}, fmt.Errorf("%w: %v", storage.ErrPreparingStatement, err)
	}
	defer stmt.Close()
	createdAtUnix := i.CreatedAt.Unix()
	_, err = stmt.ExecContext(ctx, i.ID, i.Title, createdAtUnix, i.Status)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return create.Output{}, fmt.Errorf("%w: %v", storage.ErrAlreadyExists, err)
		}
		return create.Output{}, fmt.Errorf("%w: %v", storage.ErrExecutingStatement, err)
	}
	return create.Output{ID: i.ID}, nil
}

func (s *Storage) GetTasks(ctx context.Context, i list.Input) (list.Output, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT id, title, created_at, status FROM tasks ORDER BY created_at DESC;")
	if err != nil {
		return list.Output{}, fmt.Errorf("%w: %v", storage.ErrPreparingStatement, err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return list.Output{}, fmt.Errorf("%w: %v", storage.ErrExecutingStatement, err)
	}
	defer rows.Close()
	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var createdAtUnix int64
		if err := rows.Scan(&task.ID, &task.Title, &createdAtUnix, &task.Status); err != nil {
			return list.Output{}, fmt.Errorf("%w: %v", storage.ErrScanningRows, err)
		}
		task.CreatedAt = time.Unix(createdAtUnix, 0)
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return list.Output{}, fmt.Errorf("%w: %v", storage.ErrScanningRows, err)
	}
	return list.Output{Tasks: tasks}, nil
}

func (s *Storage) GetTaskByID(ctx context.Context, i listid.Input) (listid.Output, error) {
	stmt, err := s.db.PrepareContext(ctx, `SELECT id, title, created_at, status FROM tasks WHERE id = $1`)
	if err != nil {
		return listid.Output{}, fmt.Errorf("%w: %v", storage.ErrPreparingStatement, err)
	}
	defer stmt.Close()
	var task domain.Task
	var createdAtUnix int64
	err = stmt.QueryRowContext(ctx, i.ID).Scan(&task.ID, &task.Title, &createdAtUnix, &task.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return listid.Output{}, fmt.Errorf("%w: %v", storage.ErrNoRows, err)
		}
		return listid.Output{}, fmt.Errorf("%w: %v", storage.ErrExecutingStatement, err)
	}
	task.CreatedAt = time.Unix(createdAtUnix, 0)
	return listid.Output{Task: task}, nil
}
