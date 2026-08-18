package accountmds

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

// InjectLoginStatus 注入登录状态
func InjectLoginStatus(ctx *gin.Context, ls *auth.LoginStatus) (err error) {
	if ls.Role == auth.RoleAPI {
		return injectLoginStatusAPIKey(ctx, ls)
	}

	if ls.State != auth.StateSucc {
		return
	}

	if ls.Claim.Audience == global.AudienceUser {
		ls.Role = auth.RoleUser
	}
	ls.SetID(global.CtxKeyUin, ls.Claim.Uin)
	ls.SetID(global.CtxKeyAPIKeyID, 0)
	// 获取用户信息
	user, err := user.GetUserIdentificationByUIN(ctx, ls.Claim.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[manager_auth] [%d] get user info failed, %v", ls.Claim.Uin, err)
		ls.Err = fmt.Errorf("user status is not normal")
		ls.State = auth.StateFailed
		return
	}
	// 判断用户状态
	if user.UinStatus != accounttype.UinStatusNormal {
		logs.WarnContextf(ctx, "[manager_auth][%s] user status is not normal, %s", user.ID, user.UinStatus)
		ls.Err = fmt.Errorf("user status is not normal, %s", user.UinStatus)
		ls.State = auth.StateFailed
		return
	}
	// 根据用户类型设置上下文键
	switch user.SubjectType {
	case global.SubjectTypeIndividual:
		ls.SetID(global.CtxKeyCompanyID, 0)
	case global.SubjectTypeCompany:
		ls.SetID(global.CtxKeyCompanyID, user.SubjectID)
		em, err := employee.GetEmployeeByUin(user.ID)
		if err != nil {
			ls.State = auth.StateFailed
			logs.ErrorContextf(ctx, "[manager_auth] [%d] get employee info failed, %v", user.ID, err)
			return err
		}
		ls.SetID(global.CtxKeyEmployeeID, em.ID)
	default:
		return fmt.Errorf("failed to get user type: %w", err)
	}

	return
}

func injectLoginStatusAPIKey(ctx *gin.Context, ls *auth.LoginStatus) error {
	if ls == nil {
		return fmt.Errorf("loginStatus is nil")
	}

	// 初始化状态和ID
	ls.SetID(global.CtxKeyUin, 0)
	ls.SetID(global.CtxKeyCompanyID, 0)
	ls.SetID(global.CtxKeyAPIKeyID, 0)
	// 默认状态为失败
	ls.State = auth.StateFailed

	// 检查 Authorization 头
	authHeader := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if authHeader == "" {
		logs.ErrorContextf(ctx, "missing or invalid Authorization header")
		return fmt.Errorf("missing or invalid Authorization header")
	}
	apikey, err := apikey.GetAPIKeyInfo(ctx, authHeader)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to validate API key: %v", err)
		return fmt.Errorf("invalid API key: %w", err)
	}
	if apikey.Status != accounttype.AccessKeyStatusNormal {
		logs.ErrorContextf(ctx, "API key status is not active: %s", apikey.Status)
		return fmt.Errorf("API key status is not active")
	}

	ls.SetID(global.CtxKeyAPIKeyID, apikey.ID)
	ls.SetID(global.CtxKeyUin, apikey.Uin)
	ls.SetID(global.CtxKeyCompanyID, apikey.CompanyID)
	ls.State = auth.StateSucc
	return nil
}

// getAuKey 在请求头中获取用户的key
func getAuKey(ctx *gin.Context) (string, error) {
	// 获取 Authorization 请求头
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	if len(authHeader) == 0 {
		return "", fmt.Errorf("missing Authorization header")
	}
	return authHeader, nil
}
