package listid

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/adapter/storage"
)

type Getter interface {
	GetTaskByID(ctx context.Context, i Input) (Output, error)
}

type Usecase struct {
	Getter Getter
}

var u *Usecase

func New(g Getter) {
	u = &Usecase{
		Getter: g,
	}
}

func (u *Usecase) GetTaskByID(ctx context.Context, i Input) (Output, error) {
	task, err := u.Getter.GetTaskByID(ctx, i)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNoRows):
			return Output{}, fmt.Errorf("%w: %v", ErrNoRows, err)
		default:
			return Output{}, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return task, nil
}
