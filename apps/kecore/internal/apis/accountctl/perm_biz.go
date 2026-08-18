package accountctl

import (
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetForestPermSetRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin uint `json:"uin"`
	}
}

type GetForestPermSetResponse struct {
	apiobj.BaseResponse
	Response struct {
		PermSet []*perm.Set `json:"perm_set"`
	}
}

func (req *GetForestPermSetRequest) Valid(resp *GetForestPermSetResponse) bool {
	if req.Request.Uin < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_uin" // 非法Uin
		return false
	}
	return true
}

type ModifyForestPermSetRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin     uint        `json:"uin"`
		PermSet []*perm.Set `json:"perm_set"`
	}
}

type ModifyForestPermSetResponse struct {
	apiobj.BaseResponse
}

func (req *ModifyForestPermSetRequest) Valid(resp *ModifyForestPermSetResponse) bool {
	if req.Request.Uin <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_uin" // 非法Uin
		return false
	}
	if len(req.Request.PermSet) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_empty_perm_set" // 权限集为空
		return false
	}
	return true
}

type UpdateForestWithPermRequest struct {
	apiobj.BaseRequest
	Request struct {
		forest.WithPermItem
	}
}

func (opt *UpdateForestWithPermRequest) Validity(resp *apiobj.BaseResponse) {
	if len(opt.Request.Name) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_empty_forest_name" // 请输入新名称
		return
	}
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id" // 请选择要修改的知识森林
		return
	}
}

type GetForestWithPermResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data *forest.WithPerm `json:"data"`
	}
}
