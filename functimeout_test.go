package parallel

import (
	"context"
	"errors"
	"testing"
	"time"
)

var testErr = errors.New("simple error")

func TestDoWithTimeout(t *testing.T) {
	type args struct {
		f       func() error
		timeout time.Duration
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantGotErr error
	}{
		{
			name: "return success",
			args: args{
				f: func() error {
					time.Sleep(time.Second)
					return nil
				},
				timeout: 3 * time.Second,
			},
		},
		{
			name: "should time out",
			args: args{
				f: func() error {
					time.Sleep(3 * time.Second)
					return nil
				},
				timeout: 2 * time.Second,
			},
			wantErr:    true,
			wantGotErr: ErrTimeout,
		},
		{
			name: "no timeout, but have other error",
			args: args{
				f: func() error {
					return testErr
				},
				timeout: 2 * time.Second,
			},
			wantErr:    true,
			wantGotErr: testErr,
		},
		{
			name: "timeout, business func also have err",
			args: args{
				f: func() error {
					time.Sleep(3 * time.Second)
					return testErr
				},
				timeout: 2 * time.Second,
			},
			wantErr:    true,
			wantGotErr: ErrTimeout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			if err := DoWithTimeout(context.Background(), tt.args.f, tt.args.timeout); (err != nil) != tt.wantErr {
				t.Errorf("DoWithTimeout() error = %v, wantErr %v", err, tt.wantErr)
			} else if !errors.Is(err, tt.wantGotErr) {
				t.Errorf("DoWithTimeout() error = %v, wantGotErr %v", err, tt.wantGotErr)
			} else if errors.Is(err, ErrTimeout) {
				t.Logf("case [%s] is timeout! time cost: %s", tt.name, time.Since(start))
				return
			}
			t.Logf("case [%s] time cost: %s", tt.name, time.Since(start))
		})
	}
}
