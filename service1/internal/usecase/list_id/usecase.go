package listid

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/adapter/storage/sqlite3"
	"taskmaster2/service1/internal/domain"
)

var (
	ErrNoRows          = errors.New("usecase: no records found")
	ErrDatabaseFailure = errors.New("usecase: database failed")
)

type Getter interface {
	GetTaskByID(ctx context.Context, id int) (domain.Record, error)
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

func (u *Usecase) GetTaskByID(ctx context.Context, id int) (domain.Record, error) {
	task, err := u.Getter.GetTaskByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sqlite3.ErrNoRows):
			return domain.Record{}, fmt.Errorf("%w: %v", ErrNoRows, err)
		default:
			return domain.Record{}, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return task, nil
}
