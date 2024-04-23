package retry

import (
	"context"
	"math/rand"
	"time"
)

type options struct {
	maxRetryCount int
	backoff       func(int) time.Duration
}

type Optinon interface {
	apply(*options)
}

type Limit int

func (l Limit) apply(opts *options) {
	opts.maxRetryCount = int(l)
}

type BackoffStrategy func(int) time.Duration

func (b BackoffStrategy) apply(opts *options) {
	opts.backoff = b
}

func Strategy(b func(int) time.Duration) BackoffStrategy {
	return BackoffStrategy(b)
}

type Process func() error

func Retry(ctx context.Context, fn Process, opts ...Optinon) error {
	options := options{
		maxRetryCount: 3,
		backoff:       Strategy(retryBackOff),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	var lastErr error = nil
	for attempt := 0; attempt < options.maxRetryCount; attempt++ {
		err := fn()
		if err != nil {
			sErr := sleep(ctx, options.backoff(attempt))
			lastErr = err
			if sErr != nil {
				break
			}
			continue
		}
		lastErr = nil
		break
	}
	return lastErr
}

func retryBackOff(retry int) time.Duration {
	minBackoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second
	if retry < 0 {
		panic("not reached")
	}
	d := minBackoff << uint(retry)
	// 防止溢出
	if d < minBackoff {
		return maxBackoff
	}
	d = minBackoff + time.Duration(rand.Int63n(int64(d)))
	if d > maxBackoff || d < minBackoff {
		d = maxBackoff
	}
	return d
}

func sleep(ctx context.Context, dur time.Duration) error {
	tick := time.NewTicker(dur)
	defer tick.Stop()
	select {
	case <-tick.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
