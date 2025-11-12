package inmemory

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrOperationCanceled = errors.New("memory: operation canceled")
	ErrNotFound          = errors.New("memory: not found: no value for given key")
	ErrAlreadyExists     = errors.New("memory: exists: the value already exists for the given key")
)

type InMemory struct {
	store *sync.Map
}

func New() *InMemory {
	return &InMemory{
		store: &sync.Map{},
	}
}

func (m *InMemory) Close() {
	m.store.Clear()
}

func (m *InMemory) CreateContext(ctx context.Context, key any, value any) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w: %v", ErrOperationCanceled, ctx.Err())
	default:
		_, ok := m.store.Load(key)
		if ok {
			return fmt.Errorf("%w", ErrAlreadyExists)
		}
		m.store.Store(key, value)
		return nil
	}
}

func (m *InMemory) UpdateOrCreateContext(ctx context.Context, key any, value any) error {
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
		return nil, fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
	default:
		value, ok := m.store.Load(key)
		if !ok {
			return nil, fmt.Errorf("%w", ErrNotFound)
		}
		return value, nil
	}
}

func (m *InMemory) AllContext(ctx context.Context) ([]any, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
	default:
		var pairs []any
		var counter uint64
		m.store.Range(func(key, value any) bool {
			counter++
			if counter%50 == 0 {
				select {
				case <-ctx.Done():
					return false
				default:
				}
			}
			pairs = append(pairs, value)
			return true
		})
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
		}
		return pairs, nil
	}
}
