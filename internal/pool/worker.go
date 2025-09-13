package pool

import (
	"math/rand"
	"time"
)

// worker represents a worker in the pool.
type worker struct {
	taskQueue chan *taskItem
}

// newWorker creates a new worker.
func newWorker() *worker {
	return &worker{
		taskQueue: make(chan *taskItem),
	}
}

// start starts the worker in a separate goroutine.
// The worker will run tasks from its taskQueue until the taskQueue is closed.
func (w *worker) start(pool *pool, workerIndex int) {
	go func() {
		for t := range w.taskQueue {
			if t != nil {
				result, err := w.runTask(t, pool)
				w.handleResult(result, err, pool)
			}
			pool.workerStack <- workerIndex
			pool.wg.Done()
		}
	}()
}

// runTask executes a Task and returns the result and error.
// If the Task fails, it will be retried according to the maxRetries of the pool.
func (w *worker) runTask(ti *taskItem, pool *pool) (interface{}, error) {
	for i := 0; i <= ti.maxRetries; i++ {
		result, err := ti.fn()

		if err == nil {
			ti.setState(Done)
			return result, err
		}
		if i == ti.maxRetries {
			ti.setState(Failed)
			return result, err
		}

		//exponential backoff + jitter
		// jitter = rand[0,backoff]
		backoff := time.Duration(1<<i) * pool.baseDelay
		jitter := time.Duration(rand.Int63n(int64(backoff)))
		sleepTime := backoff + jitter

		time.Sleep(sleepTime)
	}
	return nil, nil
}

// handleResult handles the result of a Task.
func (w *worker) handleResult(result interface{}, err error, pool *pool) {
	if err != nil && pool.errorCallback != nil {
		pool.errorCallback(err)
	} else if pool.resultCallback != nil {
		pool.resultCallback(result)
	}
}
