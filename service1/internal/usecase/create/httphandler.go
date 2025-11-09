package create

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/utils/httputils"
	"taskmaster2/service1/pkg/renderjson"
)

func Handler(w http.ResponseWriter, r *http.Request) {
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
	input := Input{Task: task}
	output, err := u.CreateTask(ctx, input)
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

func readBody(r io.Reader) (domain.Task, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return domain.Task{}, fmt.Errorf("%w: %v", ErrReadingBody, err)
	}
	var task domain.Task
	if err := renderjson.Unmarshal(body, &task); err != nil {
		return domain.Task{}, fmt.Errorf("%w: %v", ErrUnmarshalingBody, err)
	}
	return task, nil
}

func validateTask(task domain.Task) error {
	if task.Title == "" {
		return ErrEmptyTitle
	}
	return nil
}
