package apis

import (
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/perm"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/types"
)

type GetBindCompanyKeyWithPermSetRequest struct {
	apiobj.BaseRequest
	Request struct {
		Count          uint                `json:"count"`
		InvitationRole accounttype.SysRole `json:"invitation_role"`
		Issuer         string              `json:"issuer"`
		Expired        time.Duration       `json:"expired"`
		PermSet        *perm.Set           `json:"perm_set"`
		DepartmentIDs  types.UintArray     `json:"department_ids"`
	}
}

type GetBindCompanyKeyWithPermSetResponse struct {
	apiobj.BaseResponse
	Response struct {
		Key string `json:"key"`
	}
}

func (req *GetBindCompanyKeyWithPermSetRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Count == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_expired_or_exhausted" // 已过期或被用尽
		return
	}
	if req.Request.InvitationRole != accounttype.SysRoleSysEmployee &&
		req.Request.InvitationRole != accounttype.SysRoleSysAdmin {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_role_parameter" // Role参数错误
		return
	}
	if req.Request.Issuer == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_issuer_parameter" // Issuer参数错误
		return
	}
}
