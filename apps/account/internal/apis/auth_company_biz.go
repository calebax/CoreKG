package apis

import (
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
)

// CompanyAuthRequest 认证公司
type CompanyAuthRequest struct {
	apiobj.BaseRequest
	Request struct {
		CompanyInfo *company.CompanyInfo `json:"company_info"`
	}
}

func (req *CompanyAuthRequest) Validity(resp *CompanyAuthResponse) {
	if req.Request.CompanyInfo == nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_company_info_empty" // 公司信息为空
		return
	}
	if req.Request.CompanyInfo.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_company_name_empty" // 公司名称不能为空
		return
	}
}

// CompanyAuthResponse 认证公司
type CompanyAuthResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

// ListPersonAuthRequest 获取等待认证的列表
type ListCompanyRequest struct {
	apiobj.BaseRequest
	Request struct {
		apiobj.PageQuery
	}
}

func (req *ListCompanyRequest) Validity(resp *ListCompanyResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "id DESC", "id ASC", "created_at", "updated_at",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "alias", "company_status":
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

// ListCompanyResponse 获取等待认证的列表
type ListCompanyResponse company.QueryCompanyListResponse

// GetCompanyRequest 获取公司详情
type GetCompanyRequest apiobj.DetailIdRequest

func (req *GetCompanyRequest) Validity(resp *GetCompanyResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_company_info_zero" // 公司信息为0
		return
	}
}

// GetCompanyResponse 获取公司详情
type GetCompanyResponse struct {
	apiobj.BaseResponse
	Response struct {
		Company *accounttype.Company `json:"company"`
	}
}

// ReviewCompanyAuthRequest 审阅用户信息
type ReviewCompanyAuthRequest struct {
	apiobj.BaseRequest
	Request struct {
		CompanyID uint `json:"company_id"`
		Review    bool `json:"review"`
	}
}

// ReviewCompanyAuthResponse 审阅用户信息
type ReviewCompanyAuthResponse struct {
	apiobj.BaseResponse
}

// GetBindCompanyKeyRequest 生成绑定公司密钥
type GetBindCompanyKeyRequest struct {
	apiobj.BaseRequest
	Request struct {
		Count          uint                `json:"count"`
		InvitationRole accounttype.SysRole `json:"invitation_role"`
		Issuer         string              `json:"issuer"`
		Expired        time.Duration       `json:"expired"`
	}
}

func (req *GetBindCompanyKeyRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Count == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
	if req.Request.InvitationRole == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
	if req.Request.Issuer == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
}

// GetBindCompanyKeyResponse 生成绑定公司密钥
type GetBindCompanyKeyResponse struct {
	apiobj.BaseResponse
	Response struct {
		Key string `json:"key"`
	}
}

// BindCompanyRequest 绑定公司
type BindCompanyRequest struct {
	apiobj.BaseRequest
	Request struct {
		Key        string `json:"key"`
		Code       string `json:"code"`
		Way        string `json:"way"`
		DomainName string `json:"domain_name"`
		UserName   string `json:"username"`
		Password   string `json:"password"`
	}
}

func (req *BindCompanyRequest) Validity(resp *BindCompanyResponse) {
	switch req.Request.Way {
	case "wechat_web":
		if len(req.Request.Code) == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_parameters"
			return
		}
	case "password_web":
		if len(req.Request.UserName) == 0 || len(req.Request.Password) == 0 {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_parameters"
			return
		}
		if strings.Contains(req.Request.UserName, "@") {
			//? is email
			if err := validate.IsEmail(req.Request.UserName); err != nil {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_invalid_email"
				return
			}
		} else {
			//? is phone number
			if err := validate.IsPhone(req.Request.UserName); err != nil {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "account_invalid_phone_format"
				return
			}
		}
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters"
		return
	}
}

// BindCompanyResponse 绑定公司
type BindCompanyResponse struct {
	apiobj.BaseResponse
	Response struct {
		UserID uint `json:"user_id"`
		// success: 登陆成功
		// register: 需要注册
		// failed: 登陆失败
		LoginStatus string `json:"login_status"`
		// UserInfo 用户信息
		UserInfo *user.UserInfo `json:"user_info,omitempty"`
		// JwtToken jwt token
		JwtToken string `json:"jwt_token,omitempty"`
		// FailedReason 登陆失败原因
		FailedReason string `json:"failed_reason,omitempty"`
		// Uin 分类uin
		Uin []*LoginUin `json:"uin,omitempty"`
		// Issuer 颁发者
		Issuer string `json:"issuer,omitempty"`
		// 是否允许注册
		AllowRegister bool `json:"allow_register"`
		// RefreshToken 用来继续选择用户后的登录
		RefreshToken string `json:"refresh_token,omitempty"`
		// 登录方式
		LoginWay auth.LoginWay `json:"login_way"`
	}
}

// GetInviteInfoRequest 获取绑定公司信息请求
type GetInviteInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		Key string `json:"key"`
	}
}

type GetInviteInfoResponse struct {
	apiobj.BaseResponse
	Response struct {
		CompanyName string `json:"company_name"`
		InviterName string `json:"inviter_name"`
	}
}

func (req *GetInviteInfoRequest) Validity(resp *GetInviteInfoResponse) {
	if len(req.Request.Key) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
}
