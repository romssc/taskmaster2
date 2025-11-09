package create

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/adapter/storage"
	"taskmaster2/service1/internal/domain"
	"time"

	"github.com/google/uuid"
)

type Creator interface {
	CreateTask(ctx context.Context, i Input) (Output, error)
}

type Usecase struct {
	Creator Creator
}

var u *Usecase

func New(c Creator) {
	u = &Usecase{Creator: c}
}

func (u *Usecase) CreateTask(ctx context.Context, i Input) (Output, error) {
	id, err := u.Creator.CreateTask(ctx, setTaskValues(i))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrAlreadyExists):
			return Output{}, fmt.Errorf("%w: %v", ErrAlreadyExists, err)
		default:
			return Output{}, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return id, nil
}

func setTaskValues(i Input) Input {
	uuid := uuid.New().ID()
	i.ID = int(uuid)
	i.CreatedAt = time.Now().Local()
	i.Status = domain.StatusCreated
	return i
}
