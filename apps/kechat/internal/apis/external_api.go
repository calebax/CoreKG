package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/kechat/mds"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/insmtx/corekg/apps/kechat/services/svccoze"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

// ExternalToken 外部token生成
// @Tags 外部调用
// @Summary 外部token生成
// @Description 外部token生成
// @Router /chat.ExternalToken [post]
// @Param user body ExternalTokenRequest true "入参"
// @Success 200 {object} ExternalTokenResponse "返回值"
func ExternalToken(ctx *gin.Context, req *ExternalTokenRequest, resp *ExternalTokenResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "kechat_validate_params_failed")
		return
	}

	token, err := svccoze.CreateExternalToken(ctx, req.Request.AccessToken)
	if err != nil {
		logs.ErrorContextf(ctx, "create external token failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_external_token_failed"
		return
	}
	resp.Response.Token = token
}

// GetPersonalAccessToken 获取个人token
// @Tags 外部调用
// @Summary 获取个人token
// @Description 获取个人token
// @Router /chat.GetPersonalAccessToken [post]
// @Success 200 {object} GetPersonalAccessTokenResponse "返回值"
func GetPersonalAccessToken(ctx *gin.Context, req *GetPersonalAccessTokenRequest, resp *GetPersonalAccessTokenResponse) {
	data, err := svccoze.GetPersonalAccessToken(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "get personal access token failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_personal_access_token_failed"
		return
	}
	resp.Response.Token = data.APIKey
	resp.Response.ExpireAt = data.ExpireAt
}

// Deprecated: 废弃
// SetExternalStatus 设置agent外部状态
// @Tags 外部调用
// @Summary 设置agent外部状态
// @Description 设置agent外部状态
// @Router /chat.SetExternalStatus [post]
// @Param user body SetExternalStatusRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func SetExternalStatus(ctx *gin.Context, req *SetExternalStatusRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}

	uin := runtime.Uin(ctx)

	var ag *chattype.ChatAgent
	if err := dbutil.Chat().First(&ag, req.Request.AgentID).Error; err != nil {
		logs.ErrorContextf(ctx, "SetExternalStatus GetChatAgentByID(%v) failed: %s", req.Request.AgentID, err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_failed" // 查询agent失败
		return
	}

	if !perm.HasManageAct(ctx, uin, ag.ID, foresttype.ResourceTypeAgent) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, ag.ID)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_no_permission")) // 无权限更新此资源
		return
	}

	ag.ExternalStatus = req.Request.Status
	if err := dbutil.Chat().
		Save(&ag).
		Error; err != nil {
		logs.ErrorContextf(ctx, "SetExternalStatus SetExternalStatus failed: %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_external_status_failed" // 更新agentExternal状态失败
		return
	}
}

// Deprecated: 废弃
// GetExternalStatus 获取agent外部状态
// @Tags 外部调用
// @Summary 获取agent外部状态
// @Description 获取agent外部状态
// @Router /chat.GetExternalStatus [post]
// @Param user body GetExternalStatusRequest true "入参"
// @Success 200 {object} GetExternalStatusResponse "返回值"
func GetExternalStatus(ctx *gin.Context, req *GetExternalStatusRequest, resp *GetExternalStatusResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}

	ag, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetExternalStatus GetChatAgentByID(%v) failed: %s", req.Request.AgentID, err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_failed" // 获取agent失败
		return
	}
	resp.Response.Status = ag.ExternalStatus
}

// NewExternalChatStream 新建聊天问题流
// @Tags 外部调用
// @Summary 新建聊天问题流
// @Description 新建聊天问题流
// @Router /chat.NewExternalChatStream [post]
// @Param request body NewExternalChatStreamRequest true "request"
// @Success 200 {object} NewExternalChatStreamResponse
func NewExternalChatStream(ctx *gin.Context, req *NewExternalChatStreamRequest, resp *NewExternalChatStreamResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	cozeApiKey := ctx.MustGet(mds.KeyCozeAPIKey).(string)
	conversationID := ctx.MustGet(mds.KeyCozeConversationID).(string)
	agentID := ctx.MustGet(mds.KeyCozeAgentID).(string)

	messageID, err := svccoze.CreateExternalMessage(ctx, cozeApiKey, conversationID, agentID, req.Request.ClientID, req.Request.Question)
	if err != nil {
		logs.ErrorContextf(ctx, "NewExternalChatStream create coze message failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_question_failed" // 创建问题失败
		return
	}

	resp.Response.QuestionID = messageID
	resp.Response.StreamKey = messageID
}

// SubmitExternalChatStream 外部调用Chat聊天问题流
// @Tags 外部调用
// @Summary 外部调用Chat聊天问题流
// @Description 外部调用Chat聊天问题流
// @Router /chat.SubmitExternalChatStream [post]
// @Param request body SubmitExternalChatStreamRequest true "request"
// @Success 200 {object} apiobj.BaseResponse
func SubmitExternalChatStream(ctx *gin.Context, req *SubmitExternalChatStreamRequest, resp *apiobj.BaseResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}

	cozeApiKey := ctx.MustGet(mds.KeyCozeAPIKey).(string)
	if err := svccoze.CreateExternalChatStream(ctx, cozeApiKey, req.Request.StreamKey); err != nil {
		logs.ErrorContextf(ctx, "SubmitExternalChatStream create coze chat failed: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed" // 聊天失败
		return
	}
}

// cozeMessageObject matches Coze v3 chat message object (content_type fixed to "text").
type cozeMessageObject struct {
	Role        string `json:"role"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type cozeAgentChatRequest struct {
	Stream             *bool               `json:"stream"`
	AdditionalMessages []cozeMessageObject `json:"messages"`
}

// CozeAgentChat coze agent api接口
func CozeAgentChat(ctx *gin.Context) {
	var req cozeAgentChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logs.ErrorContextf(ctx, "CozeAgentChat BindJSON failed: %v", err)
		ctx.JSON(http.StatusBadRequest, apiobj.BaseResponse{
			Code:    errcode.ErrCode_BadRequest,
			Message: "kechat_invalid_request",
		})
		return
	}
	if len(req.AdditionalMessages) == 0 {
		logs.WarnContextf(ctx, "CozeAgentChat missing additional_messages")
		ctx.JSON(http.StatusBadRequest, apiobj.BaseResponse{
			Code:    errcode.ErrCode_BadRequest,
			Message: "kechat_invalid_request",
		})
		return
	}

	stream := false
	if req.Stream != nil {
		stream = *req.Stream
	}

	messages := make([]coze.ChatV3Message, 0, len(req.AdditionalMessages))
	for _, msg := range req.AdditionalMessages {
		role := msg.Role
		content := msg.Content
		contentType := msg.ContentType
		if role == "" || content == "" {
			logs.WarnContextf(ctx, "CozeAgentChat invalid message object")
			ctx.JSON(http.StatusBadRequest, apiobj.BaseResponse{
				Code:    errcode.ErrCode_BadRequest,
				Message: "kechat_invalid_request",
			})
			return
		}
		if contentType == "" {
			contentType = "text"
		}
		if contentType != "text" {
			logs.WarnContextf(ctx, "CozeAgentChat invalid content_type")
			ctx.JSON(http.StatusBadRequest, apiobj.BaseResponse{
				Code:    errcode.ErrCode_BadRequest,
				Message: "kechat_invalid_request",
			})
			return
		}
		if role != "user" && role != "assistant" {
			logs.WarnContextf(ctx, "CozeAgentChat invalid role")
			ctx.JSON(http.StatusBadRequest, apiobj.BaseResponse{
				Code:    errcode.ErrCode_BadRequest,
				Message: "kechat_invalid_request",
			})
			return
		}
		messages = append(messages, coze.ChatV3Message{
			Role:        role,
			Content:     content,
			ContentType: contentType,
		})
	}

	cozeApiKey := ctx.MustGet(mds.KeyCozeAPIKey).(string)
	agentID := ctx.MustGet(mds.KeyCozeAgentID).(string)
	userID := uuid.NewString()

	if err := svccoze.CreateCozeChatAPI(ctx, cozeApiKey, agentID, userID, messages, stream); err != nil {
		logs.ErrorContextf(ctx, "CozeAgentChat create coze chat failed: %v", err)
		ctx.JSON(http.StatusBadRequest, apiobj.BaseResponse{
			Code:    errcode.ErrCode_InternalError,
			Message: "kechat_chat_failed",
		})
		return
	}
}
