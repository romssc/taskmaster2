package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	listid "taskmaster2/service1/internal/usecase/list_id"
)

func New() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/list", list.Handler)
	m.HandleFunc("/list/{id}", listid.Handler)
	m.HandleFunc("/create", create.Handler)
	return m
}
