package list

import (
	"context"
	"fmt"
)

type Getter interface {
	GetTasks(ctx context.Context, i Input) (Output, error)
}

type Usecase struct {
	Getter Getter
}

var u *Usecase

func New(g Getter) {
	u = &Usecase{Getter: g}
}

func (u *Usecase) GetTasks(ctx context.Context, i Input) (Output, error) {
	tasks, err := u.Getter.GetTasks(ctx, i)
	if err != nil {
		return Output{}, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}
	return tasks, nil
}
