package distributelock

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	distributelock "distributed-learning-lab/distribute_lock"
	"distributed-learning-lab/util/log"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

// Use Docker to start a Redis instance for testing
func startRedisContainer() {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=test_redis", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to check Redis container:", err)
		return
	}

	if string(output) != "test_redis\n" {
		cmd = exec.Command("docker", "run", "--name", "test_redis", "-d", "-p", "6379:6379", "redis")
		err = cmd.Run()
		if err != nil {
			fmt.Println("Failed to start Redis container:", err)
		}
	} else {
		cmd = exec.Command("docker", "start", "test_redis")
		err = cmd.Run()
		if err != nil {
			fmt.Println("Failed to restart Redis container:", err)
		}
	}
}

// Stop the Redis container
func stopRedisContainer() {
	cmd := exec.Command("docker", "stop", "test_redis")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to stop Redis container:", err)
	}
}

type DistributeLockSuite struct {
	suite.Suite
	redisAddr string
	dl        distributelock.Lock
	redis     *redis.Client
}

func (suite *DistributeLockSuite) SetupSuite() {
	startRedisContainer()
	suite.redisAddr = "localhost:6379"
	suite.dl = distributelock.NewRedisLock(
		distributelock.WithAddr(suite.redisAddr),
		distributelock.WithLockTTl(10*time.Second),
	)
	suite.redis = redis.NewClient(&redis.Options{Addr: suite.redisAddr})
}

func (suite *DistributeLockSuite) TearDownSuite() {
	stopRedisContainer()
}

func (suite *DistributeLockSuite) TestTryLock() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	suite.Run("should acquire lock when it's not held", func() {
		lockey, value := "mylock", "ip01"
		defer func() {
			_, err := suite.redis.Del(ctx, lockey).Result()
			suite.NoError(err)
		}()
		ok, err := suite.dl.TryLock(ctx, lockey, value)
		suite.True(ok)
		suite.NoError(err)
		val, err := suite.redis.Get(ctx, lockey).Result()
		suite.NoError(err)
		suite.Equal(val, value)
	})

	suite.Run("should not acquire lock when it's already held", func() {
		lockey := "mylock"
		node1 := "ip01"
		ok, err := suite.dl.TryLock(ctx, lockey, node1)
		suite.True(ok)
		suite.NoError(err)
		defer func() {
			_, err := suite.redis.Del(ctx, lockey).Result()
			suite.NoError(err)
		}()

		node2 := "ip02"
		ok, err = suite.dl.TryLock(ctx, lockey, node2)
		suite.False(ok)
		suite.NoError(err)
	})

	suite.Run("should handle lock expiration", func() {
		// Set a short expiration time and test if the lock is released after that
		dl := distributelock.NewRedisLock(
			distributelock.WithAddr(suite.redisAddr),
			distributelock.WithLockTTl(2*time.Second),
		)
		locKey, lockVal := "mylock", "node1"
		ok, err := dl.TryLock(ctx, locKey, lockVal)
		suite.True(ok)
		suite.NoError(err)
		time.Sleep(3 * time.Second)
		_, err = suite.redis.Get(ctx, locKey).Result()
		suite.ErrorIs(err, redis.Nil)
		// suite.Equal(val, )
	})

	suite.Run("should handle network errors", func() {
		// Simulate a network error here
		dl := distributelock.NewRedisLock(
			distributelock.WithAddr("9.9.9.9:100"),
			distributelock.WithLockTTl(3*time.Second),
		)
		locKey, lockVal := "mylock", "node1"
		_, err := dl.TryLock(ctx, locKey, lockVal)
		var netErr net.Error
		suite.ErrorAs(err, &netErr)
	})

	suite.Run("should handle concurrent lock attempts", func() {
		// Start multiple goroutines and try to acquire the lock
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		const numGoroutines = 100
		results := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			node := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, err := suite.dl.TryLock(ctx, "mylock", fmt.Sprintf("ip%d", node))
				suite.NoError(err)
				results <- ok
			}()
		}

		wg.Wait()
		close(results)

		numTrue := 0
		for res := range results {
			if res {
				numTrue++
			}
		}

		// Only one goroutine should have acquired the lock
		suite.Equal(1, numTrue)
	})
}

func (suite *DistributeLockSuite) TestLock() {
	lockey := "lock"
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ok, err := suite.dl.Lock(ctx, lockey, "node1")
		suite.True(ok)
		suite.NoError(err)
		log.Debugf("node1  get lock: %v", ok)
	}()
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		now := time.Now()
		ok, _ := suite.dl.Lock(ctx, lockey, "node2")
		suite.True(time.Now().After(now.Add(5 * time.Second)))
		log.Debugf("node1  get lock: %v, duration: %v", ok, time.Since(now))
	}()
	wg.Wait()
}

func TestDistributeLockSuite(t *testing.T) {
	suite.Run(t, new(DistributeLockSuite))
}
