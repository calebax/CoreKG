package mutex

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNoResource = fmt.Errorf("no resource")
)

type Res interface {
	Key() string
}

type ResID uint

func (r ResID) Key() string {
	return fmt.Sprintf("%d", r)
}

type ResString string

func (r ResString) Key() string {
	return string(r)
}

// DistributedPool 是一个分布式池
type DistributedPool struct {
	name    string
	ctx     context.Context
	rds     *redis.Client
	timeout time.Duration

	lkr sync.Locker
}

// NewDistributedPool 创建一个分布式池
func NewDistributedPool(name string, rds *redis.Client, timeout time.Duration) *DistributedPool {
	np := &DistributedPool{
		name:    name,
		rds:     rds,
		ctx:     context.Background(),
		timeout: timeout,
		lkr:     NewRedisLocker(rds, name, time.Second),
	}
	return np
}

// Get 获取一个资源
func (p *DistributedPool) Get() (string, error) {
	p.lkr.Lock()
	defer p.lkr.Unlock()
	var (
		poolKey   = p.resourceZsetKey()
		memberKey string
	)
	{
		// 1. 从有序集合中获取一个资源
		res, err := p.rds.ZRangeByScore(p.ctx, poolKey, &redis.ZRangeBy{
			Max:   fmt.Sprintf("%d", nowUnix()),
			Count: 1,
		}).Result()
		if err != nil {
			return "", err
		}
		if len(res) != 1 {
			return "", ErrNoResource
		}
		memberKey = res[0]
	}
	{
		// 2. 将资源的Score更新到超时时间
		z := redis.Z{
			Score:  afterNowFloat(p.timeout),
			Member: memberKey,
		}
		err := p.rds.ZAdd(p.ctx, poolKey, z).Err()
		if err != nil {
			return "", err
		}
	}
	return memberKey, nil
}

// Put 归还一个资源
func (p *DistributedPool) Put(item Res) error {
	p.lkr.Lock()
	defer p.lkr.Unlock()
	z := redis.Z{
		Score:  nowUnixFloat() - 1,
		Member: item.Key(),
	}
	return p.rds.ZAdd(p.ctx, p.resourceZsetKey(), z).Err()

}

// Add 添加一个资源
func (p *DistributedPool) Add(item Res) error {
	p.lkr.Lock()
	defer p.lkr.Unlock()

	z := redis.Z{
		Score:  0,
		Member: item.Key(),
	}
	return p.rds.ZAddNX(p.ctx, p.resourceZsetKey(), z).Err()
}

// Remove 移除一个资源
func (p *DistributedPool) Remove(item Res) error {
	p.lkr.Lock()
	defer p.lkr.Unlock()
	return p.rds.ZRem(p.ctx, p.resourceZsetKey(), item.Key()).Err()
}

// resourceZsetKey
func (p *DistributedPool) resourceZsetKey() string {
	return fmt.Sprintf("core:dist_pool:%s:resources", p.name)
}

// nowUnix
func nowUnix() int64 {
	return time.Now().Unix()
}

func nowUnixFloat() float64 {
	return float64(time.Now().Unix())
}

func afterNowFloat(d time.Duration) float64 {
	return float64(time.Now().Unix()) + d.Seconds()
}
