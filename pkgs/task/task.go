package task

import (
	"context"
	"fmt"
)

// TaskTypeInfo 单任务类型回调注册项。
type TaskTypeInfo struct {
	TaskType string
	CallBack func(ctx context.Context, tsk *Task) error
}

var callBackMap = map[string]*TaskTypeInfo{}

// GetCallBackMap 获取全部回调注册表。
func GetCallBackMap() map[string]*TaskTypeInfo {
	return callBackMap
}

// GetCallBack 获取任务类型对应的回调，未注册时返回错误。
func GetCallBack(taskType string) (*TaskTypeInfo, error) {
	if _, ok := callBackMap[taskType]; !ok {
		return nil, fmt.Errorf("taskType %s not found", taskType)
	}
	return callBackMap[taskType], nil
}

// RegisterCallBack 注册任务回调函数。
func RegisterCallBack(taskType string, callBack func(ctx context.Context, tsk *Task) error) {
	callBackMap[taskType] = &TaskTypeInfo{
		TaskType: taskType,
		CallBack: callBack,
	}
}
