package distributelock

import "context"

type Lock interface {
	// 非阻塞， 当获取不到锁时，立即返回
	TryLock(ctx context.Context, key, value string) (ok bool, err error)
	// 阻塞， 当获取不到锁时，阻塞等待
	Lock(ctx context.Context, key, value string) (ok bool, err error)
	UnLock(ctx context.Context, key, value string) (bool, error)
}
