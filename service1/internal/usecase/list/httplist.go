package list

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"service1/internal/adapter/storage/inmemory"
	"service1/internal/domain"
)

var (
	ErrDatabaseFailure   = errors.New("list: database failed")
	ErrOperationCanceled = errors.New("list: operation canceled, request killed")
)

type Config struct{}

type Getter interface {
	GetTasks(ctx context.Context) ([]domain.Record, error)
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
		u.sendJSON(w, domain.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	output, err := u.GetTasks(ctx)
	if err != nil && !errors.Is(err, ErrOperationCanceled) {
		u.sendJSON(w, domain.ErrInternal, http.StatusInternalServerError)
		return
	}

	u.sendJSON(w, output, http.StatusOK)
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

func (u *Usecase) GetTasks(ctx context.Context) ([]domain.Record, error) {
	tasks, err := u.Getter.GetTasks(ctx)
	if err != nil {
		switch {
		case errors.Is(err, inmemory.ErrOperationCanceled):
			return nil, fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		case errors.Is(err, inmemory.ErrIncompatible):
			return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		default:
			return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return tasks, nil
}
