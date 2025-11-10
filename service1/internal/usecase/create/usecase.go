package create

import (
	"context"
	"errors"
	"fmt"

	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/domain"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAlreadyExists     = errors.New("usecase: task already exists in the database")
	ErrDatabaseFailure   = errors.New("usecase: database failed")
	ErrBrokerUnavailable = errors.New("usecase: broker unavailable")
	ErrBrokerFailure     = errors.New("usecase: broker failed")
)

type Creator interface {
	CreateTask(ctx context.Context, task domain.Record) (int, error)
}

type Publisher interface {
	PublishEvent(ctx context.Context, event domain.Event) error
}

type Usecase struct {
	Creator   Creator
	Publisher Publisher
}

var u *Usecase

func New(c Creator, p Publisher) {
	u = &Usecase{
		Creator:   c,
		Publisher: p,
	}
}

func (u *Usecase) CreateTask(ctx context.Context, task domain.Record) (int, error) {
	task, event := setValues(task)
	id, err := u.Creator.CreateTask(ctx, task)
	if err != nil {
		switch {
		case errors.Is(err, inmemory.ErrAlreadyExists):
			return 0, fmt.Errorf("%w: %v", ErrAlreadyExists, err)
		default:
			return 0, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	if err := u.Publisher.PublishEvent(ctx, event); err != nil {
		switch {
		case errors.Is(err, kafkaa.ErrClosed):
			return 0, fmt.Errorf("%w: %v", ErrBrokerUnavailable, err)
		default:
			return 0, fmt.Errorf("%w: %v", ErrBrokerFailure, err)
		}
	}
	return id, nil
}

func setValues(task domain.Record) (domain.Record, domain.Event) {
	uuid := uuid.New().ID()
	timestamp := time.Now().Local()

	task.ID = int(uuid)
	task.CreatedAt = timestamp
	task.Status = domain.StatusProcessing
	event := domain.Event{
		ID: int(uuid),
	}

	return task, event
}
