package mutex

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestDistPool(t *testing.T) {
	if true {
		return
	}
	rds := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	ctx := context.Background()
	if err := rds.Ping(ctx).Err(); err != nil {
		t.Skip(err)
		return
	}
	pooMap := new(sync.Map)
	pool := NewDistributedPool("test", rds, time.Second*10)
	{
		for _, k := range []string{"A", "B", "C", "D", "E", "F"} {
			item := ResString(k)
			if err := pool.Add(item); err != nil {
				t.Fatal(err)
			}
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				time.Sleep(time.Duration(rand.Int63n(500)) * time.Millisecond)
				k, err := pool.Get()
				if err != nil {
					fmt.Printf("get failed: %v\n", err)
					time.Sleep(time.Duration(rand.Int63n(150)+200) * time.Millisecond)
					continue
				}

				t.Logf("%d.%d: -%s", idx, j, k)
				pv, ok := pooMap.Load(k)
				if !ok {
					pv = 0
				}
				pooMap.Store(k, pv.(int)+1)

				time.Sleep(time.Duration(rand.Int63n(1000)) * time.Millisecond)
				if err := pool.Put(ResString(k)); err != nil {
					fmt.Printf("put failed: %v\n", err)
					return
				}
				t.Logf("%d.%d: +%s", idx, j, k)
			}
		}(i)

	}

	wg.Wait()
	pooMap.Range(func(key, value interface{}) bool {
		fmt.Printf("%s: %v\n", key, value)
		return true
	})

}
