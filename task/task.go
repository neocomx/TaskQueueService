package task

import (
	"context"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

type Task struct {
	ID        string
	Payload   string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
	Error     string
}

type Processor interface {
	Process(ctx context.Context, t *Task) error
}
