package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"service1/internal/domain"

	"github.com/segmentio/kafka-go"
)

var (
	ErrClosingConnection = errors.New("kafka: failed to close connection")
	ErrMarshalingEvent   = errors.New("kafka: failed to prepare event")
	ErrProducingEvent    = errors.New("kafka: failed to produce event")
	ErrClosed            = errors.New("kafka: failed due to closed broker")
	ErrOperationCanceled = errors.New("kafka: operation canceled, no events written")
)

type Config struct {
	Address            []string      `yaml:"address"`
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
	e, mrErr := p.encoder.Marshal(event)
	if mrErr != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingEvent, mrErr)
	}
	if wmErr := p.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(event.Record.ID)),
		Value: e,
		Time:  time.Now(),
	}); wmErr != nil {
		switch {
		case errors.Is(wmErr, context.Canceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, wmErr)
		case errors.Is(wmErr, kafka.ErrGroupClosed):
			return fmt.Errorf("%w: %v", ErrClosed, wmErr)
		default:
			return fmt.Errorf("%w: %v", ErrProducingEvent, wmErr)
		}
	}
	return nil
}
