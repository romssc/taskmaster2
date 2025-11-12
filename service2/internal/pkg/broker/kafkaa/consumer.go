package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"log"
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
)

type Config struct {
	Brokers        []string      `yaml:"brokers"`
	Topic          string        `yaml:"topic"`
	GroupID        string        `yaml:"group_id"`
	CommitInterval time.Duration `yaml:"commit_interval"`
	SessionTimeout time.Duration `yaml:"session_timeout"`
	StartOffset    int           `yaml:"start_offset"`

	Handler Handlers

	Encoder Encoder
	Decoder Decoder
}

type Consumer struct {
	reader *kafka.Reader

	handlers Handlers

	Encoder Encoder
	Decoder Decoder
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

type Handler interface {
	EventHandler(ctx context.Context, event domain.Event) error
}

type Handlers struct {
	Update Handler
}

type Encoder interface {
	Marshal(data any) ([]byte, error)
}

type Decoder interface {
	Unmarshal(data []byte, v any) error
}

func New(c Config) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        c.Brokers,
			Topic:          c.Topic,
			GroupID:        c.GroupID,
			CommitInterval: c.CommitInterval,
			SessionTimeout: c.SessionTimeout,
			StartOffset:    int64(c.StartOffset),
		}),
		handlers: c.Handler,
		Encoder:  c.Encoder,
		Decoder:  c.Decoder,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", ErrConsumerClosed, ctx.Err())
		default:
			if err := c.process(ctx); err != nil {
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

func (c *Consumer) process(ctx context.Context) error {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return fmt.Errorf("%w: %v", ErrConsumerClosed, ctx.Err())
		default:
			return fmt.Errorf("%w: %v", ErrFetchingMessages, err)
		}
	}
	var event domain.Event
	if err := c.Decoder.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshalingMessage, err)
	}
	var e error
	switch event.Action {
	case domain.ActionUpdate:
		e = c.handlers.Update.EventHandler(ctx, event)
	default:
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("%w: %v", ErrCommitting, err)
		}
		return fmt.Errorf("%w", ErrUnknownAction)
	}
	if e != nil {
		return fmt.Errorf("%w: %v", ErrProcessingMessage, e)
	}
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("%w: %v", ErrCommitting, err)
	}
	return nil
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosing, err)
	}
	return nil
}
