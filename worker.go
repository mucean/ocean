package ocean

import (
	"sync"
	"errors"
	"time"
	"fmt"
)

const closedStatus int = 1
const stopStatus int = 2
const enableStatus int = 3

func NewWorker() *worker {
	worker := &worker{startTime:time.Now(), status:enableStatus}
	worker.run()

	return worker
}

type worker struct {
	runOnce sync.Once
	closeOnce sync.Once
	sync.Mutex
	p *Pool
	t chan Task
	status int
	err error
	startTime time.Time
	endTime time.Time
}

func (w *worker) run() {
	w.runOnce.Do(func() {
		if w.status == enableStatus {
			panic(w.err.Error())
		}

		w.endTime = time.Now()
		go func() {
			defer func() {
				if err := recover(); err != nil {
					w.close(errors.New(fmt.Sprint(err)))
				}
			}()
			for t := <-w.t; w.status != closedStatus; t = <-w.t {
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
	if w.status != enableStatus {
		return w.err
	}
	w.t <- t

	return nil
}

func (w *worker) Close() {
	w.close(errors.New("worker is closed"))
}

func (w *worker) Enable() error {
	if w.status == closedStatus {
		return w.err
	}

	if w.status == enableStatus {
		return nil
	}

	return w.changeStatus(enableStatus)
}

func (w *worker) Stop(err error) error {
	if e := w.changeStatus(stopStatus); e != nil {
		return e
	}
	w.err = err

	return nil
}

func (w *worker) recycle() {
	w.p.recycle(w)
}

func (w *worker) close(err error) {
	w.closeOnce.Do(func () {
		close(w.t)
		w.changeStatus(closedStatus)
		w.err = err
	})
}

func (w *worker) enableStatus() {
	w.changeStatus(enableStatus)
}

func (w *worker) changeStatus(status int) error {
	w.Lock()
	defer w.Unlock()
	if w.status == closedStatus {
		return w.err
	}

	w.status = status

	return nil
}