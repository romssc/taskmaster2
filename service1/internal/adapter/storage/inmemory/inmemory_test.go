package inmemory

import (
	"context"
	"errors"
	inmemory "taskmaster2/in_memory"
	"taskmaster2/service1/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockKeeper struct {
	ac   mockAllContext
	c    mockClose
	cc   mockCreateContext
	lc   mockLoadContext
	uocc mockUpdateOrCreateContext
}

type mockAllContext struct {
	records []any
	err     error
}

type mockClose struct{}

type mockCreateContext struct {
	err error
}

type mockLoadContext struct {
	record any
	err    error
}

type mockUpdateOrCreateContext struct {
	err error
}

func (m *mockKeeper) AllContext(ctx context.Context) ([]any, error) {
	return m.ac.records, m.ac.err
}

func (m *mockKeeper) Close() {}

func (m *mockKeeper) CreateContext(ctx context.Context, key any, value any) error {
	return m.cc.err
}

func (m *mockKeeper) LoadContext(ctx context.Context, key any) (any, error) {
	return m.lc.record, m.lc.err
}

func (m *mockKeeper) UpdateOrCreateContext(ctx context.Context, key any, value any) error {
	return m.uocc.err
}

func Test_CreateTask_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		task    domain.Record
		storage *Storage
		result  int
		err     error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				cc: mockCreateContext{},
			}},
			result: 0,
			err:    nil,
		},
		{
			name: "already exists",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				cc: mockCreateContext{err: inmemory.ErrAlreadyExists},
			}},
			result: 0,
			err:    ErrAlreadyExists,
		},
		{
			name: "database failure",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				cc: mockCreateContext{err: errors.New("")},
			}},
			result: 0,
			err:    ErrExecuting,
		},
		{
			name: "context closed mid databse call",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				cc: mockCreateContext{err: inmemory.ErrOperationCanceled},
			}},
			result: 0,
			err:    ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			id, err := cs.storage.CreateTask(cs.ctx, cs.task)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result, id)
		})
	}
}

func Test_UpdateOrCreateTask_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		task    domain.Record
		storage *Storage
		err     error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				uocc: mockUpdateOrCreateContext{},
			}},
			err: nil,
		},
		{
			name: "database failure",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				uocc: mockUpdateOrCreateContext{err: errors.New("")},
			}},
			err: ErrExecuting,
		},
		{
			name: "context closed mid databse call",
			ctx:  context.Background(),
			task: domain.Record{
				ID: 0,
			},
			storage: &Storage{store: &mockKeeper{
				uocc: mockUpdateOrCreateContext{err: inmemory.ErrOperationCanceled},
			}},
			err: ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			err := cs.storage.UpdateOrCreateTask(cs.ctx, cs.task)
			assert.ErrorIs(t, err, cs.err)
		})
	}
}

func Test_GetTaskByID_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		id      int
		storage *Storage
		result  domain.Record
		err     error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   0,
			storage: &Storage{store: &mockKeeper{
				lc: mockLoadContext{record: any(domain.Record{})},
			}},
			result: domain.Record{},
			err:    nil,
		},
		{
			name: "not found",
			ctx:  context.Background(),
			id:   0,
			storage: &Storage{store: &mockKeeper{
				lc: mockLoadContext{err: inmemory.ErrNotFound},
			}},
			result: domain.Record{},
			err:    ErrNotFound,
		},
		{
			name: "database failure",
			ctx:  context.Background(),
			id:   0,
			storage: &Storage{store: &mockKeeper{
				lc: mockLoadContext{err: errors.New("")},
			}},
			result: domain.Record{},
			err:    ErrExecuting,
		},
		{
			name: "incompatible data",
			ctx:  context.Background(),
			id:   0,
			storage: &Storage{store: &mockKeeper{
				lc: mockLoadContext{},
			}},
			result: domain.Record{},
			err:    ErrIncompatible,
		},
		{
			name: "context closed mid databse call",
			ctx:  context.Background(),
			id:   0,
			storage: &Storage{store: &mockKeeper{
				lc: mockLoadContext{err: inmemory.ErrOperationCanceled},
			}},
			result: domain.Record{},
			err:    ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			id, err := cs.storage.GetTaskByID(cs.ctx, cs.id)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result, id)
		})
	}
}

func Test_GetTasks_Unit(t *testing.T) {
	t.Parallel()
	var a []any
	r := any(domain.Record{})
	a = append(a, r)
	cases := []struct {
		name    string
		ctx     context.Context
		storage *Storage
		result  []domain.Record
		err     error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			storage: &Storage{store: &mockKeeper{
				ac: mockAllContext{records: a},
			}},
			result: []domain.Record{},
			err:    nil,
		},
		{
			name: "database failure",
			ctx:  context.Background(),
			storage: &Storage{store: &mockKeeper{
				ac: mockAllContext{err: errors.New("")},
			}},
			result: nil,
			err:    ErrExecuting,
		},
		{
			name: "incompatible data",
			ctx:  context.Background(),
			storage: &Storage{store: &mockKeeper{
				ac: mockAllContext{records: []any{""}},
			}},
			result: nil,
			err:    ErrIncompatible,
		},
		{
			name: "context closed mid databse call",
			ctx:  context.Background(),
			storage: &Storage{store: &mockKeeper{
				ac: mockAllContext{err: inmemory.ErrOperationCanceled},
			}},
			result: nil,
			err:    ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			records, err := cs.storage.GetTasks(cs.ctx)
			assert.ErrorIs(t, err, cs.err)
			if cs.result != nil {
				assert.NotNil(t, records)
			} else {
				assert.Nil(t, records)
			}
		})
	}
}
