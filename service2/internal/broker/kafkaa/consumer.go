package kafkaa

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var (
	ErrOperationCanceled = errors.New("kafka: operation canceled")

	ErrNotStarted       = errors.New("kafka: consumer is not started")
	ErrFetchingMessages = errors.New("kafka: failed while trying to fetch new messages")
	ErrCommitting       = errors.New("kafka: failed to commit offset")
	ErrClosingConsumer  = errors.New("kafka: failed to close consumer")
	ErrTooManyRetries   = errors.New("kafka: too many retries")
)

type Config struct {
	Brokers        []string      `mapstructure:"brokers"`
	Topic          string        `mapstructure:"topic"`
	GroupID        string        `mapstructure:"group_id"`
	CommitInterval time.Duration `mapstructure:"commit_interval"`
	SessionTimeout time.Duration `mapstructure:"session_timeout"`
	StartOffset    int           `mapstructure:"start_offset"`

	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	RetryAmount     int           `mapstructure:"retry_amount"`
	WorkerCount     int           `mapstructure:"worker_count"`
	JobsMultiplier  int           `mapstructure:"jobs_multiplier"`

	Handler Handler
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
	Route(ctx context.Context, message kafka.Message)
}

type Consumer struct {
	config Config

	reader  Reader
	handler Handler
}

func New(c Config) *Consumer {
	if c.WorkerCount < 1 {
		c.WorkerCount = 1
	}
	if c.JobsMultiplier < 1 {
		c.JobsMultiplier = 1
	}
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
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	jobs := make(chan kafka.Message, c.config.WorkerCount*c.config.JobsMultiplier)
	done := make(chan struct{})

	var wg sync.WaitGroup
	workerCtx, workerCancel := context.WithCancel(context.Background())

	defer func() {
		if jobs != nil {
			close(jobs)
		}
		select {
		case <-time.After(c.config.ShutdownTimeout):
			workerCancel()
		case <-done:
			// WORKERS DRAINED
		}
	}()

	go func() {
		defer func() {
			wg.Wait()
			close(done)
		}()
		for w := 0; w < c.config.WorkerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for message := range jobs {
					c.handler.Route(workerCtx, message)
				}
			}()
		}
	}()

	backoff := time.Second * 0
	for a := 0; a <= c.config.RetryAmount; a++ {
		if err := c.consume(ctx, jobs, backoff); err != nil {
			if errors.Is(err, ErrOperationCanceled) {
				return err
			}
			if a == c.config.RetryAmount {
				return fmt.Errorf("%w: %v", ErrTooManyRetries, err)
			}
			switch backoff {
			case 0:
				backoff = time.Second * 2
			default:
				backoff *= 2
			}
			log.Println(err)
			continue
		}
		a = 0
		backoff = 0
	}

	return nil
}

func (c *Consumer) consume(ctx context.Context, jobs chan kafka.Message, backoff time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
	case <-time.After(backoff):
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("%w: %v", ErrOperationCanceled, err)
			default:
				return fmt.Errorf("%w: %v", ErrFetchingMessages, err)
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
		case jobs <- message:
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				switch {
				case errors.Is(err, context.Canceled):
					return fmt.Errorf("%w: %v", ErrOperationCanceled, err)
				default:
					return fmt.Errorf("%w: %v", ErrCommitting, err)
				}
			}
		}
		return nil
	}
}

func (c *Consumer) Shutdown() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrClosingConsumer, err)
	}
	return nil
}
