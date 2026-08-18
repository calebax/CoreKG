package accountmds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/insmtx/corekg/apps/kechat/models/chatagent"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/apipath"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

// MustLoginAPIKey 确保APIKey登录
func MustLoginAPIKey(ctx *gin.Context) (ls *auth.LoginStatus, apiKeyID uint, apiPath string, err error) {
	ls = ctx.MustGet(global.CtxKeyLoginStatus).(*auth.LoginStatus)
	if ls.State != auth.StateSucc {
		logs.ErrorContextf(ctx, "[privilege_auth] unauthorized, %+v", ls)
		err = errors.New("unauthorized")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	apiPath = apipath.ExtractAPIFromRequestURI(ctx.Request.RequestURI)
	apiKeyID = ls.GetID(global.CtxKeyAPIKeyID)
	if apiKeyID == 0 || apiPath == "" {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] API key id is 0 or apiPath is empty")
		err = errors.New("unauthorized")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}
	logs.InfoContextf(ctx, "[MustLoginAPIKey] apiPath: %s,apiKeyID:%s", apiPath, apiKeyID)
	return
}

// VerifyAPIKey 验证APIKey有效性
func VerifyAPIKey(ctx context.Context, apiKeyID uint) (*accounttype.APIKey, error) {
	keyInfo, err := apikey.GetApiKeyByID(ctx, apiKeyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[VerifyAPIKey]apikey.GetApiKeyByID(%d) failed: %v", apiKeyID, err)
		return nil, errors.New("apikey无效")
	}

	if keyInfo.IsExpired() {
		logs.ErrorContextf(ctx, "[VerifyAPIKey]apikey(id:%v|key:%v) expired at (%v)", apiKeyID, keyInfo.APIKey, keyInfo.ExpiredAt)
		return nil, fmt.Errorf("apikey[key:%v]已过期", keyInfo.APIKey)
	}
	return keyInfo, nil
}

// VerifyAPIKeyAgentPrivilege 验证APIKey是否有Agent调用权限
func VerifyAPIKeyAgentPrivilege(ctx *gin.Context, keyInfo *accounttype.APIKey, body io.Reader) bool {
	bts, err := io.ReadAll(body)
	if err != nil {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] read request body failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &apiobj.BaseResponse{
			Code: http.StatusBadRequest,
		})
		return false
	}

	var req *agentclient.ChatRequestBody
	if err = json.Unmarshal(bts, &req); err != nil {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] unmarshal request failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &apiobj.BaseResponse{
			Code: http.StatusInternalServerError,
		})
		return false
	}
	ag, err := chatagent.GetAgentDetailByName(ctx, req.Model)
	if err != nil {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] get agent detail failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &apiobj.BaseResponse{
			Code: http.StatusInternalServerError,
		})
		return false
	}
	if ag.CompanyID != keyInfo.CompanyID {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] apikey[%v]permission denied for model %s", keyInfo.APIKey, req.Model)
		ctx.AbortWithStatusJSON(http.StatusForbidden, &apiobj.BaseResponse{
			Code: http.StatusForbidden,
		})
		return false
	}

	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bts))
	return true
}

// VerifyAPIKeyAPIPrivilege 验证APIKey是否有调用某个API的权限
func VerifyAPIKeyAPIPrivilege(ctx *gin.Context, keyInfo *accounttype.APIKey, apiPath string) bool {
	requiredPriv, err := apikey.GetAPIPrivilegeByAPI(ctx, apiPath)
	if err != nil {
		logs.ErrorContextf(ctx, "[RequireAPIKeyPrivilege] get API privilege failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusForbidden, &apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
		})
		return false
	}

	hasAccess, err := apikey.GetApiKeyPrivilegeByAPIKeyIDAndAPIID(ctx, keyInfo.ID, requiredPriv.ID)
	if err != nil || !hasAccess {
		logs.WarnContextf(ctx, "[RequireAPIKeyPrivilege] API key [id:%d|key:%v] unauthorized for %s", keyInfo.ID, keyInfo.APIKey, apiPath)
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
		})
		return false
	}
	return true
}
