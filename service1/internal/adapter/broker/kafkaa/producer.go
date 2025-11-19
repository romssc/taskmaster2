package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"time"

	"service1/internal/domain"

	"github.com/segmentio/kafka-go"
)

var (
	ErrOperationCanceled = errors.New("kafka: operation canceled, no events written")

	ErrClosingConnection = errors.New("kafka: failed to close connection")
	ErrMarshalingEvent   = errors.New("kafka: failed to prepare event")
	ErrProducingEvent    = errors.New("kafka: failed to produce event")
	ErrClosed            = errors.New("kafka: failed due to closed broker")
)

type Config struct {
	Address            []string
	Topic              string        `yaml:"topic"`
	BatchTimeout       time.Duration `yaml:"batch_timeout"`
	RequiredAcks       int           `yaml:"required_acks"`
	AllowTopicCreation bool          `yaml:"allow_topic_creation"`

	Encoder Encoder
}

type Publisher interface {
	Close() error
	Stats() kafka.WriterStats
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type Encoder interface {
	Marshal(data any) ([]byte, error)
}

type Producer struct {
	producer Publisher

	encoder Encoder
}

func New(c Config) *Producer {
	return &Producer{
		producer: &kafka.Writer{
			Addr:                   kafka.TCP(c.Address...),
			Topic:                  c.Topic,
			BatchTimeout:           c.BatchTimeout,
			RequiredAcks:           kafka.RequiredAcks(c.RequiredAcks),
			AllowAutoTopicCreation: c.AllowTopicCreation,
		},

		encoder: c.Encoder,
	}
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosingConnection, err)
	}
	return nil
}

func (p *Producer) PublishEvent(ctx context.Context, event domain.Event) error {
	eventByte, marshalErr := p.encoder.Marshal(event)
	if marshalErr != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingEvent, marshalErr)
	}
	if writeErr := p.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(domain.ActionUpdate),
		Value: eventByte,
		Time:  time.Now(),
	}); writeErr != nil {
		switch {
		case errors.Is(writeErr, context.Canceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, writeErr)
		case errors.Is(writeErr, kafka.ErrGroupClosed):
			return fmt.Errorf("%w: %v", ErrClosed, writeErr)
		default:
			return fmt.Errorf("%w: %v", ErrProducingEvent, writeErr)
		}
	}
	return nil
}
