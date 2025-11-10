package httputils

import (
	"net/http"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/json/standartjson"
)

func ErrorJSON(w http.ResponseWriter, data any, code int) {
	d, err := standartjson.Marshal(data)
	if err != nil {
		http.Error(w, domain.ErrInternal.Message, domain.ErrInternal.Code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(d)
}

func SendJSON(w http.ResponseWriter, data any) {
	d, err := standartjson.Marshal(data)
	if err != nil {
		http.Error(w, domain.ErrInternal.Message, domain.ErrInternal.Code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(d)
}
