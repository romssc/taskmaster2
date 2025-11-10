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
)

type Usecase struct{}

func New() *Usecase {
	return &Usecase{}
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
		select {
		case <-ctx.Done():
			return fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
		case <-time.After(time.Second * 10):
		}
		return nil
	}
}
