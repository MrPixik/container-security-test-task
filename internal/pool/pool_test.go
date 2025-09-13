package pool

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolSubmitAndExecute(t *testing.T) {
	var counter int32
	p := New(2)

	for i := 0; i < 5; i++ {
		_, err := p.Submit(func() (interface{}, error) {
			atomic.AddInt32(&counter, 1)
			return nil, nil
		}, 0)
		if err != nil {
			t.Fatalf("unexpected error on submit: %v", err)
		}
	}

	p.Wait()
	p.Stop()

	if counter != 5 {
		t.Errorf("expected counter to be 5, got %d", counter)
	}
}

func TestPoolQueueFull(t *testing.T) {
	p := New(1, WithTaskBufferSize(1))

	if _, err := p.Submit(func() (interface{}, error) { return nil, nil }, 0); err != nil {
		t.Fatalf("unexpected error on submit: %v", err)
	}

	if _, err := p.Submit(func() (interface{}, error) { return nil, nil }, 0); !errors.Is(err, ErrTaskQueueIsFull) {
		t.Fatalf("expected ErrTaskQueueIsFull, got %v", err)
	}

	p.Stop()
}

func TestPoolResultCallback(t *testing.T) {
	var resultSum int32
	p := New(2, WithResultCallback(func(r interface{}) {
		if val, ok := r.(int); ok {
			atomic.AddInt32(&resultSum, int32(val))
		}
	}))

	for i := 1; i <= 3; i++ {
		p.Submit(func() (interface{}, error) {
			return i, nil
		}, 0)
	}

	p.Wait()
	p.Stop()

	if resultSum != 6 {
		t.Errorf("expected resultSum to be 6, got %d", resultSum)
	}
}

func TestPoolErrorCallback(t *testing.T) {
	var errCount int32
	p := New(1, WithErrorCallback(func(e error) {
		atomic.AddInt32(&errCount, 1)
	}))

	p.Submit(func() (interface{}, error) {
		return nil, errors.New("Task error")
	}, 0)

	p.Wait()
	p.Stop()

	if errCount != 1 {
		t.Errorf("expected errCount to be 1, got %d", errCount)
	}
}

func TestPoolStopMultipleTimes(t *testing.T) {
	p := New(1)

	for i := 0; i < 3; i++ {
		p.Submit(func() (interface{}, error) { return nil, nil }, 0)
	}

	p.Wait()
	time.Sleep(time.Second)
	if err := p.Stop(); err != nil {
		t.Errorf("unexpected error on first stop: %v", err)
	}

	if err := p.Stop(); err != nil {
		t.Errorf("unexpected error on second stop: %v", err)
	}
}

func TestPoolTaskExecutionOrder(t *testing.T) {
	var counter int64 = 0
	p := New(2)

	for i := 0; i < 5; i++ {
		p.Submit(func() (interface{}, error) {
			atomic.AddInt64(&counter, 1)
			return nil, nil
		}, 0)
	}

	p.Wait()
	p.Stop()

	if counter != 5 {
		t.Fatalf("expected 5 tasks executed, got %d", counter)
	}
}

func TestTaskStates(t *testing.T) {
	p := New(2)

	taskCount := 3
	taskIndexes := make([]int, 0, taskCount)

	for i := 0; i < taskCount; i++ {
		idx, err := p.Submit(func() (interface{}, error) {
			time.Sleep(500 * time.Millisecond)
			return i, nil
		}, 0)
		if err != nil {
			t.Fatalf("failed to submit Task: %v", err)
		}
		taskIndexes = append(taskIndexes, idx)
	}

	states := p.GetTasksState()
	for _, idx := range taskIndexes {
		if states[idx] != Queued {
			t.Errorf("expected Task %d to be Queued, got %s", idx, states[idx])
		}
	}

	time.Sleep(50 * time.Millisecond)
	states = p.GetTasksState()
	runningOrDone := false
	for _, idx := range taskIndexes {
		if states[idx] == Running || states[idx] == Done {
			runningOrDone = true
		}
	}
	if !runningOrDone {
		t.Errorf("expected at least one Task to be Running or Done, got %+v", states)
	}

	p.Wait()

	states = p.GetTasksState()
	for _, idx := range taskIndexes {
		if states[idx] != Done {
			t.Errorf("expected Task %d to be Done, got %s", idx, states[idx])
		}
	}
}
