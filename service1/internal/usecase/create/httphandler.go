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
	"taskmaster2/service1/internal/pkg/json/standartjson"
	"taskmaster2/service1/internal/pkg/server/httputils"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyTitle        = errors.New("create: client sent a request with an empty task title")
	ErrUnmarshalingBody  = errors.New("create: failed to unmarshal body")
	ErrReadingBody       = errors.New("create: failed to read a body")
	ErrAlreadyExists     = errors.New("create task already exists in the database")
	ErrDatabaseFailure   = errors.New("create: database failed")
	ErrBrokerUnavailable = errors.New("create: broker unavailable")
	ErrBrokerFailure     = errors.New("create: broker failed")
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

func New(c Creator, p Publisher) *Usecase {
	return &Usecase{
		Creator:   c,
		Publisher: p,
	}
}

func (u *Usecase) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, domain.ErrMethodNotAllowed.Code)
		return
	}
	defer r.Body.Close()
	task, err := readBody(r.Body)
	if err != nil {
		switch {
		case errors.Is(err, ErrReadingBody):
			httputils.ErrorJSON(w, domain.ErrMalformedBody, domain.ErrMalformedBody.Code)
			return
		case errors.Is(err, ErrUnmarshalingBody):
			httputils.ErrorJSON(w, domain.ErrMalformedBody, domain.ErrMalformedBody.Code)
			return
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
			return
		}
	}
	err = validateTask(task)
	if err != nil {
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
	output, err := u.createTask(ctx, task)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyExists):
			httputils.ErrorJSON(w, domain.ErrAlreadyExists, domain.ErrAlreadyExists.Code)
		case errors.Is(err, ErrBrokerUnavailable):
			httputils.ErrorJSON(w, domain.ErrBrokerUnavailable, domain.ErrBrokerUnavailable.Code)
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
		}
		return
	}

	httputils.SendJSON(w, output)
}

func readBody(r io.Reader) (domain.Record, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return domain.Record{}, fmt.Errorf("%w: %v", ErrReadingBody, err)
	}
	var task domain.Record
	if err := standartjson.Unmarshal(body, &task); err != nil {
		return domain.Record{}, fmt.Errorf("%w: %v", ErrUnmarshalingBody, err)
	}
	return task, nil
}

func validateTask(task domain.Record) error {
	if task.Title == "" {
		return ErrEmptyTitle
	}
	return nil
}

func (u *Usecase) createTask(ctx context.Context, task domain.Record) (int, error) {
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
