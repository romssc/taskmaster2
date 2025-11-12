package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/server/httputils"
	"time"
)

var (
	ErrEmptyTitle           = errors.New("create: invalid body: empty task title")
	ErrStorageAlreadyExists = errors.New("create: duplicate")
	ErrStorageFailure       = errors.New("create: storage failed")
	ErrBrokerUnavailable    = errors.New("create: broker unavailable")
	ErrBrokerFailure        = errors.New("create: broker failed")
	ErrOperationCanceled    = errors.New("create: operation canceled, request killed")
	ErrGeneratingID         = errors.New("create: failed to generate id")
)

type Config struct {
	FailTimeout time.Duration `yaml:"fail_timeout"`
}

type Creator interface {
	CreateTask(ctx context.Context, task domain.Record) (int, error)
	UpdateOrCreateTask(ctx context.Context, task domain.Record) error
}

type Publisher interface {
	PublishEvent(ctx context.Context, event domain.Event) error
}

type Generator interface {
	Gen() (int, error)
}

type Timer interface {
	TimeNow() int64
}

type Decoder interface {
	Unmarshal(data []byte, v any) error
}

type Usecase struct {
	Config Config

	Creator   Creator
	Publisher Publisher
	Generator Generator
	Timer     Timer
	Decoder   Decoder
}

func New(c Config, cr Creator, p Publisher, g Generator, t Timer, d Decoder) *Usecase {
	return &Usecase{
		Config: c,

		Creator:   cr,
		Publisher: p,
		Generator: g,
		Timer:     t,
		Decoder:   d,
	}
}

func (u *Usecase) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, domain.ErrMethodNotAllowed.Code)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httputils.ErrorJSON(w, domain.ErrMalformedBody, domain.ErrMalformedBody.Code)
		return
	}
	var task domain.Record
	if err := u.Decoder.Unmarshal(body, &task); err != nil {
		httputils.ErrorJSON(w, domain.ErrMalformedBody, domain.ErrMalformedBody.Code)
		return
	}

	if err := validateTask(task); err != nil {
		switch {
		case errors.Is(err, ErrEmptyTitle):
			httputils.ErrorJSON(w, domain.ErrEmptyTitle, domain.ErrEmptyTitle.Code)
			return
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
			return
		}
	}

	ctx := r.Context()
	event, err := u.CreateTask(ctx, task)
	if err != nil && !errors.Is(err, ErrOperationCanceled) {
		switch {
		case errors.Is(err, ErrStorageAlreadyExists):
			httputils.ErrorJSON(w, domain.ErrAlreadyExists, domain.ErrAlreadyExists.Code)
		case errors.Is(err, ErrBrokerUnavailable):
			httputils.ErrorJSON(w, domain.ErrBrokerUnavailable, domain.ErrBrokerUnavailable.Code)
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
		}
		return
	}

	httputils.SendJSON(w, event.Record.ID)
}

func validateTask(task domain.Record) error {
	if task.Title == "" {
		return fmt.Errorf("%w", ErrEmptyTitle)
	}
	return nil
}

func (u *Usecase) CreateTask(ctx context.Context, task domain.Record) (domain.Event, error) {
	event, ceErr := u.createEvent(task)
	if ceErr != nil {
		return domain.Event{}, ceErr
	}
	record := event.Record
	_, ctErr := u.Creator.CreateTask(ctx, record)
	if ctErr != nil {
		switch {
		case errors.Is(ctErr, inmemory.ErrOperationCanceled):
			return domain.Event{}, fmt.Errorf("%w: %v", ErrOperationCanceled, ctErr)
		case errors.Is(ctErr, inmemory.ErrAlreadyExists):
			return domain.Event{}, fmt.Errorf("%w: %v", ErrStorageAlreadyExists, ctErr)
		default:
			return domain.Event{}, fmt.Errorf("%w: %v", ErrStorageFailure, ctErr)
		}
	}
	if pubErr := u.Publisher.PublishEvent(ctx, event); pubErr != nil {
		c, cancel := context.WithTimeout(context.Background(), u.Config.FailTimeout)
		defer cancel()
		_, markErr := u.markFailure(c, record)
		if markErr != nil {
			return domain.Event{}, markErr
		}
		switch {
		case errors.Is(pubErr, kafkaa.ErrOperationCanceled):
			return domain.Event{}, fmt.Errorf("%w: %v", ErrOperationCanceled, pubErr)
		case errors.Is(pubErr, kafkaa.ErrClosed):
			return domain.Event{}, fmt.Errorf("%w: %v", ErrBrokerUnavailable, pubErr)
		default:
			return domain.Event{}, fmt.Errorf("%w: %v", ErrBrokerFailure, pubErr)
		}
	}
	return event, nil
}

func (u *Usecase) createEvent(task domain.Record) (domain.Event, error) {
	id, err := u.Generator.Gen()
	if err != nil {
		return domain.Event{}, fmt.Errorf("%w: %v", ErrGeneratingID, err)
	}
	time := u.Timer.TimeNow()
	return domain.Event{
		Action: domain.ActionUpdate,
		Record: domain.Record{
			ID:        id,
			Title:     task.Title,
			CreatedAt: time,
			Status:    domain.StatusNew,
		},
	}, nil
}

func (u *Usecase) markFailure(ctx context.Context, task domain.Record) (domain.Record, error) {
	task.Status = domain.StatusFailed
	if err := u.Creator.UpdateOrCreateTask(ctx, task); err != nil {
		switch {
		case errors.Is(err, inmemory.ErrOperationCanceled):
			return domain.Record{}, fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		default:
			return domain.Record{}, fmt.Errorf("%w: %v", ErrStorageFailure, err)
		}
	}
	return task, nil
}
