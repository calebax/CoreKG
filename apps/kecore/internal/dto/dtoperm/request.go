package dtoperm

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type SetResourcePermRequest struct {
	apiobj.BaseRequest
	Request SetResourcePermEmbedRequest `json:"request"`
}

type PermOption struct {
	// 管理员列表
	ManageList []uint `json:"manage_list"`
	// 查看列表
	UserList []uint `json:"user_list"`
	// 禁用列表
	BanList   []uint               `json:"ban_list"`
	ScopeType foresttype.ScopeType `json:"scope_type"`
}

type SetResourcePermEmbedRequest struct {
	ResourceType foresttype.ResourceType `json:"resource_type"`
	ResourceID   uint                    `json:"resource_id"`
	PermOption   PermOption              `json:"perm_option"`
}

func (req *SetResourcePermRequest) Validity(resp *SetResourcePermResponse) {
	if req.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "资源类型为空"
		return
	}
	if req.Request.ResourceID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "资源ID为空"
		return
	}
}

type GetResourcePermRequest struct {
	apiobj.BaseRequest
	Request GetResourcePermEmbedRequest `json:"request"`
}
type GetResourcePermEmbedRequest struct {
	ResourceType foresttype.ResourceType `json:"resource_type"`
	ResourceID   uint                    `json:"resource_id"`
}

func (opt *GetResourcePermRequest) Validity(resp *GetResourcePermResponse) {
	if opt.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "资源类型为空"
		return
	}
	if opt.Request.ResourceID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "资源ID为空"
		return
	}
}
