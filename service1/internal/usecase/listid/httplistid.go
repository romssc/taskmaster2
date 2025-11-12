package listid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/server/httputils"
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

type Usecase struct {
	Config Config

	Getter Getter
}

func New(c Config, g Getter) *Usecase {
	return &Usecase{
		Config: c,

		Getter: g,
	}
}

func (u *Usecase) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, domain.ErrMethodNotAllowed.Code)
		return
	}

	idRaw := r.PathValue("id")
	id, err := validatePathValues(idRaw)
	if err != nil {
		httputils.ErrorJSON(w, domain.ErrMalformedPathValue, domain.ErrMalformedPathValue.Code)
		return
	}

	ctx := r.Context()
	output, err := u.GetTaskByID(ctx, id)
	if err != nil && !errors.Is(err, ErrOperationCanceled) {
		switch {
		case errors.Is(err, ErrNotFound):
			httputils.ErrorJSON(w, domain.ErrNotFound, domain.ErrNotFound.Code)
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
		}
		return
	}

	httputils.SendJSON(w, output)
}

func validatePathValues(id string) (int, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, ErrMalformedID
	}
	return i, nil
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
