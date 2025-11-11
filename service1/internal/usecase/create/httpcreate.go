package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/json/standartjson"
	"taskmaster2/service1/internal/pkg/server/httputils"
)

var (
	ErrEmptyTitle        = errors.New("create: invalid body: empty task title")
	ErrUnmarshalingBody  = errors.New("create: failed to unmarshal body")
	ErrReadingBody       = errors.New("create: failed to read a body")
	ErrBrokerUnavailable = errors.New("create: broker unavailable")
	ErrBrokerFailure     = errors.New("create: broker failed")
	ErrOperationCanceled = errors.New("create: operation canceled, request killed")
	ErrGeneratingID      = errors.New("create: failed to generate id")
)

type Publisher interface {
	PublishEvent(ctx context.Context, event domain.Event) error
}

type Generator interface {
	Gen() (int, error)
}

type Usecase struct {
	Publisher Publisher
	Generator Generator
}

func New(p Publisher, g Generator) *Usecase {
	return &Usecase{
		Publisher: p,
		Generator: g,
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
	if err := standartjson.Unmarshal(body, &task); err != nil {
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
	output, err := u.CreateTask(ctx, task)
	if err != nil {
		switch {
		case errors.Is(err, ErrBrokerUnavailable):
			httputils.ErrorJSON(w, domain.ErrBrokerUnavailable, domain.ErrBrokerUnavailable.Code)
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
		}
		return
	}

	httputils.SendJSON(w, output)
}

func validateTask(task domain.Record) error {
	if task.Title == "" {
		return fmt.Errorf("%w", ErrEmptyTitle)
	}
	return nil
}

func (u *Usecase) CreateTask(ctx context.Context, task domain.Record) (int, error) {
	id, err := u.Generator.Gen()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrGeneratingID, err)
	}
	event := domain.Event{
		Action: domain.ActionUpdate,
		Record: domain.Record{
			ID:    id,
			Title: task.Title,
		},
	}
	if err := u.Publisher.PublishEvent(ctx, event); err != nil {
		switch {
		case errors.Is(err, kafkaa.ErrOperationCanceled):
			return 0, fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		case errors.Is(err, kafkaa.ErrClosed):
			return 0, fmt.Errorf("%w: %v", ErrBrokerUnavailable, err)
		default:
			return 0, fmt.Errorf("%w: %v", ErrBrokerFailure, err)
		}
	}
	return event.Record.ID, nil
}
