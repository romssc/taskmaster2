package httprouter

import (
	"net/http"

	"service1/internal/usecase/create"
	"service1/internal/usecase/list"
	"service1/internal/usecase/listid"
)

type Config struct {
	Create create.Config `mapstructure:"create"`
	List   list.Config   `mapstructure:"list"`
	ListID listid.Config `mapstructure:"list_id"`
}

type Routes struct {
	Create *create.Usecase
	List   *list.Usecase
	ListID *listid.Usecase
}

func New(r *Routes) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/list", r.List.HTTPHandler)
	m.HandleFunc("/list/{id}", r.ListID.HTTPHandler)
	m.HandleFunc("/create", r.Create.HTTPHandler)
	return m
}
