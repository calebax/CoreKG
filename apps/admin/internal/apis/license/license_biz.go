package license

import (
	"strings"

	"github.com/insmtx/corekg/apps/admin/models/admintype"
	adminlicense "github.com/insmtx/corekg/apps/admin/models/license"
	"github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GenerateLicenseRequest struct {
	apiobj.BaseRequest
	Request struct {
		//签发主体
		Subject string `json:"subject"`
		//环境类型(option: kubernetes, physical)
		Env license.EnvType `json:"env"`
		//环境UID
		UID string `json:"uid"`
		//签发人
		Issuer string `json:"issuer"`
		//到期时间
		ExpiredAt int `json:"expired_at"`
		//备注
		Note string `json:"note"`
		// Version(option: all, agent)
		VersionKey string `json:"version_key"`
	}
}

func (req *GenerateLicenseRequest) Validate(resp *apiobj.BaseResponse) {
	if len(req.Request.Subject) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "签发主体非法"
		return
	}
	if len(req.Request.UID) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "非法环境唯一标识"
		return
	}
	if len(req.Request.Issuer) <= 0 || (strings.ToLower(req.Request.Issuer) != "yygu" && strings.ToLower(req.Request.Issuer) != "h3c") {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "非法签发人"
		return
	}
	if req.Request.ExpiredAt < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "到期时间"
		return
	}
	switch req.Request.Env {
	case license.EnvTypeKubernetes, license.EnvTypePhysical:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "未知环境类型"
		return
	}
	if _, ok := global.VersionKeyMap[req.Request.VersionKey]; !ok {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "版本类型不存在"
		return
	}
	if len(global.VersionKeyMap[req.Request.VersionKey]) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "版本类型为空"
		return
	}
}

type ListLicenseRequest apiobj.QueryRequest

type ListLicenseResponse struct {
	apiobj.BaseResponse
	Response adminlicense.ListLicenseResponse
}

func (req *ListLicenseRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "expired_at", "created_at desc", "updated_at desc", "expired_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "orderBy不能为空"
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "serial", "env", "uid", "subject", "issuer", "expired_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的字段只能有一个值"
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的值不能为空"
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "查询条件中的字段不存在, " + v.Field
			return
		}
	}
}

type DistributeLicenseRequest struct {
	apiobj.BaseRequest
	Request struct {
		//签发主体
		Subject string `json:"subject"`
		//环境类型(option: kubernetes, physical)
		Env license.EnvType `json:"env"`
		//环境UID
		UID string `json:"uid"`
		//签发人
		Issuer string `json:"issuer"`
		//到期时间
		ExpiredAt int `json:"expired_at"`
		//备注
		Note string `json:"note"`
		//版本
		VersionKey string `json:"version_key"`
	}
}

type DistributeLicenseResponse struct {
	apiobj.BaseResponse
	Response struct {
		License admintype.License
	}
}

func (req *DistributeLicenseRequest) Validate(resp *apiobj.BaseResponse) {
	if len(req.Request.Subject) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "签发主体非法"
		return
	}
	if len(req.Request.UID) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "非法环境唯一标识"
		return
	}
	if len(req.Request.Issuer) <= 0 || strings.ToLower(req.Request.Issuer) != "h3c" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "非法签发人"
		return
	}
	if req.Request.ExpiredAt < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "到期时间"
		return
	}
	switch req.Request.Env {
	case license.EnvTypeKubernetes, license.EnvTypePhysical:
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "未知环境类型"
		return
	}
}
