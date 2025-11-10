package update

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service2/internal/domain"
	"time"
)

var (
	ErrOperationCanceled = errors.New("update: operation canceled")
	ErrNotFound          = errors.New("update: no records found")
	ErrDatabaseFailure   = errors.New("update: database failed")
)

type Updater interface {
	UpdateStatus(ctx context.Context, id int, status domain.Status) error
}

type Usecase struct {
	Updater Updater
}

func New(up Updater) *Usecase {
	return &Usecase{
		Updater: up,
	}
}

func (u *Usecase) EventHandler(ctx context.Context, event domain.Event) error {
	if err := u.update(ctx, event); err != nil {
		return err
	}
	return nil
}

func (u *Usecase) update(ctx context.Context, event domain.Event) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
	default:
		if err := u.Updater.UpdateStatus(ctx, event.ID, domain.StatusProcessing); err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w: %v", ErrNotFound, err)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
		case <-time.After(time.Second * 10):
		}
		if err := u.Updater.UpdateStatus(ctx, event.ID, domain.StatusCompleted); err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf("%w: %v", ErrNotFound, err)
			}
			return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
		return nil
	}
}
