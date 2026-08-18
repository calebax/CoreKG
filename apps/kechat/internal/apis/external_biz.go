package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ExternalTokenRequest 外部Token请求
type ExternalTokenRequest struct {
	apiobj.BaseRequest
	Request struct {
		AccessToken string `json:"access_token"`
	}
}

// ExternalTokenResponse 外部Token响应
type ExternalTokenResponse struct {
	apiobj.BaseResponse
	Response struct {
		Token string `json:"token"`
	}
}

// GetPersonalAccessTokenRequest 获取个人token请求
type GetPersonalAccessTokenRequest struct {
	apiobj.BaseRequest
}

// GetPersonalAccessTokenResponse 获取个人token响应
type GetPersonalAccessTokenResponse struct {
	apiobj.BaseResponse
	Response struct {
		Token    string `json:"token"`
		ExpireAt int64  `json:"expire_at"`
	}
}

// Validity 校验外部Token请求
func (req *ExternalTokenRequest) Validity(resp *ExternalTokenResponse) {
	if len(req.Request.AccessToken) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_agent_token_required" // 请求agent_token为空
	}
}

// SetExternalStatusRequest 设置外部状态请求
type SetExternalStatusRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID uint                    `json:"agent_id"`
		Status  chattype.ExternalStatus `json:"status"`
	}
}

// Validity 校验设置外部状态请求
func (req *SetExternalStatusRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.AgentID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_agent_id" // agent_id非法
		return
	}
	switch req.Request.Status {
	case chattype.ExternalStatusNormal, chattype.ExternalStatusDisable:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_status_option" // 非法状态选项
	}
}

// GetExternalStatusRequest 获取外部状态请求
type GetExternalStatusRequest struct {
	apiobj.BaseResponse
	Request struct {
		AgentID uint `json:"agent_id"`
	}
}

// GetExternalStatusResponse 获取外部状态响应
type GetExternalStatusResponse struct {
	apiobj.BaseResponse
	Response struct {
		Status chattype.ExternalStatus `json:"status"`
	}
}

// Validity 校验获取外部状态请求
func (req *GetExternalStatusRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.AgentID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_agent_id" // agent_id非法
		return
	}
}

// NewExternalChatStreamRequest 新建外部聊天流请求
type NewExternalChatStreamRequest struct {
	apiobj.BaseRequest
	Request struct {
		ClientID string             `json:"client_id"`
		Question string             `json:"question"`
		Input    chattype.InputList `json:"input"`
	}
}

// Validity 校验新建外部聊天流请求
func (r *NewExternalChatStreamRequest) Validity(p *apiobj.BaseResponse) {
	if len(r.Request.ClientID) == 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "kechat_client_id_required" // clientID不可为空
		return
	}
}

// NewExternalChatStreamResponse 新建外部聊天流响应
type NewExternalChatStreamResponse struct {
	apiobj.BaseResponse
	Response struct {
		QuestionID string `json:"question_id"`          // 问题ID
		StreamKey  string `json:"stream_key,omitempty"` // 流密钥
	}
}

// SubmitExternalChatStreamRequest 提交外部聊天流请求
type SubmitExternalChatStreamRequest struct {
	apiobj.BaseRequest
	Request struct {
		StreamKey string `json:"stream_key"`
		ClientID  string `json:"client_id"`
	}
}

// Validity 校验提交外部聊天流请求
func (r *SubmitExternalChatStreamRequest) Validity(p *apiobj.BaseResponse) {
	if len(r.Request.ClientID) == 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "kechat_client_id_required" // clientID不可为空
		return
	}
	if len(r.Request.StreamKey) == 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "kechat_stream_key_required" // streamKey不可为空
		return
	}
}
