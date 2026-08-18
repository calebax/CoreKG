package apis

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// CreateAPIKey 创建API key
// @Tags API KEY 管理
// @Summary 创建API key
// @Description 创建API key
// @Router /account.CreateAPIKey [post]
// @Param user body CreateAPIKeyRequest true "入参"
// @Success 200 {object} CreateAPIKeyResponse "返回值"
func CreateAPIKey(ctx *gin.Context, req *CreateAPIKeyRequest, resp *CreateAPIKeyResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreatAPIKey validate params failed")
		return
	}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	_, err := apikey.CreatAPIKey(ctx, uin, companyID, req.Request.Name, req.Request.Purpose, req.Request.ExpiredAt)
	if err != nil {
		logs.ErrorContextf(ctx, "CreatAPIKey CreateApiKey err： %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_api_key_failed" // 创建用户key失败
		return
	}
}

// ListAPIKey API key列表
// @Tags API KEY 管理
// @Summary API key列表
// @Description API key列表
// @Router /account.ListAPIKey [post]
// @Param user body ListAPIKeyRequest true "入参"
// @Success 200 {object} ListAPIKeyResponse "返回值"
func ListAPIKey(ctx *gin.Context, req *ListAPIKeyRequest, resp *ListAPIKeyResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListApiKey validate params failed")
		return
	}
	uin := runtime.Uin(ctx)

	req.Request.Uin = uin
	err := apikey.QueryListApiKey(ctx, req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_list_api_key_failed" // 查询API key列表失败
		logs.WarnContextf(ctx, "ListApiKey failed ,err %s", err)
		return
	}
}

// DeleteAPIKey 删除API key
// @Tags API KEY 管理
// @Summary 删除API key
// @Description 删除API key
// @Router /account.DeleteApiKey [post]
// @Param user body DeleteAPIKeyRequest true "入参"
// @Success 200 {object} DeleteAPIKeyResponse "返回值"
func DeleteAPIKey(ctx *gin.Context, req *DeleteAPIKeyRequest, resp *DeleteAPIKeyResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "DeleteAPIKey validate params failed")
		return
	}
	uin := runtime.Uin(ctx)
	// 校验apikey是否存在
	apiKey, err := apikey.GetApiKeyByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_api_key_not_exist" // API key不存在
		logs.WarnContextf(ctx, "DeleteAPIKeyByID CreateUserKey err： %v", err)
		return
	}
	if apiKey.Uin != uin {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_api_key_not_belong_user" // 此API Key不属于当前用户
		return
	}
	err = apikey.DeleteAPIKeyByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteAPIKeyByID CreateUserKey err： %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_delete_api_key_failed" // 删除用户key失败
		return
	}
}

// ListAPIKeyPrivilege API Key 权限列表
// @Tags API KEY 管理
// @Summary API Key 权限列表
// @Description API Key 权限列表
// @Router /account.ListAPIKeyPrivilege [post]
// @Param user body ListAPIKeyPrivilegeRequest true "入参"
// @Success 200 {object} ListAPIKeyPrivilegeResponse "返回值"
func ListAPIKeyPrivilege(ctx *gin.Context, req *ListAPIKeyPrivilegeRequest, resp *ListAPIKeyPrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListAPIKeyPrivilege validate params failed")
		return
	}
	err := apikey.QueryListApiKeyPrivilege(ctx, req.Request.PageQuery, req.Request.ID, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_server_error" // 服务器错误
		logs.WarnContextf(ctx, "ListAPIKeyPrivilege failed ,err %s", err)
		return
	}
}

// AddAPIKeyPrivilege 添加API key权限
// @Tags API KEY 管理
// @Summary 添加API key权限
// @Description 添加API key权限
// @Router /account.AddAPIKeyPrivilege [post]
// @Param user body AddAPIKeyPrivilegeRequest true "入参"
// @Success 200 {object} AddAPIKeyPrivilegeResponse "返回值"
func AddAPIKeyPrivilege(ctx *gin.Context, req *AddAPIKeyPrivilegeRequest, resp *AddAPIKeyPrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "AddApiKeyPrivilege validate params failed")
		return
	}

	IsExis, err := apikey.IsExistApiKeyPrivilege(ctx, req.Request.ID, req.Request.PrivilegeIDs)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_privilege_check_failed" // 权限校验失败
		logs.WarnContextf(ctx, "IsExistApiKeyPrivilege failed ,err %s", err)
		return
	}
	if IsExis {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_privilege_already_exist" // 权限已存在
		return
	}
	err = apikey.AddApiKeyPrivilege(req.Request.ID, req.Request.PrivilegeIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "AddApiKeyPrivilege err： %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_add_api_key_privilege_failed" // 添加用户key权限失败
		return
	}
}

// DeleteAPIKeyPrivilege 删除API key权限
// @Tags API KEY 管理
// @Summary 删除API key权限
// @Description 删除API key权限
// @Router /account.DeleteAPIKeyPrivilege [post]
// @Param user body DeleteAPIKeyPrivilegeRequest true "入参"
// @Success 200 {object} DeleteAPIKeyPrivilegeResponse "返回值"
func DeleteAPIKeyPrivilege(ctx *gin.Context, req *DeleteAPIKeyPrivilegeRequest, resp *DeleteAPIKeyPrivilegeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "DeleteApiKeyPrivilege validate params failed")
		return
	}
	err := apikey.DeleteApiKeyPrivilege(req.Request.KeyID, req.Request.PrivilegeIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteApiKeyPrivilege err： %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_delete_api_key_privilege_failed" // 删除用户key权限失败
		return
	}

}

// ForwardAuth 转发认证
// @Tags API KEY 管理
// @Summary 转发认证
// @Description 转发认证
// @Router /account.ForwardAuth [get]
func ForwardAuth(ctx *gin.Context) {
	ls := runtime.LoginStatus(ctx)
	reqAct := extractRequestAction(ctx.Request)
	if reqAct == nil {
		logs.ErrorContextf(ctx, "ForwardAuth extractRequestAction failed, nil")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	logs.DebugContextf(ctx, "ForwardAuth action: %s, %v", reqAct, ls)

}

// RequestAction 转发请求
type RequestAction struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Proto      string `json:"proto"`
	Method     string `json:"method"`
	URI        string `json:"uri"`
	RemoteAddr string `json:"remote_addr"`

	AuthToken string `json:"auth_token"`
}

func (a RequestAction) String() string {
	return fmt.Sprintf("%s %s://%s:%d%s (%s)", a.Method, a.Proto, a.Host, a.Port, a.URI, a.AuthToken)
}

// extractRequestAction 提取请求中的转发信息
func extractRequestAction(req *http.Request) *RequestAction {
	action := &RequestAction{
		Host:       req.Header.Get("X-Forwarded-Host"),
		Proto:      req.Header.Get("X-Forwarded-Proto"),
		Method:     req.Header.Get("X-Forwarded-Method"),
		URI:        req.Header.Get("X-Forwarded-Uri"),
		RemoteAddr: req.Header.Get("X-Forwarded-For"),
		AuthToken:  req.Header.Get("Authorization"),
	}

	if port := req.Header.Get("X-Forwarded-Port"); port != "" {
		action.Port, _ = strconv.Atoi(port)
	}

	return action
}
