package kafkarouter

import (
	"context"

	"service2/internal/domain"
	"service2/internal/usecase/update"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Update update.Config `mapstructure:"update"`
}

type Routes struct {
	Update *update.Usecase
}

type Router struct {
	Handlers *Handlers
}

type Handlers struct {
	update EventHandler
}

type EventHandler func(ctx context.Context, message kafka.Message)

func New(r *Routes) *Router {
	return &Router{
		Handlers: &Handlers{
			update: r.Update.EventHandler,
		},
	}
}

func (r *Router) Route(ctx context.Context, message kafka.Message) {
	switch domain.Action(string(message.Key)) {
	case domain.ActionUpdate:
		r.Handlers.update(ctx, message)
	}
}
