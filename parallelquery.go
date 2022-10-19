// Package parallel use ParallelingWithTimeout to query, and set the timeout
package parallel

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// open the comment, compatibles with versions earlier than 1.18
type any = interface{}

type (
	Result   any
	TaskFunc func(ctx context.Context, i int) (Result, error)
)

var (
	ErrTimeout = errors.New("service is timeout")
)

type Worker struct {
	tasks []TaskFunc
}

func New(tasks []TaskFunc) *Worker {
	return &Worker{tasks: tasks}
}

func (wo *Worker) Add(task TaskFunc) *Worker {
	wo.tasks = append(wo.tasks, task)
	return wo
}

// ParallelingWithTimeout do the query
func (wo *Worker) ParallelingWithTimeout(ctx context.Context, timeout time.Duration) (res []Result, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type result struct {
		s   Result
		err error
	}

	ch := make(chan result)
	notifyChan := make(chan struct{}) // make sure that all goroutines can exit finally
	defer close(notifyChan)

	for i, task := range wo.tasks {
		task := task
		go func(i int) {
			res, err := task(ctx, i)
			select {
			case <-notifyChan:
			case ch <- result{s: res, err: err}:
			}
		}(i)
	}

	for i := 0; i < len(wo.tasks); i++ {
		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case ret := <-ch:
			if ret.err != nil {
				cancel()
				return nil, ret.err
			}
			res = append(res, ret.s)
		}
	}

	return res, nil
}

// ParallelingWithTimeoutV2 do the query
func (wo *Worker) ParallelingWithTimeoutV2(ctx context.Context, timeout time.Duration) (res []Result, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wg := &sync.WaitGroup{}

	type result struct {
		s   Result
		err error
	}

	ch := make(chan result)
	notifyChan := make(chan struct{}) // make sure that all goroutines can exit finally

	for i, task := range wo.tasks {
		task := task
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := task(ctx, i)
			select {
			case <-notifyChan:
			case ch <- result{s: res, err: err}:
			}
		}()
	}

	overChan := make(chan error)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(notifyChan)
		for i := 0; i < len(wo.tasks); i++ {
			select {
			case <-ctx.Done():
				overChan <- ErrTimeout
				return
			case ret := <-ch:
				if ret.err != nil {
					err = ret.err
					return
				}
				res = append(res, ret.s)
			}
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
		close(overChan)
	}()

	over := <-overChan
	if over != nil {
		err = over
	}
	return
}

// ParallelingWithTimeoutV3 do the query
func (wo *Worker) ParallelingWithTimeoutV3(ctx context.Context, timeout time.Duration) (res []Result, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	type result struct {
		s   Result
		err error
	}

	ch := make(chan result)
	notifyChan := make(chan struct{}) // make sure that all goroutines can exit finally

	for i, task := range wo.tasks {
		i := i
		task := task
		wg.Go(func() error {
			res, err := task(ctx, i)
			select {
			case <-notifyChan:
			case ch <- result{s: res, err: err}:
			}
			return nil
		})
	}

	overChan := make(chan error)
	wg.Go(func() error {
		defer close(notifyChan)
		for i := 0; i < len(wo.tasks); i++ {
			select {
			case <-ctx.Done():
				overChan <- ErrTimeout
				return nil
			case ret := <-ch:
				if ret.err != nil {
					return ret.err
				}
				res = append(res, ret.s)
			}
		}
		return nil
	})

	go func() {
		err = wg.Wait()
		close(ch)
		close(overChan)
	}()

	over := <-overChan
	if over != nil {
		err = over
	}
	return
}
