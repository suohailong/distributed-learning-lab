package distributelock

import (
	"context"
	"net"
	"testing"
	"time"

	distributelock "distributed-learning-lab/distribute_lock"

	"github.com/stretchr/testify/assert"
)

func TestTryLock_NetworkError(t *testing.T) {
	dl := distributelock.NewRedisLock(
		distributelock.WithAddr("localhost:8888"),
		distributelock.WithLockTTl(5*time.Second),
	)
	ctx, cancnel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancnel()
	_, err := dl.TryLock(ctx, "mylock", "本机Ip")
	assert.ErrorIs(t, err, &net.OpError{})
}
