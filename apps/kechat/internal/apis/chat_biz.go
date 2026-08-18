package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// SubmitChatQuestionStreamRequest 聊天问题流请求
type SubmitChatQuestionStreamRequest struct {
	apiobj.BaseRequest
	Request struct {
		QuestionID string `json:"question_id"`
	}
}

// SubmitChatQuestionStreamResponse 聊天问题流请求
type SubmitChatQuestionStreamResponse struct {
	apiobj.BaseResponse
}

// AgentChatResponse 机器人聊天响应
type AgentChatResponse struct {
	apiobj.BaseResponse
}

// ChatGetMessageRequest 获取消息请求
type ChatGetMessageRequest struct {
	apiobj.BaseRequest
	Request struct {
		QuestionID string `json:"question_id"`
	}
}

// Validity 校验请求有效性
func (opt *ChatGetMessageRequest) Validity(resp *ChatGetMessageResponse) {
	if opt.Request.QuestionID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_session" // 请选择会话
		return
	}
}

// ChatGetMessageResponse 获取消息响应
type ChatGetMessageResponse struct {
	apiobj.BaseResponse
}

// SubmitChatQuestionRequest 提交聊天问题请求
type SubmitChatQuestionRequest struct {
	apiobj.BaseRequest
	Request SubmitChatQuestionEmbedRequest
}

// SubmitChatQuestionEmbedRequest 嵌套的聊天问题请求
type SubmitChatQuestionEmbedRequest struct {
	// SessionID 会话id, 传值为 0 时，新建会话
	SessionID uint `json:"session_id"`
	// 模型id
	ModelID uint `json:"model_id" validate:"required"`
	// BaseType session 基础类型，standard：标准知识库（多模态）
	BaseType chattype.ResourceQASessionBaseType `json:"base_type"`
	// ResourceType 资源类型，forest：知识库列表，file_list: 文件列表
	ResourceType chattype.ResourceType `json:"resource_type"`
	// ResourceIDs 各种资源的id数组
	ResourceIDs types.UintArray `json:"resource_ids"`
	// ResourceNames 资源名称
	ResourceNames types.StringArray `json:"resource_names"`
	// Question 问题
	Question string `json:"question" validate:"required"`
}

// Validity 校验请求有效性
func (opt *SubmitChatQuestionRequest) Validity(resp *SubmitChatQuestionResponse) {
	validBaseTypeMap := map[chattype.ResourceQASessionBaseType]struct{}{
		chattype.ResourceQASessionBaseTypeStandard: {},
	}

	validResourceTypeMap := map[chattype.ResourceType]struct{}{
		chattype.ResourceTypeForest:   {},
		chattype.ResourceTypeFileList: {},
	}

	if opt.Request.SessionID == 0 {
		if opt.Request.BaseType == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_select_session_type" // 请选择会话类型
			return
		}
		if _, ok := validBaseTypeMap[opt.Request.BaseType]; !ok {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_session_type" // 无效的会话类型
			return
		}
		if _, ok := validResourceTypeMap[opt.Request.ResourceType]; !ok {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_resource_type" // 无效的资源类型
			return
		}
		if opt.Request.ResourceType == "" {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_select_resource_type" // 请选择资源类型
			return
		}
		if len(opt.Request.ResourceIDs) == 0 && len(opt.Request.ResourceNames) == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_select_resource" // 请选择资源
			return
		}
	}
	if opt.Request.ModelID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_model" // 请选择模型
		return
	}
	if opt.Request.Question == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_enter_question" // 请输入问题
		return
	}
}

// SubmitChatQuestionResponse 提交聊天问题响应
type SubmitChatQuestionResponse struct {
	apiobj.BaseResponse
	Response SubmitChatQuestionEmbedResponse
}

// SubmitChatQuestionEmbedResponse 嵌套的聊天问题响应
type SubmitChatQuestionEmbedResponse struct {
	// QuestionID 问题id
	QuestionID string `json:"question_id"`
	// Answer 回答
	Answer string `json:"answer"`
	// QueryReferenceList 查询引用
	QueryReferenceList *chattype.QueryReferenceList `json:"query_reference_list"`
}
