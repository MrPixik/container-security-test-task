package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTaskQueueIsFull = fmt.Errorf("Task quenue is full")
	ErrNoTaskByIndex   = fmt.Errorf("no Task with this index")
)

// Pool represents a pool of workers.
type Pool interface {
	// Submit submits a Task to the pool.
	Submit(t Task, maxRetries int) (int, error)
	// Wait waits for all tasks in buffer to be completed.
	Wait()
	// Stop stops the pool and all its workers.
	Stop() error
	// GetTaskBufferSize returns the size of the Task buffer.
	GetTaskBufferSize() int
	// GetTasksState returns all states of the tasks.
	GetTasksState() []TaskState
	// GetTaskStateById returns states of the Task by it's id.
	GetTaskStateById(id int) (TaskState, error)
}

// pool represents a pool of workers.
type pool struct {
	workers     []*worker
	workerStack chan int
	workersNum  int

	// taskItem are added to this channel first, then sends to workers.
	taskBuffer chan *taskItem
	tasksMu    sync.RWMutex
	// slice of pointers to all taskItems.
	tasks []*taskItem
	// Set by WithTaskBufferSize(), used to set the size of the taskBuffer. Default is 1e6.
	taskBufferSize int
	// Set by WithRetryCount(), used to retry a Task when it fails. Default is 0.
	maxRetries int
	// baseDelay sets base delay time for backoff. Default is 0
	baseDelay time.Duration
	// Set by WithResultCallback(), used to handle the result of a Task. Default is nil.
	resultCallback func(interface{})
	// Set by WithErrorCallback(), used to handle the error of a Task. Default is nil.
	errorCallback func(error)
	ctx           context.Context
	// cancel is used to cancel the context. It is called when Stop() is called.
	cancel   context.CancelFunc
	stopOnce sync.Once
	wg       *sync.WaitGroup
}

// New creates a new pool of workers.
func New(workersNum int, opts ...Option) Pool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &pool{
		workersNum:  workersNum,
		workers:     make([]*worker, workersNum),
		workerStack: make(chan int, workersNum),

		taskBuffer:     nil,
		tasks:          make([]*taskItem, 0),
		taskBufferSize: 1e6,

		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}
	// Apply options
	for _, opt := range opts {
		opt(pool)
	}

	pool.taskBuffer = make(chan *taskItem, pool.taskBufferSize)

	// Create workers
	for i := 0; i < pool.workersNum; i++ {
		worker := newWorker()
		pool.workers[i] = worker
		pool.workerStack <- i
		worker.start(pool, i)
	}
	go pool.assignTask()
	return pool
}

// Submit adds a Task to the pool.
func (p *pool) Submit(t Task, maxRetries int) (int, error) {
	taskItm := &taskItem{
		fn:         t,
		state:      Queued,
		maxRetries: maxRetries,
	}

	select {
	case p.taskBuffer <- taskItm:

		p.tasksMu.Lock()
		taskItm.index = len(p.tasks)
		p.tasks = append(p.tasks, taskItm)
		p.tasksMu.Unlock()

		return taskItm.index, nil
	default:
		return -1, ErrTaskQueueIsFull
	}
}

// Wait waits for all tasks to be dispatched and completed.
func (p *pool) Wait() {
	for {
		if len(p.taskBuffer) == 0 {
			p.wg.Wait()
			time.Sleep(100 * time.Millisecond)
			return
		}
	}
}

// Stop stops all workers and releases resources.
func (p *pool) Stop() error {
	p.stopOnce.Do(func() {

		p.cancel()

		close(p.taskBuffer)

		for _, worker := range p.workers {
			close(worker.taskQueue)
		}

		p.wg.Wait()

		p.workers = nil
		p.workerStack = nil
	})
	return nil
}

// assignTask assigns Task tasks to workers.
func (p *pool) assignTask() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case t := <-p.taskBuffer:
			workerIndex := <-p.workerStack
			p.wg.Add(1)
			t.setState(Running)
			p.workers[workerIndex].taskQueue <- t
		}
	}
}

// GetTaskBufferSize returns the size of the Task queue.
func (p *pool) GetTaskBufferSize() int {
	return p.taskBufferSize
}

// GetTasksState returns all states of the tasks.
func (p *pool) GetTasksState() []TaskState {
	p.tasksMu.RLock()
	defer p.tasksMu.RUnlock()
	states := make([]TaskState, len(p.tasks))
	for i, t := range p.tasks {
		states[i] = t.getState()
	}
	return states
}

// GetTaskStateById returns states of the Task by it's id.
func (p *pool) GetTaskStateById(id int) (TaskState, error) {
	p.tasksMu.RLock()
	defer p.tasksMu.RUnlock()
	if id >= 0 && id < len(p.tasks) {
		state := p.tasks[id].state
		return state, nil
	} else {
		return Failed, ErrNoTaskByIndex
	}
}
