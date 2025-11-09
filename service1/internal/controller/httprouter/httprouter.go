package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	listid "taskmaster2/service1/internal/usecase/list_id"
)

func New() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/list", list.HTTPHandler)
	m.HandleFunc("/list/{id}", listid.HTTPHandler)
	m.HandleFunc("/create", create.HTTPHandler)
	return m
}
