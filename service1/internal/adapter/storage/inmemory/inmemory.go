package inmemory

import (
	"context"
	"errors"
	"fmt"
	"taskmaster2/service1/internal/domain"
	inmemory "taskmaster2/service1/pkg/in_memory"
)

var (
	ErrOperationCanceled = errors.New("inmemory: operation canceled, no records written")
	ErrAlreadyExists     = errors.New("inmemory: already exists")
	ErrExecuting         = errors.New("inmemory: failed to execute")
	ErrIncompatible      = errors.New("inmemory: data incompatible: memory stores different type")
	ErrNotFound          = errors.New("inmemory: no records found")
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

func (s *Storage) CreateTask(ctx context.Context, task domain.Record) (int, error) {
	if err := s.store.CreateContext(ctx, task.ID, task); err != nil {
		switch {
		case errors.Is(err, inmemory.ErrOperationCanceled):
			return 0, fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		case errors.Is(err, inmemory.ErrAlreadyExists):
			return 0, fmt.Errorf("%w: %v", ErrAlreadyExists, err)
		default:
			return 0, fmt.Errorf("%v: %w", ErrExecuting, err)
		}
	}
	return task.ID, nil
}

func (s *Storage) UpdateOrCreateTask(ctx context.Context, task domain.Record) error {
	if err := s.store.UpdateOrCreateContext(ctx, task.ID, task); err != nil {
		switch {
		case errors.Is(err, inmemory.ErrOperationCanceled):
			return fmt.Errorf("%w: %v", ErrOperationCanceled, err)
		default:
			return fmt.Errorf("%v: %w", ErrExecuting, err)
		}
	}
	return nil
}

func (s *Storage) GetTaskByID(ctx context.Context, id int) (domain.Record, error) {
	record, err := s.store.LoadContext(ctx, id)
	if err != nil {
		if errors.Is(err, inmemory.ErrNotFound) {
			return domain.Record{}, fmt.Errorf("%w: %v", ErrNotFound, err)
		}
		return domain.Record{}, fmt.Errorf("%v: %w", ErrExecuting, err)
	}
	task, ok := record.(domain.Record)
	if !ok {
		return domain.Record{}, fmt.Errorf("%w", ErrIncompatible)
	}
	return task, nil
}

func (s *Storage) GetTasks(ctx context.Context) ([]domain.Record, error) {
	records, err := s.store.AllContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", ErrExecuting, err)
	}
	var counter uint64
	tasks := make([]domain.Record, 0, len(records))
	for _, record := range records {
		counter++
		if counter%50 == 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("%v: %w", ErrOperationCanceled, ctx.Err())
			default:
			}
		}
		task, ok := record.(domain.Record)
		if !ok {
			return nil, fmt.Errorf("%w", ErrIncompatible)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
