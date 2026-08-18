package ginctx

import (
	"context"
	"encoding/hex"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/logs"
)

// 定义自定义类型作为 context key，避免键冲突
type contextKey string

func InitJobCtx(jobName, jobUUID string) *gin.Context {
	// 创建带有作业信息的日志上下文
	logCtx := logs.WithContextFields(context.Background(), "job", jobName)

	// 生成请求ID用于日志追踪
	reqID := hex.EncodeToString(uuid.Must(uuid.NewV4(), nil).Bytes())
	logCtx = context.WithValue(logCtx, contextKey(constants.CtxKeyRequestID), reqID)

	// 创建基础的HTTP请求对象
	req := &http.Request{
		URL: &url.URL{},
	}
	req = req.WithContext(logCtx)

	// 创建并返回gin上下文
	ctx := &gin.Context{
		Request: req,
	}

	return ctx
}
