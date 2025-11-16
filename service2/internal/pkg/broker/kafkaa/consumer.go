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
	"golang.org/x/sync/errgroup"
)

var (
	ErrFetchingMessages    = errors.New("kafka: failed while fetching new messages")
	ErrUnmarshalingMessage = errors.New("kafka: failed while unmarshaling message")
	ErrConsumerClosed      = errors.New("kafka: failed due to consumer being closed")
	ErrProcessingMessage   = errors.New("kafka: failed to process message")
	ErrClosingConsumer     = errors.New("kafka: failed to close consumer")
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

	WorkerCount int `yaml:"worker_count"`

	Handler Handler

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
	Pick(ctx context.Context, action domain.Action, event domain.Event) error
}

type Encoder interface {
	Marshal(data any) ([]byte, error)
}

type Decoder interface {
	Unmarshal(data []byte, v any) error
}

type Consumer struct {
	config Config

	reader  Reader
	handler Handler

	encoder Encoder
	decoder Decoder
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
		config:  c,
		reader:  reader,
		handler: c.Handler,
		encoder: c.Encoder,
		decoder: c.Decoder,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	jobs := make(chan kafka.Message, c.config.WorkerCount*2)
	eg, egCtx := errgroup.WithContext(ctx)
	for w := 0; w < c.config.WorkerCount; w++ {
		eg.Go(func() error {
			for message := range jobs {
				if prErr := c.proccess(egCtx, message); prErr != nil && !errors.Is(prErr, ErrOperationCanceled) {
					log.Println(prErr)
				}
			}
			return nil
		})
	}
	eg.Go(func() error {
		defer close(jobs)
		for {
			select {
			case <-egCtx.Done():
				return nil
			default:
				message, err := c.fetch(egCtx)
				if err != nil && !errors.Is(err, ErrOperationCanceled) {
					log.Println(err)
					continue
				}
				select {
				case jobs <- message:
				case <-egCtx.Done():
					return nil
				}
			}
		}
	})
	return eg.Wait()
}

func (c *Consumer) proccess(ctx context.Context, message kafka.Message) error {
	var event domain.Event
	if umErr := c.decoder.Unmarshal(message.Value, &event); umErr != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshalingMessage, umErr)
	}
	pickErr := c.handler.Pick(ctx, event.Action, event)
	if pickErr != nil {
		switch {
		case errors.Is(pickErr, kafkarouter.ErrOperationCanceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		default:
			return fmt.Errorf("%w: %v", ErrProcessingMessage, pickErr)
		}
	}
	return nil
}

func (c *Consumer) fetch(ctx context.Context) (kafka.Message, error) {
	message, fetchErr := c.reader.FetchMessage(ctx)
	if fetchErr != nil {
		switch {
		case errors.Is(fetchErr, context.Canceled):
			return kafka.Message{}, fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		case errors.Is(fetchErr, kafka.ErrGroupClosed):
			return kafka.Message{}, fmt.Errorf("%w: %v", ErrConsumerClosed, fetchErr)
		default:
			return kafka.Message{}, fmt.Errorf("%w: %v", ErrFetchingMessages, fetchErr)
		}
	}
	if commitErr := c.reader.CommitMessages(ctx, message); commitErr != nil {
		switch {
		case errors.Is(commitErr, context.Canceled):
			return kafka.Message{}, fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		case errors.Is(commitErr, kafka.ErrGroupClosed):
			return kafka.Message{}, fmt.Errorf("%w: %v", ErrConsumerClosed, commitErr)
		default:
			return kafka.Message{}, fmt.Errorf("%w: %v", ErrCommitting, commitErr)
		}
	}
	return message, nil
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		switch {
		case errors.Is(err, kafka.ErrGroupClosed):
			return fmt.Errorf("%w: %v", ErrConsumerClosed, err)
		default:
			return fmt.Errorf("%w: %v", ErrClosingConsumer, err)
		}
	}
	return nil
}
