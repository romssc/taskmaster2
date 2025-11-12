package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"log"
	"taskmaster2/service2/internal/controller/kafkarouter"
	"taskmaster2/service2/internal/domain"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	ErrFetchingMessages    = errors.New("kafka: failed while fetching new messages")
	ErrUnmarshalingMessage = errors.New("kafka: failed while unmarshaling message")
	ErrConsumerClosed      = errors.New("kafka: failed due to consumer being closed")
	ErrProcessingMessage   = errors.New("kafka: failed to process message")
	ErrClosing             = errors.New("kafka: failed to close")
	ErrUnknownAction       = errors.New("kafka: failed to process message: unknown action")
	ErrCommitting          = errors.New("kafka: failed to commit")
	ErrOperationCanceled   = errors.New("kafka: operation canceled")
)

type Config struct {
	Brokers        []string      `yaml:"brokers"`
	Topic          string        `yaml:"topic"`
	GroupID        string        `yaml:"group_id"`
	CommitInterval time.Duration `yaml:"commit_interval"`
	SessionTimeout time.Duration `yaml:"session_timeout"`
	StartOffset    int           `yaml:"start_offset"`

	Handler Handler

	Encoder Encoder
	Decoder Decoder
}

type Consumer struct {
	reader     Reader
	proccessor Proccessor
}

type Reader interface {
	Close() error
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Config() kafka.ReaderConfig
	FetchMessage(ctx context.Context) (kafka.Message, error)
	Lag() int64
	Offset() int64
	ReadLag(ctx context.Context) (lag int64, err error)
	ReadMessage(ctx context.Context) (kafka.Message, error)
	SetOffset(offset int64) error
	SetOffsetAt(ctx context.Context, t time.Time) error
	Stats() kafka.ReaderStats
}

func New(c Config) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.Brokers,
		Topic:          c.Topic,
		GroupID:        c.GroupID,
		CommitInterval: c.CommitInterval,
		SessionTimeout: c.SessionTimeout,
		StartOffset:    int64(c.StartOffset),
	})
	return &Consumer{
		reader: reader,
		proccessor: &Proccess{
			reader:  reader,
			handler: c.Handler,
			encoder: c.Encoder,
			decoder: c.Decoder,
		},
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", ErrConsumerClosed, ctx.Err())
		default:
			if err := c.proccessor.Process(ctx); err != nil {
				switch {
				case errors.Is(err, ErrConsumerClosed):
					return fmt.Errorf("%w: %v", ErrConsumerClosed, ctx.Err())
				default:
					log.Println(err)
				}
			}
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosing, err)
	}
	return nil
}

type Proccessor interface {
	Process(ctx context.Context) error
}

type Proccess struct {
	reader Reader

	handler Handler

	encoder Encoder
	decoder Decoder
}

type Handler interface {
	Pick(ctx context.Context, action domain.Action, event domain.Event) error
}

type Encoder interface {
	Marshal(data any) ([]byte, error)
}

type Decoder interface {
	Unmarshal(data []byte, v any) error
}

func (p *Proccess) Process(ctx context.Context) error {
	msg, fetchErr := p.reader.FetchMessage(ctx)
	if fetchErr != nil {
		switch {
		case errors.Is(fetchErr, context.Canceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		case errors.Is(fetchErr, kafka.ErrGroupClosed):
			return fmt.Errorf("%w: %v", ErrConsumerClosed, fetchErr)
		default:
			return fmt.Errorf("%w: %v", ErrFetchingMessages, fetchErr)
		}
	}
	var event domain.Event
	if umErr := p.decoder.Unmarshal(msg.Value, &event); umErr != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshalingMessage, umErr)
	}
	pickErr := p.handler.Pick(ctx, event.Action, event)
	if pickErr != nil {
		switch {
		case errors.Is(pickErr, kafkarouter.ErrOperationCanceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		default:
			return fmt.Errorf("%w: %v", ErrProcessingMessage, pickErr)
		}
	}
	if commitErr := p.reader.CommitMessages(ctx, msg); commitErr != nil {
		switch {
		case errors.Is(commitErr, context.Canceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		case errors.Is(commitErr, kafka.ErrGroupClosed):
			return fmt.Errorf("%w: %v", ErrConsumerClosed, commitErr)
		default:
			return fmt.Errorf("%w: %v", ErrCommitting, commitErr)
		}
	}
	return nil
}
