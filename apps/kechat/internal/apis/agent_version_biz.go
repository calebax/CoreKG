package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ListAgentVersionRequest 获取机器人版本请求
type ListAgentVersionRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID uint `json:"agent_id"`
		apiobj.PageQuery
	}
}

func (req *ListAgentVersionRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "agent_id", "created_at", "updated_at",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_order_by_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "agent_id", "created_at", "updated_at":
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

// ListAgentVersionResponse 获取机器人列表响应
type ListAgentVersionResponse struct {
	apiobj.BaseResponse
	Response chatagent.QueryAgentVersionListResponse
}

// ChooseAgentVersionRequest 选择机器人版本请求
type ChooseAgentVersionRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID   uint `json:"agent_id"`
		VersionID uint `json:"version_id"`
	}
}

func (req *ChooseAgentVersionRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.AgentID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_agent_id_required" // agent_id不能为空
		return
	}
	if req.Request.VersionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_version_id_required" // version_id不能为空
		return
	}
}

// ChooseAgentVersionResponse 选择机器人版本响应
type ChooseAgentVersionResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
