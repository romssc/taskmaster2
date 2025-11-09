package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"
)

func New() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/list", list.HTTPHandler)
	m.HandleFunc("/list/{id}", listid.HTTPHandler)
	m.HandleFunc("/create", create.HTTPHandler)
	return m
}
