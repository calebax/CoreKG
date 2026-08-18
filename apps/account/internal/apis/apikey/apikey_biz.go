package apikey

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateAgentApiKeyRequest struct {
	apiobj.BaseRequest
	Request struct {
		Name    string `json:"name"`
		AgentID uint   `json:"agent_id"`
		Expire  int    `json:"expire"`
	}
}

func (r *CreateAgentApiKeyRequest) Valid(p *apiobj.BaseResponse) {
	if r.Request.AgentID <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_agent_id" // agentID非法
		return
	}
	if r.Request.Expire < 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_expire" // expire非法
		return
	}
}

type DeleteAgentApikeyRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID  uint `json:"agent_id"`
		ApikeyID uint `json:"apikey_id"`
	}
}

func (r *DeleteAgentApikeyRequest) Valid(p *apiobj.BaseResponse) {
	if r.Request.AgentID <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_agent_id" // agentID非法
		return
	}
	if r.Request.ApikeyID <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_apikey_id" // apikeyID非法
	}
}

type ListAgentAPIKeyRequest struct {
	apiobj.BaseRequest
	Request struct {
		apiobj.PageQuery
	}
}

type ListAgentAPIKeyResponse struct {
	apiobj.BaseResponse
	Response apikey.ListAgentApikeyResponse
}

func (req *ListAgentAPIKeyRequest) Valid(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "expired_at", "api_key", "created_at", "updated_at",
			"name desc", "expired_at desc", "api_key desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "agent_id":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_filter_field_data" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

type SetAgentApiKeyStatusRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID  uint                        `json:"agent_id"`
		ApikeyID uint                        `json:"apikey_id"`
		Status   accounttype.AccessKeyStatus `json:"status"`
	}
}

func (r *SetAgentApiKeyStatusRequest) Valid(p *apiobj.BaseResponse) {
	if r.Request.AgentID <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_agent_id" // agentID非法
		return
	}
	if r.Request.ApikeyID <= 0 {
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_apikey_id" // apikeyID非法
	}
	switch r.Request.Status {
	case accounttype.AccessKeyStatusDisabled, accounttype.AccessKeyStatusNormal:
	default:
		p.Code = errcode.ErrCode_BadRequest
		p.Message = "account_invalid_status" // 未知期望状态
	}
}
