package create

import (
	"context"
	"errors"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPublisher struct {
	err error
}

func (m *mockPublisher) PublishEvent(ctx context.Context, event domain.Event) error {
	return m.err
}

type mockGenerator struct {
	id  int
	err error
}

func (m *mockGenerator) Gen() (int, error) {
	return m.id, m.err
}

func TestCreateTask_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		task    domain.Record
		usecase *Usecase
		result  int
		err     error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{id: 1, err: nil},
			},
			result: 1,
			err:    nil,
		},
		{
			name: "broker unavailable",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Publisher: &mockPublisher{err: kafkaa.ErrClosed},
				Generator: &mockGenerator{},
			},
			result: 0,
			err:    ErrBrokerUnavailable,
		},
		{
			name: "broker failure",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Publisher: &mockPublisher{err: errors.New("")},
				Generator: &mockGenerator{},
			},
			result: 0,
			err:    ErrBrokerFailure,
		},
		{
			name: "context closed mid kafka call",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Publisher: &mockPublisher{err: kafkaa.ErrOperationCanceled},
				Generator: &mockGenerator{},
			},
			result: 0,
			err:    ErrOperationCanceled,
		},
		{
			name: "generator failed to generate id",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{err: ErrGeneratingID},
			},
			result: 0,
			err:    ErrGeneratingID,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			id, err := cs.usecase.CreateTask(cs.ctx, cs.task)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result, id)
		})
	}
}

func TestValidTask_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		task domain.Record
		err  error
	}{
		{
			name: "success",
			task: domain.Record{Title: "Title"},
			err:  nil,
		},
		{
			name: "empty title",
			task: domain.Record{},
			err:  ErrEmptyTitle,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			err := validateTask(cs.task)
			assert.ErrorIs(t, err, cs.err)
		})
	}
}
