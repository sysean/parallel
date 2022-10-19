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
		ret := <-ch
		if ret.err != nil {
			close(notifyChan)
			cancel()
			return nil, ret.err
		}
		res = append(res, ret.s)
	}

	close(notifyChan)
	close(ch)

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
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, err := task(ctx, i)
			select {
			case <-notifyChan:
			case ch <- result{s: res, err: err}:
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(notifyChan)
		for i := 0; i < len(wo.tasks); i++ {
			ret := <-ch
			if ret.err != nil {
				err = ret.err
				cancel()
				return
			}
			res = append(res, ret.s)
		}
	}()

	wg.Wait()
	close(ch)

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

	wg.Go(func() error {
		defer close(notifyChan)
		for i := 0; i < len(wo.tasks); i++ {
			ret := <-ch
			if ret.err != nil {
				return ret.err
			}
			res = append(res, ret.s)
		}
		return nil
	})

	err = wg.Wait()
	close(ch)

	return
}
