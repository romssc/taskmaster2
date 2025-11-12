package kafkarouter

import (
	"taskmaster2/service2/internal/pkg/broker/kafkaa"
	"taskmaster2/service2/internal/usecase/update"
)

type Config struct {
	Update *update.Usecase
}

func New(c *Config) kafkaa.Handlers {
	return kafkaa.Handlers{
		Update: c.Update,
	}
}
