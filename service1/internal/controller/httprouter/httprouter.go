package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/pkg/idgen/gen"
	"taskmaster2/service1/internal/pkg/server/httpserver"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"
)

func New(c httpserver.Config, s *inmemory.Storage, b *kafkaa.Producer, g *gen.Generator) http.Handler {
	create := create.New(c.Routes.Create, s, b, g)
	list := list.New(s)
	listid := listid.New(s)
	m := http.NewServeMux()
	m.HandleFunc("/list", list.HTTPHandler)
	m.HandleFunc("/list/{id}", listid.HTTPHandler)
	m.HandleFunc("/create", create.HTTPHandler)
	return m
}
