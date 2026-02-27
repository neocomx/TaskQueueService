package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/neocomx/TaskQueueService/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockProcessor struct {
	mock.Mock
}

func (m *MockProcessor) Process(ctx context.Context, tsk *task.Task) error {
	args := m.Called(ctx, tsk)
	return args.Error(0)
}

func TestPool_ProcessesTaskSuccessfully(t *testing.T) {
	store := task.NewStore()
	processor := new(MockProcessor)

	processor.On("Process", mock.Anything, mock.Anything).Return(nil)

	pool := NewPool(1, store, processor, 5*time.Second)
	pool.Start()

	saved := store.Save("hello")
	pool.Submit(saved)
	pool.Shutdown()

	found, err := store.Get(saved.ID)
	require.NoError(t, err)
	assert.Equal(t, task.StatusDone, found.Status)
	assert.Empty(t, found.Error)

	processor.AssertNumberOfCalls(t, "Process", 1)
}

func TestPool_HandlesProcessorError(t *testing.T) {
	store := task.NewStore()
	processor := new(MockProcessor)

	processor.On("Process", mock.Anything, mock.Anything).Return(fmt.Errorf("something went wrong"))

	pool := NewPool(1, store, processor, 5*time.Second)
	pool.Start()

	saved := store.Save("hello")
	pool.Submit(saved)
	pool.Shutdown()

	found, err := store.Get(saved.ID)
	require.NoError(t, err)
	assert.Equal(t, task.StatusFailed, found.Status)
	assert.Equal(t, "something went wrong", found.Error)
}

func TestPool_HandlesTimeout(t *testing.T) {
	store := task.NewStore()
	processor := new(MockProcessor)

	processor.On("Process", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			<-ctx.Done()
		}).
		Return(context.DeadlineExceeded)

	pool := NewPool(1, store, processor, 100*time.Millisecond)
	pool.Start()

	saved := store.Save("slow task")
	pool.Submit(saved)
	pool.Shutdown()

	found, err := store.Get(saved.ID)
	require.NoError(t, err)
	assert.Equal(t, task.StatusFailed, found.Status)
	assert.NotEmpty(t, found.Error)
}

func TestPool_ProcessesMultipleTasksConcurrently(t *testing.T) {
	store := task.NewStore()
	processor := new(MockProcessor)

	processor.On("Process", mock.Anything, mock.Anything).Return(nil)

	pool := NewPool(3, store, processor, 5*time.Second)
	pool.Start()

	for i := 0; i < 9; i++ {
		saved := store.Save("task")
		pool.Submit(saved)
	}

	pool.Shutdown()

	tasks := store.List()
	assert.Len(t, tasks, 9)
	for _, tsk := range tasks {
		assert.Equal(t, task.StatusDone, tsk.Status)
	}

	processor.AssertNumberOfCalls(t, "Process", 9)
}
