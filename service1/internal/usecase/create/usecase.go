package create

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/adapter/storage/sqlite3"
	"taskmaster2/service1/internal/domain"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAlreadyExists   = errors.New("usecase: task already exists in the database")
	ErrDatabaseFailure = errors.New("usecase: database failed")
)

type Creator interface {
	CreateTask(ctx context.Context, task domain.Record) (int, error)
}

type Usecase struct {
	Creator Creator
}

var u *Usecase

func New(c Creator) {
	u = &Usecase{Creator: c}
}

func (u *Usecase) CreateTask(ctx context.Context, task domain.Record) (int, error) {
	id, err := u.Creator.CreateTask(ctx, setTaskValues(task))
	if err != nil {
		switch {
		case errors.Is(err, sqlite3.ErrAlreadyExists):
			return 0, fmt.Errorf("%w: %v", ErrAlreadyExists, err)
		default:
			return 0, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return id, nil
}

func setTaskValues(task domain.Record) domain.Record {
	uuid := uuid.New().ID()
	task.ID = int(uuid)
	task.CreatedAt = time.Now().Local()
	task.Status = domain.StatusProcessing
	return task
}
