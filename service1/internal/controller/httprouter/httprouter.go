package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"
)

type Config struct {
	Create *create.Usecase
	List   *list.Usecase
	ListID *listid.Usecase
}

func New(c *Config) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/list", c.List.HTTPHandler)
	m.HandleFunc("/list/{id}", c.ListID.HTTPHandler)
	m.HandleFunc("/create", c.Create.HTTPHandler)
	return m
}
