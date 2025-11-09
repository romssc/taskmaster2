package create

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/utils/httputils"
	"taskmaster2/service1/pkg/json/standartjson"
)

var (
	ErrEmptyTitle       = errors.New("handler: client sent a request with an empty task title")
	ErrUnmarshalingBody = errors.New("handler: failed to unmarshal body")
	ErrReadingBody      = errors.New("handler: failed to read a body")
)

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, domain.ErrMethodNotAllowed.Code)
		return
	}
	defer r.Body.Close()
	task, err := readBody(r.Body)
	if err != nil {
		switch err {
		case ErrReadingBody:
			httputils.ErrorJSON(w, domain.ErrMalformedBody, domain.ErrMalformedBody.Code)
			return
		case ErrUnmarshalingBody:
			httputils.ErrorJSON(w, domain.ErrMalformedBody, domain.ErrMalformedBody.Code)
			return
		default:
			httputils.ErrorJSON(w, domain.ErrInternal, domain.ErrInternal.Code)
			return
		}
	}
	err = validateTask(task)
	if err != nil {
		switch err {
		case ErrEmptyTitle:
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
		case errors.Is(err, ErrAlreadyExists):
			httputils.ErrorJSON(w, domain.ErrAlreadyExists, domain.ErrAlreadyExists.Code)
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
