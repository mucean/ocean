package ocean

import (
	"sync"
	"errors"
	"time"
)

var statusClosed = -1
var statusStop = 0
var statusEnable = 1
var stopErr = errors.New("worker can't work at this time")

func newWorker() worker {
	return worker{startTime:time.Now()}
}

type worker struct {
	runOnce sync.Once
	closeOnce sync.Once
	p *Pool
	t chan Task
	status int
	err error
	startTime time.Time
	endTime time.Time
}

func (w *worker) run() {
	w.runOnce.Do(func() {
		if w.status == statusClosed {
			panic(w.err.Error())
		}

		w.endTime = time.Now()
		go func() {
			defer func() {
				if err := recover(); err != nil {
					w.close()
				}
			}()
			for t := <-w.t; w.status != statusClosed; t = <-w.t {
				if t != nil {
					t.Handle()
				}
				w.endTime = time.Now()
				w.recycle()
			}
		}()
	})
}

func (w *worker) Send(t Task) error {
	if w.status != statusEnable {
		return w.err
	}
	w.t <- t

	return nil
}

func (w *worker) Close() {
	w.close()
}

func (w *worker) enable() error {
	if w.status == statusClosed {
		return w.err
	}

	w.enableStatus()

	return nil
}

func (w *worker) stop() error {
	if w.status == statusClosed {
		return w.err
	}

	w.stopStatus()
	w.err = stopErr

	return nil
}

func (w *worker) recycle() {
	w.p.recycle(w)
	w.stop()
}

func (w *worker) close() {
	w.closeOnce.Do(func () {
		close(w.t)
		w.closeStatus()
		w.err = errors.New("worker is closed")
	})
}

func (w *worker) closeStatus() {
	w.changeStatus(statusClosed)
}

func (w *worker) stopStatus() {
	w.changeStatus(statusStop)
}

func (w *worker) enableStatus() {
	w.changeStatus(statusEnable)
}

func (w *worker) changeStatus(status int) {
	w.status = status
}