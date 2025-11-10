package list

import (
	"net/http"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/server/httputils"
)

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
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
