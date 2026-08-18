package models

import (
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type EventResult struct {
	mu           sync.Mutex
	orderMapping map[string]int
	taskID       string
	taskOrder    atomic.Int32
}

func NewEventResult() *EventResult {
	return &EventResult{
		orderMapping: make(map[string]int),
	}
}

func (e *EventResult) GetTaskID() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.taskID == "" {
		e.taskID = uuid.New().String()
	}
	return e.taskID
}

func (e *EventResult) RenewTaskID() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.taskOrder.Store(1)
	e.taskID = uuid.New().String()
	return e.taskID
}

func (e *EventResult) GetTaskOrder() int {
	return int(e.taskOrder.Add(1))
}

func (e *EventResult) GetAndIncrOrder(key string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	order, ok := e.orderMapping[key]
	if !ok {
		e.orderMapping[key] = 1
		return 1
	}
	e.orderMapping[key] = order + 1
	return order + 1
}
