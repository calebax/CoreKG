package mds

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

const (
	KeyCozeAPIKey         = "coze_api_key"
	KeyCozeAgentID        = "coze_agent_id"
	KeyCozeConversationID = "coze_conversation_id"
)

// CozeAgentAuthMD parses Authorization header as "Bearer agentCode" or "Bearer agentCode-conversation_id".
func CozeAgentAuthMD(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if token == "" {
		logs.WarnContextf(ctx, "coze auth missing authorization header")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	authStr := strings.TrimPrefix(token, auth.AuthBearer)
	authStr = strings.TrimSpace(strings.TrimPrefix(authStr, "Bearer"))
	if authStr == "" {
		logs.WarnContextf(ctx, "coze auth empty token")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	agentCode := authStr
	if idx := strings.Index(authStr, "-"); idx >= 0 {
		agentCode = strings.TrimSpace(authStr[:idx])
	}
	if agentCode == "" {
		logs.WarnContextf(ctx, "coze auth invalid token format")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cozeApiKey, agentID, err := coze.GetCozeAPIKeyByAgentCode(ctx, agentCode)
	if err != nil {
		logs.ErrorContextf(ctx, "get coze api key by agent code failed: %v", err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if cozeApiKey == "" || agentID == "" {
		logs.WarnContextf(ctx, "coze auth invalid agent code")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set(KeyCozeAPIKey, cozeApiKey)
	ctx.Set(KeyCozeAgentID, agentID)
	ctx.Next()
}

// CozeConversationAuthMD parses Authorization header as "Bearer agentCode-conversation_id".
func CozeConversationAuthMD(ctx *gin.Context) {
	token := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if token == "" {
		logs.WarnContextf(ctx, "coze auth missing authorization header")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	authStr := strings.TrimPrefix(token, auth.AuthBearer)
	authStr = strings.TrimSpace(strings.TrimPrefix(authStr, "Bearer"))
	if authStr == "" {
		logs.WarnContextf(ctx, "coze auth empty token")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	parts := strings.SplitN(authStr, "-", 2)
	if len(parts) != 2 || parts[1] == "" {
		logs.WarnContextf(ctx, "coze auth missing conversation id")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conversationID := strings.TrimSpace(parts[1])
	if conversationID == "" {
		logs.WarnContextf(ctx, "coze auth missing conversation id")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set(KeyCozeConversationID, conversationID)
	ctx.Next()
}
