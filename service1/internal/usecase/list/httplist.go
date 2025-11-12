package list

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/server/httputils"
)

var (
	ErrDatabaseFailure   = errors.New("list: database failed")
	ErrOperationCanceled = errors.New("list: operation canceled, request killed")
)

type Config struct{}

type Getter interface {
	GetTasks(ctx context.Context) ([]domain.Record, error)
}

type Usecase struct {
	Getter Getter
}

func New(g Getter) *Usecase {
	return &Usecase{
		Getter: g,
	}
}

func (u *Usecase) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	output, err := u.GetTasks(ctx)
	if err != nil {
		httputils.ErrorJSON(w, domain.ErrInternal, http.StatusInternalServerError)
		return
	}

	httputils.SendJSON(w, output)
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
