package update

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"service2/internal/domain"

	"github.com/segmentio/kafka-go"
)

var (
	ErrOperationCanceled = errors.New("update: operation canceled")

	ErrUnmarshalingMessage = errors.New("update: failed while unmarshaling message")
)

type Config struct{}

type Decoder interface {
	Unmarshal(data []byte, v any) error
}

type Usecase struct {
	Config Config

	Decoder Decoder
}

func (u *Usecase) EventHandler(ctx context.Context, message kafka.Message) {
	var event domain.Event
	if umErr := u.Decoder.Unmarshal(message.Value, &event); umErr != nil {
		log.Println(fmt.Errorf("%v: %v", ErrUnmarshalingMessage, umErr))
		return
	}
	if upErr := u.Update(ctx, event); upErr != nil && !errors.Is(upErr, ErrOperationCanceled) {
		log.Println(upErr)
		return
	}
}

func (u *Usecase) Update(ctx context.Context, event domain.Event) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
	case <-time.After(time.Second * 7):
		log.Println(event)
	}
	return nil
}
