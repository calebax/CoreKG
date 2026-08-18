package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/agentperm"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetAgentPermSetRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin uint `json:"uin"`
	}
}

type GetAgentPermSetResponse struct {
	apiobj.BaseResponse
	Response struct {
		PermSet []*agentperm.PermSet `json:"perm_set"`
	}
}

func (req *GetAgentPermSetRequest) Validity(resp *GetAgentPermSetResponse) {
}

type ModifyChatPermSetRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin     uint                 `json:"uin"`
		PermSet []*agentperm.PermSet `json:"perm_set"`
	}
}

type ModifyChatPermSetResponse struct {
	apiobj.BaseResponse
}

func (req *ModifyChatPermSetRequest) Validity(resp *ModifyChatPermSetResponse) {
	if req.Request.Uin <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_uin" // 非法Uin
	}
	if len(req.Request.PermSet) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_empty_perm_set" // 权限集为空
	}
}

// GetAgentInfoWithPermResponse 获取机器人详情带权限列表响应
type GetAgentInfoWithPermResponse struct {
	apiobj.BaseResponse
	Response struct {
		chatagent.AgentWithPerm
		chatagent.AgentItemInfo
	}
}
