package task

import (
	"fmt"
	"sync"
	"time"
)

type Store struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewStore() *Store {
	return &Store{
		tasks: make(map[string]*Task),
	}
}

func (s *Store) Save(payload string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("task-%d", time.Now().UnixNano())
	t := &Task{
		ID:        id,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.tasks[id] = t
	return t
}

func (s *Store) Get(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("Task %s not found", id)
	}

	return t, nil
}

func (s *Store) Update(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[t.ID]; !ok {
		return fmt.Errorf("task %s not found", t.ID)
	}

	t.UpdatedAt = time.Now()
	s.tasks[t.ID] = t
	return nil
}

func (s *Store) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}

	return tasks
}
