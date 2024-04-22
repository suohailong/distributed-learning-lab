package distributelock

import (
	"context"
	"errors"
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
		ttl: defaultTTL,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *redisLock) TryLock(ctx context.Context, key, value string) (ok bool, err error) {
	err = retry.Retry(ctx, func() error {
		isOk, cmdErr := r.re.SetNX(ctx, key, value, r.ttl).Result()
		if cmdErr != nil && errors.Is(cmdErr, &net.OpError{}) {
			return cmdErr
		}
		ok = isOk
		err = cmdErr
		return nil
	}, retry.Limit(3))

	return ok, err
}

func (r *redisLock) Lock(ctx context.Context, key, value string) (ok bool, err error) {
	cmd := r.re.SetNX(ctx, key, value, r.ttl)
	ok, err = cmd.Result()
	if err != nil {
		return ok, err
	}
	for !ok {
		cmd := r.re.SetNX(ctx, key, value, r.ttl)
		ok, err = cmd.Result()
		if err != nil {
			return ok, err
		}
	}
	return ok, nil
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
