package kafkarouter

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service2/internal/domain"
	"taskmaster2/service2/internal/usecase/update"
)

var (
	ErrOperationCanceled = errors.New("kafkarouter: operation canceled")
	ErrFailedToProccess  = errors.New("kafkarouter: failed to proccess")
)

type Config struct {
	Update *update.Usecase
}

type Router struct {
	Handlers *Handlers
}

type Handler interface {
	EventHandler(ctx context.Context, event domain.Event) error
}

type Handlers struct {
	Update Handler
}

func New(c *Config) *Router {
	return &Router{
		Handlers: &Handlers{
			Update: c.Update,
		},
	}
}

func (r *Router) Pick(ctx context.Context, action domain.Action, event domain.Event) error {
	switch action {
	case domain.ActionUpdate:
		err := r.Handlers.Update.EventHandler(ctx, event)
		if err != nil {
			switch {
			case errors.Is(err, update.ErrOperationCanceled):
				return fmt.Errorf("%w: %v", ErrOperationCanceled, err)
			default:
				return fmt.Errorf("%w: %v", ErrFailedToProccess, err)
			}
		}
	}
	return nil
}
