package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrOperationCanceled = errors.New("memory: operation canceled")
	ErrNotFound          = errors.New("memory: not found")
)

type InMemory struct {
	store sync.Map
}

func New() *InMemory {
	return &InMemory{
		store: sync.Map{},
	}
}

func (m *InMemory) CreateContext(ctx context.Context, key any, value any) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
	default:
		m.store.Store(key, value)
		return nil
	}
}

func (m *InMemory) LoadContext(ctx context.Context, key any) (any, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
	default:
		value, ok := m.store.Load(key)
		if !ok {
			return nil, fmt.Errorf("%w: no value for given key", ErrNotFound)
		}
		return value, nil
	}
}
