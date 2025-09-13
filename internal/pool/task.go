package pool

import "sync"

// TaskState represents a type of tasks states.
type TaskState string

const (
	Queued  TaskState = "queued"
	Running TaskState = "running"
	Done    TaskState = "done"
	Failed  TaskState = "failed"
)

// Task represents a function that will be executed by a worker.
type Task func() (interface{}, error)

// taskItem represents a Task that will be added to the pool.
type taskItem struct {
	index      int
	fn         Task
	state      TaskState
	maxRetries int
	mu         sync.Mutex
}

// setState sets state for the taskItem
func (t *taskItem) setState(state TaskState) {
	t.mu.Lock()
	t.state = state
	t.mu.Unlock()
}

// getState gets state for the taskItem
func (t *taskItem) getState() TaskState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state
}
