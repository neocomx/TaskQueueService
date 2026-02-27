package task

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SaveAndGet(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{"simple task", "hello, world!"},
		{"task with symbols", "task!@#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			saved := store.Save(tt.payload)

			require.NotEmpty(t, saved.ID)
			assert.Equal(t, StatusPending, saved.Status)
			assert.Equal(t, tt.payload, saved.Payload)

			found, err := store.Get(saved.ID)
			require.NoError(t, err)
			assert.Equal(t, saved.ID, found.ID)
		})
	}
}

func TestStore_GetNotFound(t *testing.T) {
	store := NewStore()
	_, err := store.Get("Non-existent-id")
	assert.Error(t, err)
	assert.EqualError(t, err, "Task Non-existent-id not found")
}

func TestStore_List(t *testing.T) {
	store := NewStore()

	store.Save("task 1")
	store.Save("task 2")
	store.Save("task 3")

	tasks := store.List()
	assert.Len(t, tasks, 3)
}

func TestStore_ConcurrentAccess(t *testing.T) {
	store := NewStore()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Save("concurrent task")
		}()
	}

	wg.Wait()

	assert.Len(t, store.List(), 50)
}
