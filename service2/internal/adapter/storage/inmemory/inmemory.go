package inmemory

import (
	"context"
	"errors"
	"fmt"
	inmemory "taskmaster2/in_memory"
	"taskmaster2/service2/internal/domain"
)

var (
	ErrExecuting    = errors.New("inmemory: failed to execute")
	ErrIncompatible = errors.New("inmemory: data incompatible: memory stores different type")
	ErrNotFound     = errors.New("inmemory: no records found")
)

type Storage struct {
	store *inmemory.InMemory
}

func New() *Storage {
	return &Storage{
		store: inmemory.New(),
	}
}

func (s *Storage) Close() {
	s.store.Close()
}

func (s *Storage) UpdateStatus(ctx context.Context, id int, status domain.Status) error {
	record, err := s.store.LoadContext(ctx, id)
	if err != nil {
		if errors.Is(err, inmemory.ErrNotFound) {
			return fmt.Errorf("%w: %v", ErrNotFound, err)
		}
		return fmt.Errorf("%v: %w", ErrExecuting, err)
	}
	task, ok := record.(domain.Record)
	if !ok {
		return fmt.Errorf("%w", ErrIncompatible)
	}
	task.Status = status
	if err := s.store.UpdateContext(ctx, id, task); err != nil {
		return fmt.Errorf("%v: %w", ErrExecuting, err)
	}
	return nil
}
