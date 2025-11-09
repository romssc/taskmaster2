package list

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/domain"
)

var (
	ErrDatabaseFailure = errors.New("usecase: database failed")
)

type Getter interface {
	GetTasks(ctx context.Context) ([]domain.Record, error)
}

type Usecase struct {
	Getter Getter
}

var u *Usecase

func New(g Getter) {
	u = &Usecase{Getter: g}
}

func (u *Usecase) GetTasks(ctx context.Context) ([]domain.Record, error) {
	tasks, err := u.Getter.GetTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}
	return tasks, nil
}
