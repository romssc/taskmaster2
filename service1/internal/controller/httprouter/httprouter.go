package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/pkg/id/uuidgen"
	"taskmaster2/service1/internal/pkg/json/standartjson"
	"taskmaster2/service1/internal/pkg/server/httpserver"
	"taskmaster2/service1/internal/pkg/timestamp/standarttime"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"
)

func New(c httpserver.Config, s *inmemory.Storage, b *kafkaa.Producer, g *uuidgen.Generator, t *standarttime.Time, d *standartjson.JSON) http.Handler {
	create := create.New(c.Routes.Create, s, b, g, t, d)
	list := list.New(c.Routes.List, s)
	listid := listid.New(c.Routes.ListID, s)
	m := http.NewServeMux()
	m.HandleFunc("/list", list.HTTPHandler)
	m.HandleFunc("/list/{id}", listid.HTTPHandler)
	m.HandleFunc("/create", create.HTTPHandler)
	return m
}
