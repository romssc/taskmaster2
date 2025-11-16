package list

import (
	"context"
	"testing"

	"service1/internal/adapter/storage/inmemory"
	"service1/internal/domain"

	"github.com/stretchr/testify/assert"
)

type mockGetter struct {
	records []domain.Record
	err     error
}

func (m *mockGetter) GetTasks(ctx context.Context) ([]domain.Record, error) {
	return m.records, m.err
}

func Test_GetTasks_Unit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ctx     context.Context
		usecase *Usecase
		result  []domain.Record
		err     error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			usecase: &Usecase{Getter: &mockGetter{records: []domain.Record{}}},
			result:  []domain.Record{},
			err:     nil,
		},
		{
			name:    "incompatible data",
			ctx:     context.Background(),
			usecase: &Usecase{Getter: &mockGetter{err: inmemory.ErrIncompatible}},
			result:  nil,
			err:     ErrDatabaseFailure,
		},
		{
			name:    "storage failure",
			ctx:     context.Background(),
			usecase: &Usecase{Getter: &mockGetter{err: inmemory.ErrExecuting}},
			result:  nil,
			err:     ErrDatabaseFailure,
		},
		{
			name:    "context closed mid storage call",
			ctx:     context.Background(),
			usecase: &Usecase{Getter: &mockGetter{err: inmemory.ErrOperationCanceled}},
			result:  nil,
			err:     ErrOperationCanceled,
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			records, err := cs.usecase.GetTasks(cs.ctx)
			assert.ErrorIs(t, err, cs.err)
			if cs.result != nil {
				assert.NotNil(t, records)
			} else {
				assert.Nil(t, records)
			}
		})
	}
}
