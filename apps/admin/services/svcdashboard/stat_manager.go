package svcdashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// MetricValue 统一的指标值接口
type MetricValue interface {
	GetIntValue() int64
	GetFloatValue() float64
}

// IntMetric 整数类型指标
type IntMetric struct {
	Value int64
}

func (m IntMetric) GetIntValue() int64 {
	return m.Value
}

func (m IntMetric) GetFloatValue() float64 {
	return float64(m.Value)
}

// FloatMetric 浮点数类型指标
type FloatMetric struct {
	Value float64
}

func (m FloatMetric) GetIntValue() int64 {
	return int64(m.Value)
}

func (m FloatMetric) GetFloatValue() float64 {
	return m.Value
}

// StatQuery 查询参数
type StatQuery struct {
	CompanyID uint      `json:"company_id"`
	BeginAt   time.Time `json:"begin_at"`
	EndAt     time.Time `json:"end_at"`
}

// StatFunc 统计函数签名
type StatFunc func(ctx context.Context, query *StatQuery) (MetricValue, error)

// StatResult 执行结果
type StatResult struct {
	Name  string
	Value MetricValue
	Error error
}

// StatManager 统计管理器
type StatManager struct {
	funcs map[string]StatFunc
	mu    sync.RWMutex
}

// NewStatManager 创建统计管理器
func NewStatManager() *StatManager {
	m := &StatManager{
		funcs: make(map[string]StatFunc),
	}

	return m
}

func (m *StatManager) BatchRegister(funcs map[string]StatFunc) error {
	for name, fn := range funcs {
		if err := m.register(name, fn); err != nil {
			return err
		}
	}
	return nil
}

// register 注册统计函数
func (m *StatManager) register(name string, fn StatFunc) error {
	if name == "" {
		return fmt.Errorf("stat name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("stat function cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.funcs[name]; exists {
		return fmt.Errorf("stat %s already registered", name)
	}

	m.funcs[name] = fn
	return nil
}

// Execute 并发执行所有统计函数，最大并发数为 15
func (m *StatManager) Execute(ctx context.Context, query *StatQuery) (map[string]MetricValue, error) {
	m.mu.RLock()
	funcs := make(map[string]StatFunc, len(m.funcs))
	for name, fn := range m.funcs {
		funcs[name] = fn
	}
	m.mu.RUnlock()

	if len(funcs) == 0 {
		return make(map[string]MetricValue), nil
	}

	result := make(map[string]MetricValue, len(funcs))
	resultMu := sync.Mutex{}

	g, gCtx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(15)

	for key, v := range funcs {
		name, fn := key, v

		g.Go(func() error {
			if err := sem.Acquire(gCtx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore for %s: %w", name, err)
			}
			defer sem.Release(1)

			value, err := fn(gCtx, query)

			resultMu.Lock()
			defer resultMu.Unlock()

			if err != nil {
				result[name] = nil
				logs.ErrorContextf(ctx, "[Execute] stat function %s failed: %v", name, err)
				return fmt.Errorf("stat function %s failed: %w", name, err)
			}

			result[name] = value
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return result, err
	}

	return result, nil
}
