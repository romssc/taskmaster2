package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"taskmaster2/service1/internal/domain"
	"time"

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
}

type Encoder interface {
	Marshal(data any) ([]byte, error)
}

type Producer struct {
	producer *kafka.Writer

	encoder Encoder
}

func New(c Config, e Encoder) *Producer {
	return &Producer{
		producer: &kafka.Writer{
			Addr:                   kafka.TCP(c.Address...),
			Topic:                  c.Topic,
			BatchTimeout:           c.BatchTimeout,
			RequiredAcks:           kafka.RequiredAcks(c.RequiredAcks),
			AllowAutoTopicCreation: c.AllowTopicCreation,
		},

		encoder: e,
	}
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosingConnection, err)
	}
	return nil
}

func (p *Producer) PublishEvent(ctx context.Context, event domain.Event) error {
	e, err := p.encoder.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMarshalingEvent, err)
	}
	if err := p.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(event.Record.ID)),
		Value: e,
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
