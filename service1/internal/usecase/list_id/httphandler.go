package listid

import (
	"errors"
	"net/http"
	"strconv"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/utils/httputils"
)

var (
	ErrMalformedID = errors.New("handler: client sent a malformed id")
)

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
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
	output, err := u.GetTaskByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoRows):
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
