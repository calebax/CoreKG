package llmchat

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

func WriteContent(ctx *gin.Context, reqID, content string) {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	sseClient.SetHeaders(ctx.Writer)
	if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, reqID, WriteResult{
		Content: content,
	}.String()); err != nil {
		defer sseClient.Close(ctx, reqID)
		logs.ErrorContextf(ctx, "[llmchat.WriteContent] Failed to write Answering response to KEQA: %v", err)
		return
	} else if stoped {
		defer sseClient.Close(ctx, reqID)
		logs.ErrorContextf(ctx, "[llmchat.WriteContent] stream Stoped by KEQA")
		return
	}
	defer sseClient.Close(ctx, reqID)
}

func WriteStreamsResult(ctx *gin.Context, reqID string, streamsResult WriteResult) {
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	sseClient.SetHeaders(ctx.Writer)
	if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, reqID, streamsResult.String()); err != nil {
		defer sseClient.Close(ctx, reqID)
		logs.ErrorContextf(ctx, "[llmchat.WriteContent] Failed to write Answering response to KEQA: %v", err)
		return
	} else if stoped {
		defer sseClient.Close(ctx, reqID)
		logs.ErrorContextf(ctx, "[llmchat.WriteContent] stream Stoped by KEQA")
		return
	}
	defer sseClient.Close(ctx, reqID)
}
