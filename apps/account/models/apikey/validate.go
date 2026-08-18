package apikey

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/apipath"
)

func ValidateAPIKey(ctx *gin.Context, key string) (*accounttype.APIKey, error) {
	// 验证API Key是否存在
	keyInfo, err := GetApiKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("API key not found or invalid: %v", err)
	}

	// 从请求路径中提取API端点,移除版本前缀
	apiPath := apipath.ExtractAPIFromRequestURI(ctx.Request.RequestURI)

	// 获取该API需要的权限
	requiredPrivilege, err := GetAPIPrivilegeByAPI(ctx, apiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to determine API privilege level: %v", err)
	}

	// 检查Key是否具有访问该API的权限
	hasAccess, err := GetApiKeyPrivilegeByAPIKeyIDAndAPIID(ctx, keyInfo.ID, requiredPrivilege.ID)
	if err != nil {
		return nil, fmt.Errorf("internal server error : %v", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("API key not authorized for this endpoint: %s", apiPath)
	}

	return keyInfo, nil
}
