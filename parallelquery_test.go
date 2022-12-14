package parallel

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelingWithTimeout(t *testing.T) {
	rand.Seed(time.Now().Unix())
	var goQuitNum int64 = 0
	const testSize = 100
	defer func() {
		// keep the program alive, observe the gorutine exit, if in a webserver application,
		// don't need this
		time.Sleep(5 * time.Second)
		qn := atomic.LoadInt64(&goQuitNum)
		t.Logf("goQuitNum = %d ~~~~~ \n", qn)
		if qn != testSize {
			t.Fatalf("goQuitNum should be %d", testSize)
		}
	}()

	tasks := make([]TaskFunc, 0, testSize)
	for i := 0; i < testSize; i++ {
		tasks = append(tasks, func(ctx context.Context, i int) (Result, error) {
			defer func() {
				fmt.Printf("gorutine[%d] quit!\n", i)
				atomic.AddInt64(&goQuitNum, 1)
			}()

			delay := randTimeBySecond(10)
			time.Sleep(delay)
			fmt.Printf("gorutine[%d] running, delay time: %s\n", i, delay)
			return nil, nil
		})
	}

	wo := New(tasks)
	start := time.Now()
	_, err := wo.ParallelingWithTimeoutV3(context.Background(), 5*time.Second)
	if err != nil {
		t.Logf("failed, got an error: %v, cost time: %s\n", err, time.Since(start))
		return
	}

	t.Logf("success over, cost time: %s\n", time.Since(start))
}

func randTimeBySecond(s int) time.Duration {
	return time.Duration(rand.Intn(s)) * time.Second
}
