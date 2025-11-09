package list

import (
	"net/http"
	"taskmaster2/internal/domain"
	"taskmaster2/internal/utils/httputils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.ErrorJSON(w, domain.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	input := Input{}
	output, err := u.GetTasks(ctx, input)
	if err != nil {
		httputils.ErrorJSON(w, domain.ErrInternal, http.StatusInternalServerError)
		return
	}

	httputils.SendJSON(w, output)
}
