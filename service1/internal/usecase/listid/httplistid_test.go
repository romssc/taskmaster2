package listid

import (
	"context"
	"testing"

	"service1/internal/adapter/storage/inmemory"
	"service1/internal/domain"

	"github.com/stretchr/testify/assert"
)

type mockGetter struct {
	record domain.Record
	err    error
}

func (m *mockGetter) GetTaskByID(ctx context.Context, id int) (domain.Record, error) {
	return m.record, m.err
}

func Test_GetTaskByID_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		id      int
		usecase *Usecase
		result  domain.Record
		err     error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      0,
			usecase: &Usecase{Getter: &mockGetter{}},
			result:  domain.Record{},
			err:     nil,
		},
		{
			name:    "record not found",
			ctx:     context.Background(),
			id:      0,
			usecase: &Usecase{Getter: &mockGetter{err: inmemory.ErrNotFound}},
			result:  domain.Record{},
			err:     ErrNotFound,
		},
		{
			name:    "storage failure",
			ctx:     context.Background(),
			id:      0,
			usecase: &Usecase{Getter: &mockGetter{err: inmemory.ErrExecuting}},
			result:  domain.Record{},
			err:     ErrDatabaseFailure,
		},
		{
			name:    "context canceled mid storage call",
			ctx:     context.Background(),
			id:      0,
			usecase: &Usecase{Getter: &mockGetter{err: inmemory.ErrOperationCanceled}},
			result:  domain.Record{},
			err:     ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			_, err := cs.usecase.GetTaskByID(cs.ctx, cs.id)
			assert.ErrorIs(t, err, cs.err)
		})
	}
}

func Test_validatePathValues_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		id     string
		result int
		err    error
	}{
		{
			name:   "success",
			id:     "0",
			result: 0,
			err:    nil,
		},
		{
			name:   "malformed id",
			id:     "xxx",
			result: 0,
			err:    ErrMalformedID,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			id, err := validatePathValues(cs.id)
			assert.ErrorIs(t, err, cs.err)
			assert.Equal(t, cs.result, id)
		})
	}
}
