package ocean

import "sync"

type Pool struct {
	sync.Mutex

	// the max goroutine number of the pool, if the current running goroutine up to the capacity of the pool,
	// next task will block until any worker done and recycle
	capacity chan int

	// number of the current running worker
	running int

	workers []*worker
}

func NewPool(capacity int) *Pool {
	return &Pool{capacity:make(chan int, capacity)}
}

func (p *Pool) recycle(w *worker) {
	p.Lock()
	defer p.Unlock()

	p.workers = append(p.workers, w)
	p.running -= 1
	p.capacity <- 1
}

func (p *Pool) enable() {
	p.Lock()
	defer p.Unlock()
	p.running -= 1
	p.capacity <- 1
}