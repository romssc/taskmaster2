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
	Address            []string      `mapstructure:"address"`
	Topic              string        `mapstructure:"topic"`
	BatchTimeout       time.Duration `mapstructure:"batch_timeout"`
	RequiredAcks       int           `mapstructure:"required_acks"`
	AllowTopicCreation bool          `mapstructure:"allow_topic_creation"`

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
	eventByte, err := p.encoder.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingEvent, err)
	}
	if err := p.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(domain.ActionUpdate),
		Value: eventByte,
		Time:  time.Now(),
	}); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		case errors.Is(err, kafka.ErrGroupClosed):
			return fmt.Errorf("%w: %v", ErrClosed, err)
		default:
			return fmt.Errorf("%w: %v", ErrProducingEvent, err)
		}
	}
	return nil
}
