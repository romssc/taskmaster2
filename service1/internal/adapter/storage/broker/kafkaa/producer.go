package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/pkg/json/standartjson"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	ErrClosingConnection = errors.New("kafka: failed to close connection")
	ErrMarshalingEvent   = errors.New("kafka: failed to prepare event")
	ErrProducingEvent    = errors.New("kafka: failed to produce event")
)

type Config struct {
	Address      []string      `yaml:"address"`
	Topic        string        `yaml:"topic"`
	BatchTimeout time.Duration `yaml:"batch_timeout"`
	RequiredAcks int           `yaml:"required_acks"`
}

type Producer struct {
	producer *kafka.Writer
}

func New(c Config) *Producer {
	return &Producer{
		producer: &kafka.Writer{
			Addr:         kafka.TCP(c.Address...),
			Topic:        c.Topic,
			BatchTimeout: c.BatchTimeout,
			RequiredAcks: kafka.RequiredAcks(c.RequiredAcks),
		},
	}
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosingConnection, err)
	}
	return nil
}

func (p *Producer) PublishEvent(ctx context.Context, event domain.Event) error {
	e, err := standartjson.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingEvent, err)
	}
	if err := p.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.ID),
		Value: e,
		Time:  time.Now(),
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrProducingEvent, err)
	}
	return nil
}
