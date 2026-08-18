package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ListChatAgentRequest 获取机器人列表请求
type ListChatAgentRequest apiobj.QueryRequest

func (req *ListChatAgentRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_order_by_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "created_at", "updated_at", "show_name", "agent_type", "uin", "public_scope", "created_type":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_field_value_required" // 查询条件中的值不能为空
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_filter_field_not_exist" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ListChatAgentResponse 获取机器人列表响应
type ListChatAgentResponse struct {
	apiobj.BaseResponse
	Response chatagent.QueryChatAgentListResponse
}

// UpdateChatAgentRequest 机器人更新请求
type UpdateChatAgentRequest struct {
	apiobj.BaseRequest
	Request chatagent.UpdateChatAgentItem
}

// Validity 验证有效性
func (req *UpdateChatAgentRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_id_required" // id不能为空
		return
	}

	if req.Request.AvatarURL == "" && req.Request.ShowName == "" && req.Request.Description == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_avatar_name_description_required" // 头像、名称和描述至少需要填写一项
		return
	}
}

// UpdateChatAgentResponse 机器人更新响应
type UpdateChatAgentResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// DeleteChatAgentRequest 删除机器人请求
type DeleteChatAgentRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

// Validity 验证有效性
func (req *DeleteChatAgentRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_id_required" // id不能为空
		return
	}
}

// DeleteChatAgentResponse 删除机器人响应
type DeleteChatAgentResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// GetAgentInfoRequest 获取机器人详情请求
type GetAgentInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

func (req *GetAgentInfoRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_id_required" // id不能为空
		return
	}
}

// GetAgentInfoResponse 获取机器人详情响应
type GetAgentInfoResponse struct {
	apiobj.BaseResponse
	Response struct {
		AgentInfo *chatagent.AgentWithVersion `json:"agent_info"`
	}
}

// CreateAgentAppRequest 创建应用请求
type CreateAgentAppRequest struct {
	apiobj.BaseRequest
	Request chatagent.AgentApp
}

func (req *CreateAgentAppRequest) Validity(resp *CreateAgentAppResponse) {
	if req.Request.AvatarURL == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_avatar_required" // 请上传应用头像
		return
	}
	if req.Request.ShowName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_name_required" // 请填写应用名称
		return
	}
	if req.Request.Description == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_description_required" // 请填写应用描述
		return
	}
	if req.Request.AgentType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_agent_type_required" // 请选择应用类型
		return
	}
}

// CreateAgentAppResponse 创建应用响应
type CreateAgentAppResponse struct {
	apiobj.BaseResponse
	Response struct {
		ID         uint   `json:"id"`
		SpaceID    string `json:"coze_space_id"`
		WorkflowID string `json:"coze_workflow_id"`
	}
}

// ListCollectAppRequest 获取我的收藏列表请求
type ListCollectAppRequest apiobj.QueryRequest

type ListCollectAppResponse struct {
	apiobj.BaseResponse
	Response chatagent.QueryAgentCollectListResponse
}

func (req *ListCollectAppRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_order_by_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "created_at", "updated_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_field_value_required" // 查询条件中的值不能为空
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_filter_field_not_exist" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// CollectAppRequest 获取小程序列表
type CollectAppRequest apiobj.DetailIdRequest

func (req *CollectAppRequest) Validity(resp *CollectAppResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_app_required" // 请选择要收藏的小程序
		return
	}
}

// CollectAppResponse 获取小程序列表
type CollectAppResponse struct {
	apiobj.BaseResponse
}
