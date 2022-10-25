package parallel

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

func DoWithTimeout(ctx context.Context, f func() error, timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	g := errgroup.Group{}
	notifyChan := make(chan error)
	g.Go(f)

	go func() {
		notifyChan <- g.Wait()
	}()

	select {
	case <-ctx.Done():
		return ErrTimeout
	case err = <-notifyChan:
		return
	}
}
