package keqa

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

// StopChatStream 停止聊天流
func StopChatStream(ctx context.Context, question_id uint) error {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()), sseclient.WithExpiration(time.Minute*5))
	if err := sseClient.Stop(ctx, fmt.Sprintf("%v", question_id)); err != nil {
		logs.ErrorContextf(ctx, "[StopChat] sseClient.Stop failed, err: %v", err)
		return err
	}
	logs.InfoContext(ctx, "[StopChat] sseClient.Stop success streamID: %v", question_id)
	return nil
}

// GetStreamMessage 获取流是否结束，结束继续返回
func GetStreamMessage(ctx *gin.Context, question_id uint) error {
	streamId := fmt.Sprintf("%v", question_id)
	sseClient := sseclient.New(
		sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5),
		sseclient.WithBlockMaxRetry(1000),
	)
	latestID, historyMessages, err := sseClient.ReadMessages(ctx, streamId)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetMessage] sseClient.ReadMessages failed, err: %v", err)
		return err
	}
	if len(historyMessages) == 0 {
		// 没消息
		return nil
	}
	if err := sseClient.SendEvent(ctx.Writer, "hitstory\n"); err != nil {
		logs.ErrorContextf(ctx, "[GetMessage] sseClient.SendEvent failed, err: %v", err)
		return err
	}
	for _, message := range historyMessages {
		// 发送历史消息
		if err := sseClient.SendEvent(ctx.Writer, message); err != nil {
			logs.ErrorContextf(ctx, "[GetMessage] sseClient.SendEvent failed, err: %v", err)
			return err
		}
	}
	// 阻塞读取剩余消息
	done, affectedRaw, readErr := sseClient.BlockRead(ctx, ctx.Writer, streamId, latestID)
	if readErr != nil {
		logs.ErrorContextf(ctx, "[GetMessage] sseClient.BlockRead failed, err: %v", readErr)
		return err
	}
	logs.InfoContextf(ctx, "[GetMessage] affectedRaw: %d, done: %v", affectedRaw, done)
	return nil
}
