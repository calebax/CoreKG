package apis

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// PersonAuthRequest 用户提交实名信息
type PersonAuthRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin      uint   `json:"uin"`
		IDCard   string `json:"id_card"`
		RealName string `json:"real_name"`
	}
}

func (req *PersonAuthRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Uin == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_uin_empty" // UIN不能为空
		return
	}
	if req.Request.IDCard == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_id_card_empty" // 身份证不能为空
		return
	}
	if req.Request.RealName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_real_name_empty" // 姓名不能为空
		return
	}
	if err := validate.IsCardNumber(req.Request.IDCard); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_id_card" // 错误的身份证号
		return
	}
}

// PersonAuthResponse 用户提交实名信息
type PersonAuthResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// ListPersonAuthRequest 获取等待认证的列表
type ListPersonAuthRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

func (req *ListPersonAuthRequest) Validity(resp *ListPersonAuthResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "alias":
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

// ListPersonAuthResponse 获取等待认证的列表
type ListPersonAuthResponse struct {
	apiobj.BaseResponse
	Response accounttype.IndividualItemList
}

// ReviewPersonAuthRequest 审阅用户信息
type ReviewPersonAuthRequest struct {
	apiobj.BaseRequest
	Request struct {
		Uin    uint `json:"uin"`
		Review bool `json:"review"`
	}
}

// ReviewPersonAuthResponse 审阅用户信息
type ReviewPersonAuthResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
