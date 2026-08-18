package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ygpkg/yg-go/logs"
)

var natsBridge *NATSBridge

func SetNATSBridge(b *NATSBridge) {
	natsBridge = b
}

func PushTaskQueue(ctx context.Context, taskType string) error {
	if natsBridge == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	tsk, err := GetOnePendingTask(taskType, "nats-dispatcher")
	if err != nil {
		return fmt.Errorf("get pending task: %w", err)
	}
	if tsk == nil {
		logs.InfoContextf(ctx, "no pending task for type %s", taskType)
		return nil
	}
	var payload map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(tsk.Payload), &payload); jsonErr == nil {
		payload["task_id"] = tsk.ID
		enriched, marshalErr := json.Marshal(payload)
		if marshalErr == nil {
			err = natsBridge.PublishTaskRPC(tsk.TaskType, enriched)
		} else {
			err = natsBridge.PublishTaskRPC(tsk.TaskType, []byte(tsk.Payload))
		}
	} else {
		err = natsBridge.PublishTaskRPC(tsk.TaskType, []byte(tsk.Payload))
	}

	if err != nil {
		logs.ErrorContextf(ctx, "nats dispatch publish failed: %v, task_id: %d", err, tsk.ID)
		return err
	}
	logs.InfoContextf(ctx, "nats dispatched task: id=%d type=%s", tsk.ID, tsk.TaskType)
	return nil
}
