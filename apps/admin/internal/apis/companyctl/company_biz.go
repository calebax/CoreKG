package companyctl

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/admin/models/company"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateCompanyRequest 创建团队请求
type CreateCompanyRequest struct {
	apiobj.BaseRequest
	Request company.CreateCompanyOption
}

// Validity 校验
func (req *CreateCompanyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "团队名称不能为空"
		return
	}
}

// CreateCompanyResponse 创建团队响应
type CreateCompanyResponse struct {
	apiobj.BaseResponse
	Response struct {
		accounttype.Company
	}
}

// ListCompanyRequest 团队列表请求
type ListCompanyRequest apiobj.QueryRequest

// Validity 校验
func (req *ListCompanyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "created_at", "updated_at", "id",
			"name desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "orderBy字段不支持"
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "id":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的字段只能有一个值"
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "查询条件中的字段不支持"
			return
		}
	}
	return
}

// ListCompanyResponse 团队列表响应
type ListCompanyResponse struct {
	apiobj.BaseResponse
	Response company.QueryCompanyListResponse
}

// ModifyCompanyRequest 修改团队请求
type ModifyCompanyRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
		company.ModifyCompanyOption
	}
}

// Validity 校验
func (req *ModifyCompanyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "团队ID不能为空"
		return
	}
	if req.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "团队名称不能为空"
		return
	}
	return
}

// ModifyCompanyResponse 修改团队响应
type ModifyCompanyResponse struct {
	apiobj.BaseResponse
	Response struct {
		accounttype.Company
	}
}
