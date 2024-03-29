package task

import (
	"sync"
)

// Worker must be implemented by types that want to use
// the run pool.
type Worker interface {
	Work()
}

// Task provides a pool of goroutines that can execute any Worker
// tasks that are submitted.
type Task struct {
	work chan Worker
	wg   sync.WaitGroup
}

// New creates a new work pool.
// instanced with singleton pattern beacuse used centrally as worker pool based on
// CPU resources
func New(maxGoroutines int) *Task {
	t := Task{
		// Using an unbuffered channel because we want the
		// guarantee of knowing the work being submitted is
		// actually being worked on after the call to Run returns.
		work: make(chan Worker),
	}

	// The goroutines are the pool. So we could add code
	// to change the size of the pool later on.

	t.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range t.work {
				w.Work()
			}
			t.wg.Done()
		}()
	}

	return &t
}

// Shutdown waits for all the goroutines to shutdown.
func (t *Task) Shutdown() {
	close(t.work)
	t.wg.Wait()
}

// Wait waits for all the goroutines to complete.
func (t *Task) Wait() {
	t.wg.Wait()
}

// Do submits work to the pool.
func (t *Task) Do(w Worker) {
	t.work <- w
}
