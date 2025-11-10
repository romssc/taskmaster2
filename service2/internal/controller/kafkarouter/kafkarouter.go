package kafkarouter

import (
	"taskmaster2/service2/internal/pkg/broker/kafkaa"
	"taskmaster2/service2/internal/usecase/update"
)

func New() kafkaa.Handlers {
	return kafkaa.Handlers{
		Update: update.New(),
	}
}
