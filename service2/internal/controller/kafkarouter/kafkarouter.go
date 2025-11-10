package kafkarouter

import (
	"taskmaster2/service2/internal/adapter/storage/inmemory"
	"taskmaster2/service2/internal/pkg/broker/kafkaa"
	"taskmaster2/service2/internal/usecase/update"
)

func New(storage *inmemory.Storage) kafkaa.Handlers {
	return kafkaa.Handlers{
		Update: update.New(storage),
	}
}
