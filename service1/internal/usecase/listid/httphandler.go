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
	ErrMalformedID     = errors.New("listid: client sent a malformed id")
	ErrNotFound        = errors.New("listid: no records found")
	ErrDatabaseFailure = errors.New("listid: database failed")
)

type Getter interface {
	GetTaskByID(ctx context.Context, id int) (domain.Record, error)
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
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, domain.ErrMethodNotAllowed.Code)
		return
	}
	id, err := readPathValues(r)
	if err != nil {
		httputils.ErrorJSON(w, domain.ErrMalformedPathValue, domain.ErrMalformedPathValue.Code)
		return
	}

	ctx := r.Context()
	output, err := u.getTaskByID(ctx, id)
	if err != nil {
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

func readPathValues(r *http.Request) (int, error) {
	id := r.PathValue("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, ErrMalformedID
	}
	return i, nil
}

func (u *Usecase) getTaskByID(ctx context.Context, id int) (domain.Record, error) {
	task, err := u.Getter.GetTaskByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, inmemory.ErrNotFound):
			return domain.Record{}, fmt.Errorf("%w: %v", ErrNotFound, err)
		default:
			return domain.Record{}, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
		}
	}
	return task, nil
}
