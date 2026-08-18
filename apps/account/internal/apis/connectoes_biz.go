package apis

import (
	"time"

	"github.com/insmtx/corekg/pkgs/connectors"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// PreConnectRequest 预授权请求
type PreConnectRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 登录方式
		Provider    string `json:"provider"`
		RedirectURL string `json:"redirectUrl"`
	}
}

func (req *PreConnectRequest) Validity(resp *apiobj.BaseResponse) {
	if !connectors.IsProviderSupported(req.Request.Provider) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_provider_not_supported" // 登录方式不支持
		return
	}
	// TODO 校验redirectUrl白名单
	if req.Request.RedirectURL == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_missing_redirect_url_parameter" // 缺少redirectUrl参数
		return
	}
}

type PreConnectResponse struct {
	apiobj.BaseResponse
	Response struct {
		State string `json:"state,omitempty"`
	}
}

type ConnectBindCache struct {
	Provider    string `json:"provider"`
	Timestamp   int64  `json:"timestamp"`
	UinID       uint   `json:"uinId"`
	CompanyID   uint   `json:"companyId"`
	RedirectURL string `json:"redirectUrl"`
}

type ListBindingsRequest apiobj.BaseRequest

type ListBindingsResponse struct {
	apiobj.BaseResponse
	Response struct {
		// 系统支持的可绑定平台
		Supported []*connectors.ProviderInfo `json:"supported"`
		Bindings  []UserBinding              `json:"bindings"` // 用户绑定列表
	}
}

type UserBinding struct {
	ID       uint      `json:"id"`
	Provider string    `json:"provider"` // 平台标识
	Account  string    `json:"account"`  // 绑定的外部账号名/邮箱
	BoundAt  time.Time `json:"boundAt"`  // 绑定时间
	Valid    bool      `json:"valid"`    // 绑定是否有效（token 是否可用）
}

type UnbindRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 绑定ID
		ID uint `json:"id"`
	}
}

func (req *UnbindRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_missing_id_parameter" // 缺少id参数
		return
	}
}

type UnbindResponse struct {
	apiobj.BaseResponse
	Request struct {
		Success string `json:"success"`
	}
}
