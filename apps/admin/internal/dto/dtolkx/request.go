package dtolkx

import (
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// SendVerifyCodeRequest 发送验证码请求
type SendVerifyCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		Phone string `json:"phone"`
	}
}

func (req *SendVerifyCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入手机号码"
		return
	}
}

type SendVerifyCodeResponse struct {
	apiobj.BaseResponse
}

type DetailInfo struct {
	admintype.LkxCustomerInfo
}

// VerifyVerifyCodeRequest 验证验证码请求
type VerifyVerifyCodeRequest struct {
	apiobj.BaseRequest
	Request struct {
		VerifyCode string     `json:"verify_code"`
		Data       DetailInfo `json:"data"`
	}
}

func (req *VerifyVerifyCodeRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.VerifyCode == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入验证码"
		return
	}

	if req.Request.Data.Phone == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入手机号码"
		return
	}

	if req.Request.Data.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入姓名"
		return
	}

	if req.Request.Data.CompanyName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入公司名称"
		return
	}

	if req.Request.Data.Email == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入邮箱"
		return
	}

	if req.Request.Data.Description == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入描述"
		return
	}
}

type VerifyVerifyCodeResponse struct {
	apiobj.BaseResponse
}
