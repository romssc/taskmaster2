package kafkarouter

import (
	"context"

	"service2/internal/domain"
	"service2/internal/usecase/update"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Update *update.Usecase
}

type Router struct {
	Config *Config

	Handlers *Handlers
}

type Handler interface {
	EventHandler(ctx context.Context, message kafka.Message)
}

type Handlers struct {
	update Handler
}

func New(c *Config) *Router {
	return &Router{
		Config: c,
		Handlers: &Handlers{
			update: c.Update,
		},
	}
}

func (r *Router) Route(ctx context.Context, message kafka.Message) {
	switch domain.Action(string(message.Key)) {
	case domain.ActionUpdate:
		r.Handlers.update.EventHandler(ctx, message)
	}
}
