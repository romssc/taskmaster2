package httprouter

import (
	"net/http"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"
)

func New(storage *inmemory.Storage, broker *kafkaa.Producer) http.Handler {
	create := create.New(storage, broker)
	list := list.New(storage)
	listid := listid.New(storage)
	m := http.NewServeMux()
	m.HandleFunc("/list", list.HTTPHandler)
	m.HandleFunc("/list/{id}", listid.HTTPHandler)
	m.HandleFunc("/create", create.HTTPHandler)
	return m
}
