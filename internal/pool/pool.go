package pool

import (
	"sync"
)

// Pool manages a goroutine pool
type Pool struct {
	workers chan struct{}
	wg      sync.WaitGroup
	done    chan struct{}
}

// New creates a new goroutine pool
func New(size int) *Pool {
	return &Pool{
		workers: make(chan struct{}, size),
		done:    make(chan struct{}),
	}
}

// Submit submits a task to the pool
func (p *Pool) Submit(task func()) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		select {
		case p.workers <- struct{}{}:
			defer func() { <-p.workers }()
			task()
		case <-p.done:
			return
		}
	}()
}

// Shutdown closes the pool
func (p *Pool) Shutdown() {
	close(p.done)
	p.wg.Wait()
}