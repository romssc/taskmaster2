package update

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service2/internal/domain"
	"time"
)

var (
	ErrOperationCanceled = errors.New("usecase: operation canceled")
	ErrNotFound          = errors.New("usecase: no records found")
	ErrDatabaseFailure   = errors.New("usecase: database failed")
)

type Updater interface {
	UpdateStatus(ctx context.Context, id int, status domain.Status) error
}

type Usecase struct {
	Updater Updater
}

var u *Usecase

func New(up Updater) {
	u = &Usecase{
		Updater: up,
	}
}

func (u *Usecase) Update(ctx context.Context, event domain.Event) error {
	select {
	case <-ctx.Done():
		if err := u.Updater.UpdateStatus(ctx, event.ID, domain.StatusFailed); err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w: %v", ErrNotFound, err)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
		return fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
	case <-time.After(time.Second * 10):
		if err := u.Updater.UpdateStatus(ctx, event.ID, domain.StatusCompleted); err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w: %v", ErrNotFound, err)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
		return nil
	}
}
