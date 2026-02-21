package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neocomx/TaskQueueService/task"
)

type Pool struct {
	queue     chan *task.Task
	store     *task.Store
	processor task.Processor
	workers   int
	wg        sync.WaitGroup
	timeout   time.Duration
}

func NewPool(workers int, store *task.Store, processor task.Processor, timeout time.Duration) *Pool {
	return &Pool{
		queue:     make(chan *task.Task, 100),
		store:     store,
		processor: processor,
		workers:   workers,
		timeout:   timeout,
	}
}

func (p *Pool) Start() {
	for i := range p.workers {
		p.wg.Add(1)
		go p.runWorker(i)
	}
	fmt.Printf("Worker pool started with %d workers \n", p.workers)
}

func (p *Pool) Submit(t *task.Task) {
	p.queue <- t
}

func (p *Pool) runWorker(id int) {
	defer p.wg.Done()

	for t := range p.queue {
		fmt.Printf("[Worker-%d] processing task %s\n", id, t.ID)

		t.Status = task.StatusProcessing
		p.store.Update(t)

		ctx, cancel := context.WithTimeout(context.Background(), p.timeout)

		err := p.processor.Process(ctx, t)
		cancel()

		if err != nil {
			t.Status = task.StatusFailed
			t.Error = err.Error()
		} else {
			t.Status = task.StatusDone
		}

		p.store.Update(t)
		fmt.Printf("[worker-%d] finished task %s -> %s \n", id, t.ID, t.Status)
	}
}

type PrintProcessor struct{}

func (p *PrintProcessor) Process(ctx context.Context, t *task.Task) error {
	fmt.Printf("[processor] working on: %s\n", t.Payload)

	if t.Payload == "fail" {
		return fmt.Errorf("Payload 'fail' always fail")
	}

	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("task timed out: %w", ctx.Err())
	}
}

func (p *Pool) Shutdown() {
	fmt.Println("Shutting down worker pool...")
	close(p.queue)
	p.wg.Wait()
	fmt.Println("Worker pool stopped")
}
