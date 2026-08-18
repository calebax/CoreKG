package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListLoginSettingRequest apiobj.QueryRequest

func (req *ListLoginSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}

	for _, v := range req.Request.Filters {
		switch v.Field {
		case "domain_name", "env", "methods", "title":
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

type ListLoginSettingResponse struct {
	apiobj.BaseResponse

	Response login_setting.QueryLoginSettingListResponse
}

type CreateLoginSettingRequest struct {
	apiobj.BaseRequest

	Request *login_setting.LoginSetting
}

func (req *CreateLoginSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.DomainName == "" || req.Request.Path == "" || req.Request.Env == "" ||
		req.Request.Title == "" || req.Request.Issuer == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "字段不能为空"
		return
	}
	if req.Request.IsEnableWeChat && req.Request.AppID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "微信登录appid不能为空"
		return
	}
	if req.Request.IsEnableWeChatCom && req.Request.AppIDCom == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "企微登录appid不能为空"
		return
	}
}

type CreateLoginSettingResponse struct {
	apiobj.BaseResponse
}

type UpdateLoginSettingRequest struct {
	apiobj.BaseRequest

	Request *login_setting.LoginSetting
}

func (req *UpdateLoginSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.DomainName == "" || req.Request.Env == "" ||
		req.Request.Title == "" || req.Request.Issuer == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "字段不能为空"
		return
	}
	if req.Request.IsEnableWeChat && req.Request.AppID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "微信登录appid不能为空"
		return
	}
	if req.Request.IsEnableWeChatCom && req.Request.AppIDCom == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "企微登录appid不能为空"
		return
	}
}

type UpdateLoginSettingResponse struct {
	apiobj.BaseResponse
}

type DeleteLoginSettingRequest struct {
	apiobj.BaseRequest

	Request *login_setting.LoginSetting
}

func (req *DeleteLoginSettingRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "错误数据，id为0"
		return
	}
}

type DeleteLoginSettingResponse struct {
	apiobj.BaseResponse
}

type GetLoginSettingRequest struct {
	apiobj.BaseRequest

	Request struct {
		DomainName string `json:"domain_name"`
		Path       string `json:"path"`
	}
}

func (req *GetLoginSettingRequest) Validity(ctx *gin.Context, resp *apiobj.BaseResponse) {
	if req.Request.DomainName == "" {
		req.Request.DomainName = ctx.Request.Host
	}
}

type settingData struct {
	WeChat struct {
		Enable bool   `json:"enable"`
		AppID  string `json:"appid"`
	} `json:"wechat"`
	WeChatCom struct {
		Enable bool `json:"enable"`
		// 企微登录appid
		AppIDCom string `json:"appid_com"`
		// 企微agentid，自建项目使用
		AgentID string `json:"agentid"`
	} `json:"wechat_com"`
	Phone struct {
		Enable bool `json:"enable"`
	} `json:"phone"`
	Emali struct {
		Enable bool `json:"enable"`
	} `json:"email"`
	Password struct {
		Enable bool `json:"enable"`
	} `json:"password"`
	Title         string `json:"title"`
	ImageUrl      string `json:"image_url"`
	AllowRegister bool   `json:"allow_register"`
	Issuer        string `json:"issuer"`
	AuthKey       string `json:"auth_key"`
	LoginURL      string `json:"login_url"`
}

type GetLoginSettingResponse struct {
	apiobj.BaseResponse
	Response settingData
}

type GetLoginSettingByIDRequest struct {
	apiobj.BaseRequest

	Request struct {
		ID uint `json:"id"`
	}
}

func (req *GetLoginSettingByIDRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "ID不能为0"
		return
	}
}

type GetLoginSettingByIDResponse struct {
	apiobj.BaseResponse
	Response struct {
		LoginSetting *login_setting.LoginSetting
	}
}
