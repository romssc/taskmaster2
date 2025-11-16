package update

import (
	"context"
	"testing"

	"service2/internal/domain"

	"github.com/stretchr/testify/assert"
)

func Test_Update_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		event   domain.Event
		usecase *Usecase
		cancel  bool
		err     error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			event:   domain.Event{},
			usecase: &Usecase{},
			cancel:  false,
			err:     nil,
		},
		{
			name:    "context closed mid work",
			ctx:     context.Background(),
			event:   domain.Event{},
			usecase: &Usecase{},
			cancel:  true,
			err:     ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			if !cs.cancel {
				err := cs.usecase.Update(cs.ctx, cs.event)
				assert.ErrorIs(t, err, cs.err)
			} else {
				ctx, cancel := context.WithCancel(cs.ctx)
				cancel()
				err := cs.usecase.Update(ctx, cs.event)
				assert.ErrorIs(t, err, cs.err)
			}
		})
	}
}
