package distributelock

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	"distributed-learning-lab/util/retry"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 10 * time.Second

type Options func(r *redisLock)

func WithAddr(addrs string) Options {
	return func(r *redisLock) {
		r.clientAddrs = addrs
	}
}

func WithLockTTl(ttl time.Duration) Options {
	return func(r *redisLock) {
		r.ttl = ttl
	}
}

type redisLock struct {
	re          *redis.Client
	ttl         time.Duration
	clientAddrs string
}

func NewRedisLock(opts ...Options) Lock {
	r := &redisLock{
		ttl:         defaultTTL,
		clientAddrs: "localhost:32479",
	}
	for _, o := range opts {
		o(r)
	}
	r.re = redis.NewClient(&redis.Options{
		Addr: r.clientAddrs,
	})
	return r
}

func (r *redisLock) TryLock(ctx context.Context, key, value string) (ok bool, err error) {
	retry.Retry(ctx, func() error {
		ok, err = r.re.SetNX(ctx, key, value, r.ttl).Result()
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) {
				return err
			}
		}
		return nil
	}, retry.Limit(3))

	return ok, err
}

func (r *redisLock) Lock(ctx context.Context, key, value string) (ok bool, err error) {
	retry.Retry(ctx, func() error {
		ok, err = r.re.SetNX(ctx, key, value, r.ttl).Result()
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) {
				return err
			}
		}
		if !ok && err == nil {
			return fmt.Errorf("failed to acquire lock")
		}
		return nil
	}, retry.Limit(math.MaxInt64))
	return ok, err
}

func (r *redisLock) UnLock(ctx context.Context, key, value string) (bool, error) {
	ov, err := r.re.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if ov == value {
		n, err := r.re.Del(ctx, key).Result()
		return n > 0, err
	}

	return false, nil
}
