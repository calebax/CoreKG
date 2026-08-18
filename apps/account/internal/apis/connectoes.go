package apis

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
	connectoModel "github.com/insmtx/corekg/apps/account/models/connectors"
	"github.com/insmtx/corekg/pkgs/connectors/tokenmgr"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

func PreConnect(ctx *gin.Context, req *PreConnectRequest, resp *PreConnectResponse) {
	if req.Validity(&resp.BaseResponse); resp.BaseResponse.Code != 0 {
		logs.WarnContextf(ctx, "PreConnect: validity failed,err = %v", resp.BaseResponse.Message)
		return
	}

	connectData := ConnectBindCache{
		Provider:    req.Request.Provider,
		Timestamp:   time.Now().Unix(),
		UinID:       runtime.Uin(ctx),
		CompanyID:   runtime.CompanyID(ctx),
		RedirectURL: req.Request.RedirectURL,
	}
	stateKey := uuid.NewString()
	rsKey := rdsKeyConnectStateKey(stateKey)
	err := redispool.SetJSON(rsKey, connectData, 45*time.Minute)
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(), "PreConnect: failed to set Redis key: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_internal_server_error" // 内部服务器错误
		return
	}
	resp.Response.State = stateKey
}

func Connect(ctx *gin.Context) {
	provider := ctx.Param("provider")
	state := ctx.Query("state")

	// 参数校验
	if provider == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    errcode.ErrCode_BadRequest,
			"message": "account_missing_provider_parameter", // 缺少provider参数
		})
		return
	}
	if state == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    errcode.ErrCode_BadRequest,
			"message": "account_missing_state_parameter", // 缺少state参数
		})
		return
	}

	// 从Redis获取缓存的连接数据
	var cachedData ConnectBindCache
	stateKey := rdsKeyConnectStateKey(state)
	err := redispool.GetJSON(stateKey, &cachedData)
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(),
			"Connect: failed to get cached data for state %s: %v", state, err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    errcode.ErrCode_BadRequest,
			"message": "account_invalid_state_parameter", // 无效的state参数
		})
		return
	}

	// 验证provider是否匹配
	if cachedData.Provider != provider {
		logs.WarnContextf(ctx, "Connect: provider mismatch, expected %s but got %s", cachedData.Provider, provider)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    errcode.ErrCode_BadRequest,
			"message": "account_provider_mismatch", // provider不匹配
		})
		return
	}

	// 续期
	redispool.SetJSON(stateKey, cachedData, 45*time.Minute)

	logs.InfoContextf(ctx, "Connect: provider %s validated successfully with state %s", provider, state)

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	q.Add("state", state)
	ctx.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
}

func Callback(ctx *gin.Context) {
	state := ctx.Query("state")
	if state == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    errcode.ErrCode_BadRequest,
			"message": "account_missing_state_parameter", // 缺少state参数
		})
		return
	}

	var cachedData ConnectBindCache
	stateKey := rdsKeyConnectStateKey(state)
	err := redispool.GetJSON(stateKey, &cachedData)
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(),
			"Callback: failed to get cached data for state %s: %v", state, err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    errcode.ErrCode_BadRequest,
			"message": "account_invalid_state_parameter", // 无效的state参数
		})
		return
	}

	q := ctx.Request.URL.Query()
	q.Add("provider", cachedData.Provider)
	ctx.Request.URL.RawQuery = q.Encode()

	// 完成OAuth2认证
	platformUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(), "Callback: failed to complete user auth: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    errcode.ErrCode_InternalError,
			"message": "account_oauth_failed", // OAuth2认证失败
		})
		return
	}
	_ = gothic.Logout(ctx.Writer, ctx.Request)

	// 关联token，写入
	createReq := &connectoModel.CreateExternalBindingReq{
		Uin:          cachedData.UinID,
		CompanyID:    cachedData.CompanyID,
		Platform:     tokenmgr.Platform(cachedData.Provider),
		Provider:     cachedData.Provider,
		ExternalID:   platformUser.UserID,
		Email:        platformUser.Email,
		Avatar:       platformUser.AvatarURL,
		AccessToken:  platformUser.AccessToken,
		RefreshToken: platformUser.RefreshToken,
		ExpiresIn:    platformUser.ExpiresAt,
	}

	// 创建外部绑定
	_, err = connectoModel.CreateExternalBinding(ctx.Request.Context(), createReq)
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(), "Callback: failed to create external binding: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    errcode.ErrCode_InternalError,
			"message": "account_binding_failed", // 绑定失败
		})
		return
	}

	redispool.Del(stateKey)
	ctx.Redirect(302, cachedData.RedirectURL)
}

func ListBindings(ctx *gin.Context, req *ListBindingsRequest, resp *ListBindingsResponse) {
	uin := runtime.Uin(ctx)
	bindingList, err := connectoModel.QueryBindings(ctx, &connectoModel.QueryExternalBindingReq{
		Uin: uin,
	})
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(), "ListBindings: failed to list external bindings: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_list_bindings_failed"
		return
	}

	resp.Response.Bindings = make([]UserBinding, 0, len(bindingList))
	for _, binding := range bindingList {
		resp.Response.Bindings = append(resp.Response.Bindings, UserBinding{
			ID:       binding.ID,
			Provider: binding.Provider,
			Account:  binding.Email,
			BoundAt:  binding.CreatedAt,
			Valid:    true,
		})
	}
	providers, err := connectoModel.ListSupportedProviders(ctx)
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(), "ListBindings: failed to list supported providers: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_list_providers_failed"
		return
	}
	resp.Response.Supported = providers
}

func Unbind(ctx *gin.Context, req *UnbindRequest, resp *UnbindResponse) {
	if req.Validity(&resp.BaseResponse); resp.BaseResponse.Code != 0 {
		logs.WarnContextf(ctx.Request.Context(), "request validity check failed: code=%d, message=%s",
			resp.BaseResponse.Code, resp.BaseResponse.Message)
		return
	}
	bindingID := req.Request.ID

	// 删除外部绑定
	err := connectoModel.DeleteExternalBinding(ctx.Request.Context(), &connectoModel.DeleteExternalBindingReq{
		ID: bindingID,
	})
	if err != nil {
		logs.ErrorContextf(ctx.Request.Context(), "Unbind: failed to delete external binding: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_unbind_failed"
		return
	}

	resp.Request.Success = "true"
}

func rdsKeyConnectStateKey(key string) string {
	return fmt.Sprintf("account_connect_state:%s", key)
}
