package mutex

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ygpkg/yg-go/logs"
)

var _ sync.Locker = (*RedisLocker)(nil)

// RedisLocker .
type RedisLocker struct {
	rds         *redis.Client
	lockTimeout time.Duration
	lockerName  string
	// retryDelay 是重试间隔时间
	retryDelay time.Duration
}

// NewRedisLocker TODO
func NewRedisLocker(rds *redis.Client, name string, timeout time.Duration) *RedisLocker {
	retryDelay := timeout / 100
	if retryDelay < time.Millisecond*10 {
		retryDelay = time.Millisecond * 10
	} else if retryDelay > time.Second {
		retryDelay = time.Second
	}

	l := &RedisLocker{
		rds:         rds,
		lockTimeout: timeout,
		lockerName:  name,
		retryDelay:  retryDelay,
	}

	return l
}

// Lock .
func (l *RedisLocker) Lock() {
	key := l.lockerKey()
	for {
		ctx := context.TODO()
		val := l.rds.SetNX(ctx, key, 1, l.lockTimeout)
		if err := val.Err(); err != nil {
			logs.ErrorContextf(ctx, "lock %s failed: %v", key, err)
			return
		}
		if val.Val() {
			return
		}
		time.Sleep(l.retryDelay)
	}
}

// Unlock .
func (l *RedisLocker) Unlock() {
	key := l.lockerKey()
	ctx := context.TODO()
	err := l.rds.Del(ctx, key).Err()
	if err != nil {
		logs.ErrorContextf(ctx, "unlock %s failed: %v", key, err)
	}
}

// lockerKey .
func (l *RedisLocker) lockerKey() string {
	return "core:locker:" + l.lockerName
}
