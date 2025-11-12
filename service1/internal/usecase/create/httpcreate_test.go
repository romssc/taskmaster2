package create

import (
	"context"
	"errors"
	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/domain"
	"taskmaster2/service1/internal/pkg/id/uuidgen"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockCreator struct {
	ct   mockCreateTask
	uoct mockUpdateOrCreateTask
}

type mockCreateTask struct {
	id  int
	err error
}

type mockUpdateOrCreateTask struct {
	err error
}

func (m *mockCreator) CreateTask(ctx context.Context, task domain.Record) (int, error) {
	return m.ct.id, m.ct.err
}

func (m *mockCreator) UpdateOrCreateTask(ctx context.Context, task domain.Record) error {
	return m.uoct.err
}

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

type mockTimer struct {
	time int64
}

func (m *mockTimer) TimeNow() int64 {
	return m.time
}

func Test_CreateTask_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		task    domain.Record
		usecase *Usecase
		result  domain.Event
		err     error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{},
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{
				Action: domain.ActionUpdate,
				Record: domain.Record{
					Status: domain.StatusNew,
				},
			},
			err: nil,
		},
		{
			name: "failed to generate id",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{},
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{err: ErrGeneratingID},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrGeneratingID,
		},
		{
			name: "record already exists",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{ct: mockCreateTask{err: inmemory.ErrAlreadyExists}},
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrStorageAlreadyExists,
		},
		{
			name: "storage failure",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{ct: mockCreateTask{err: inmemory.ErrExecuting}},
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrStorageFailure,
		},
		{
			name: "context closed mid storage call",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{ct: mockCreateTask{err: inmemory.ErrOperationCanceled}},
				Publisher: &mockPublisher{},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrOperationCanceled,
		},
		{
			name: "broker unavailable",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{},
				Publisher: &mockPublisher{err: kafkaa.ErrClosed},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrBrokerUnavailable,
		},
		{
			name: "broker failure",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{},
				Publisher: &mockPublisher{err: errors.New("")},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrBrokerFailure,
		},
		{
			name: "context closed mid kafka call",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{},
				Publisher: &mockPublisher{err: kafkaa.ErrOperationCanceled},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrOperationCanceled,
		},
		{
			name: "failed to mark failure",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{uoct: mockUpdateOrCreateTask{err: inmemory.ErrExecuting}},
				Publisher: &mockPublisher{err: kafkaa.ErrProducingEvent},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrStorageFailure,
		},
		{
			name: "context closed mid marking failure",
			ctx:  context.Background(),
			task: domain.Record{},
			usecase: &Usecase{
				Creator:   &mockCreator{uoct: mockUpdateOrCreateTask{err: inmemory.ErrOperationCanceled}},
				Publisher: &mockPublisher{err: kafkaa.ErrProducingEvent},
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrOperationCanceled,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			event, err := cs.usecase.CreateTask(cs.ctx, cs.task)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result.Action, event.Action)
			assert.Equal(t, cs.result.Record.Status, event.Record.Status)
		})
	}
}

func Test_createEvent_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		task    domain.Record
		usecase *Usecase
		result  domain.Event
		err     error
	}{
		{
			name: "success",
			task: domain.Record{},
			usecase: &Usecase{
				Generator: &mockGenerator{},
				Timer:     &mockTimer{},
			},
			result: domain.Event{
				Action: domain.ActionUpdate,
				Record: domain.Record{
					Status: domain.StatusNew,
				},
			},
			err: nil,
		},
		{
			name: "failed to generate id",
			task: domain.Record{},
			usecase: &Usecase{
				Generator: &mockGenerator{err: uuidgen.ErrGenerating},
				Timer:     &mockTimer{},
			},
			result: domain.Event{},
			err:    ErrGeneratingID,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			event, err := cs.usecase.createEvent(cs.task)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result.Action, event.Action)
			assert.Equal(t, cs.result.Record.Status, event.Record.Status)
		})
	}
}

func Test_markFailure_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		task    domain.Record
		usecase *Usecase
		result  domain.Record
		err     error
	}{
		{
			name: "success",
			task: domain.Record{},
			usecase: &Usecase{
				Creator: &mockCreator{uoct: mockUpdateOrCreateTask{}},
			},
			result: domain.Record{Status: domain.StatusFailed},
			err:    nil,
		},
		{
			name: "context closed mid storage call",
			task: domain.Record{},
			usecase: &Usecase{
				Creator: &mockCreator{uoct: mockUpdateOrCreateTask{err: inmemory.ErrOperationCanceled}},
			},
			result: domain.Record{},
			err:    ErrOperationCanceled,
		},
		{
			name: "storage failure",
			task: domain.Record{},
			usecase: &Usecase{
				Creator: &mockCreator{uoct: mockUpdateOrCreateTask{err: inmemory.ErrExecuting}},
			},
			result: domain.Record{},
			err:    ErrStorageFailure,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			record, err := cs.usecase.markFailure(cs.ctx, cs.task)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result.Status, record.Status)
		})
	}
}

func Test_validateTask_Unit(t *testing.T) {
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
