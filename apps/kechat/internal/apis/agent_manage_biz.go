package apis

import (
	"context"

	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateAgentRequest 创建指令型机器人命令请求
type CreateAgentRequest struct {
	apiobj.BaseRequest
	Request chatagent.CreateAgentItem
}

// Validity 验证有效性
func (req *CreateAgentRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ShowName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_name_required" // 名称不能为空
		return
	}
	if len(req.Request.ChatModelIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_id_required" // 模型ID不能为空
		return
	}
	// if req.Request.IsEnableAIGC == 0 {
	// 	resp.Code = errcode.ErrCode_BadRequest
	// 	resp.Message = "是否开启AI不能为空"
	// }
	if req.Request.PromptTemplate == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_prompt_required" // 提示词不能为空
		return
	}
	// if req.Request.PromptJoinType == "" {
	// 	resp.Code = errcode.ErrCode_BadRequest
	// 	resp.Message = "提示词类型不能为空"
	// }
}

// CreateAgentResponse 创建指令型机器人命令响应
type CreateAgentResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// UpdateAgentRequest 更新指令型机器人命令请求
type UpdateAgentRequest struct {
	apiobj.BaseRequest
	Request chatagent.UpdateAgentItem
}

// Validity 验证有效性
func (req *UpdateAgentRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_id_required" // id不能为空
		return
	}
	if req.Request.ShowName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_name_required" // 名称不能为空
		return
	}
	if len(req.Request.ChatModelIDs) == 0 {
		if req.Request.AgentType != chattype.AgentTypeWorkflow {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_model_required" // 模型不能为空
			return
		}
	}
	if req.Request.PromptTemplate == "" {
		if req.Request.AgentType != chattype.AgentTypeWorkflow {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_prompt_required" // 提示词不能为空
			return
		}
	}
	if req.Request.AgentType == "knowledge" {
		req.Request.AgentType = chattype.AgentTypeRolePlay
	}
}

// UpdateAgentResponse 更新指令型机器人命令响应
type UpdateAgentResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// GetAgentDetailRequest 获取机器人详情请求
type GetAgentDetailRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

func (req *GetAgentDetailRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_id_required" // id不能为空
		return
	}
}

// GetAgentDetailResponse 获取机器人详情响应
type GetAgentDetailResponse struct {
	apiobj.BaseResponse
	Response struct {
		chatagent.AgentWithPerm
		chatagent.AgentItemInfo
	}
}

// TestPromptAgentRequest 测试指令型机器人命令请求
type TestAgentRequest struct {
	apiobj.BaseRequest
	Request struct {
		Input           chattype.InputList `json:"input"`
		PromptTemplate  string             `json:"prompt_template"`
		Temperature     float32            `json:"temperature"`
		ChatModelIDs    types.UintArray    `json:"chat_model_ids"`
		GreetingMessage string             `json:"greeting_message"`
		Question        string             `json:"question"`
		ForestIDS       []uint             `json:"forest_ids"`
		AgentID         uint               `json:"id"`
		agentInfo       *chatagent.AgentWithVersion
	}
}

func (req *TestAgentRequest) Validity(ctx context.Context, resp *apiobj.BaseResponse) {
	if req.Request.AgentID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_id_required" // id不能为空
		return
	}
	if len(req.Request.ChatModelIDs) == 0 || req.Request.Temperature == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_or_temperature_required" // 模型为空或者温度为0
		return
	}
	// 获取机器人信息
	agentInfo, err := chatagent.GetAgentDetail(ctx, req.Request.AgentID)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to fetch agent info: %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_fetch_agent_info_failed" // 获取机器人信息失败
		return
	}
	agentInfo.PromptTemplate = req.Request.PromptTemplate
	agentInfo.GreetingMessage = req.Request.GreetingMessage
	agentInfo.Temperature = req.Request.Temperature
	agentInfo.ForestOption.ForestIDs = req.Request.ForestIDS
	req.Request.agentInfo = agentInfo

	if agentInfo.AgentType == chattype.AgentTypePrompt {
		if req.Request.PromptTemplate == "" || req.Request.Input == nil ||
			len(req.Request.Input) == 0 {
			logs.ErrorContextf(ctx, "[chat] [TestPromptAgent] Invalid parameters: %s", "agent type is prompt")
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_parameters" // 参数错误
			return
		}
	}
	if agentInfo.AgentType == chattype.AgentTypeRolePlay {
		if req.Request.PromptTemplate == "" || req.Request.GreetingMessage == "" ||
			req.Request.Question == "" {
			logs.ErrorContextf(ctx, "[chat] [TestPromptAgent] Invalid parameters: %s", "agent type is prompt")
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_parameters" // 参数错误
			return
		}
	}
}
