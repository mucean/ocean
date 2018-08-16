package ocean

import (
	"context"
	"time"
)

type Task interface {
	Handle()
}

func NewTask(f func()) Task {
	return &task{f}
}

func NewDeadlineTask(f func(), deadline time.Time) Task {
	return WithDeadline(NewTask(f), deadline)
}

func NewTimeoutTask(f func(), timeout time.Duration) Task {
	return WithTimeout(NewTask(f), timeout)
}

type task struct {
	handler func()
}

func (t *task) Handle() {
	t.handler()
}

func WithDeadline(t Task, time time.Time) Task {
	return &deadlineTask{t, time}
}

func WithTimeout(t Task, dur time.Duration) Task {
	return WithDeadline(t, time.Now().Add(dur))
}

type deadlineTask struct {
	t Task
	deadline time.Time
}

func (t *deadlineTask) Handle() {
	ctx, cancel := context.WithDeadline(context.Background(), t.deadline)
	done := make(chan bool)

	go func() {
		t.t.Handle()
		done <-true
	}()

	select {
	case <-ctx.Done():
	case <-done:
		cancel()
	}
}