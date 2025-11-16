package listid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"service1/internal/adapter/storage/inmemory"
	"service1/internal/domain"
)

var (
	ErrMalformedID       = errors.New("listid: client sent a malformed id")
	ErrNotFound          = errors.New("listid: no records found")
	ErrDatabaseFailure   = errors.New("listid: database failed")
	ErrOperationCanceled = errors.New("listid: operation canceled, request killed")
)

type Config struct{}

type Getter interface {
	GetTaskByID(ctx context.Context, id int) (domain.Record, error)
}

type Encoder interface {
	Marshal(data any) ([]byte, error)
}

type Usecase struct {
	Config Config

	Getter Getter

	Encoder Encoder
}

func (u *Usecase) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		u.sendJSON(w, domain.ErrMethodNotAllowed, domain.ErrMethodNotAllowed.Code)
		return
	}

	idRaw := r.PathValue("id")
	id, validateErr := validatePathValues(idRaw)
	if validateErr != nil {
		u.sendJSON(w, domain.ErrMalformedPathValue, domain.ErrMalformedPathValue.Code)
		return
	}

	ctx := r.Context()
	output, uErr := u.GetTaskByID(ctx, id)
	if uErr != nil && !errors.Is(uErr, ErrOperationCanceled) {
		switch {
		case errors.Is(uErr, ErrNotFound):
			u.sendJSON(w, domain.ErrNotFound, domain.ErrNotFound.Code)
		default:
			u.sendJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
		}
		return
	}

	u.sendJSON(w, output, http.StatusOK)
}

func validatePathValues(id string) (int, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, ErrMalformedID
	}
	return i, nil
}

func (u *Usecase) sendJSON(w http.ResponseWriter, data any, code int) {
	d, err := u.Encoder.Marshal(data)
	if err != nil {
		http.Error(w, domain.ErrInternal.Message, domain.ErrInternal.Code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(d)
}

func (u *Usecase) GetTaskByID(ctx context.Context, id int) (domain.Record, error) {
	task, err := u.Getter.GetTaskByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, inmemory.ErrOperationCanceled):
			return domain.Record{}, fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		case errors.Is(err, inmemory.ErrNotFound):
			return domain.Record{}, fmt.Errorf("%w: %v", ErrNotFound, err)
		default:
			return domain.Record{}, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return task, nil
}
