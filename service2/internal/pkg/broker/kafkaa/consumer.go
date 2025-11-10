package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"log"
	"taskmaster2/service2/internal/domain"
	"taskmaster2/service2/internal/pkg/json/standartjson"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	ErrFetchingMessages    = errors.New("kafka: failed while fetching new messages")
	ErrUnmarshalingMessage = errors.New("kafka: failed while unmarshaling message")
	ErrConsumerClose       = errors.New("kafka: failed due to consumer being closed")
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
}

type Consumer struct {
	reader  *kafka.Reader
	handler Handlers
}

type EventHandler func(ctx context.Context, event domain.Event) error

type Handlers struct {
	Update EventHandler
}

func New(h Handlers, c Config) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        c.Brokers,
			Topic:          c.Topic,
			GroupID:        c.GroupID,
			CommitInterval: c.CommitInterval,
			SessionTimeout: c.SessionTimeout,
			StartOffset:    int64(c.StartOffset),
		}),
		handler: h,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", ErrConsumerClose, ctx.Err())
		default:
			if err := c.process(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return fmt.Errorf("%w: %v", ErrConsumerClose, ctx.Err())
				}
				log.Println(err)
			}
		}
	}
}

func (c *Consumer) process(ctx context.Context) error {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchingMessages, err)
	}
	var event domain.Event
	if err := standartjson.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshalingMessage, err)
	}
	var e error
	switch event.Action {
	case domain.ActionUpdate:
		e = c.handler.Update(ctx, event)
	default:
		c.reader.CommitMessages(ctx, msg)
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
