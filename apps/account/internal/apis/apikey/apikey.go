package apikey

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
)

// CreateAgentApiKey 创建agentKey
// @Tags 轻应用APIKEY
// @Summary 创建agentKey
// @Description 创建agentKey
// @Router /account.CreateAgentApiKey [post]
// @Param user body CreateAgentApiKeyRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func CreateAgentApiKey(ctx *gin.Context, req *CreateAgentApiKeyRequest, resp *apiobj.BaseResponse) {
	if req.Valid(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreateAgentApiKey validate params failed")
		return
	}
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)

	ag, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		logs.ErrorContextf(ctx, "chat_agent.GetChatAgentByID(%d) failed: %v", req.Request.AgentID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_agent_info_failed" // 获取agent信息失败
		return
	}

	// 检查当前用户是否是管理员或机器人所有者
	if !perm.HasManageAct(ctx, uin, ag.ID, foresttype.ResourceTypeAgent) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, ag.ID)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_no_permission_update_resource")) // 无权限更新此资源
		return
	}

	if len(req.Request.Name) == 0 {
		req.Request.Name = fmt.Sprintf("%v_%v", ag.ShowName, random.String(6))
	}

	key := &accounttype.APIKey{
		CompanyID:    companyID,
		Uin:          uin,
		Name:         req.Request.Name,
		APIKey:       apikey.GenerateSecretKey(),
		Purpose:      fmt.Sprintf("agent-%v-api", ag.ShowName),
		ResourceType: accounttype.ResourceTypeAgent,
		ResourceID:   ag.ID,
		Status:       accounttype.AccessKeyStatusNormal,
	}

	if req.Request.Expire == 0 {
		//fix: if not have explicit expire then set it to nil, that mean unlimited
		key.ExpiredAt = nil
	} else {
		exp := time.Now().Add(time.Duration(req.Request.Expire) * 24 * time.Hour)
		key.ExpiredAt = &exp
	}
	if err = dbutil.Account().Create(&key).Error; err != nil {
		logs.ErrorContextf(ctx, "create apikey failed: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_create_apikey_failed")) // 创建apikey失败
	}
}

// DeleteAgentApikey 删除agentKey
// @Tags 轻应用APIKEY
// @Summary 删除agentKey
// @Description 删除agentKey
// @Router /account.DeleteAgentApikey [post]
// @Param user body DeleteAgentApikeyRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func DeleteAgentApikey(ctx *gin.Context, req *DeleteAgentApikeyRequest, resp *apiobj.BaseResponse) {
	if req.Valid(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "DeleteAgentApikey validate params failed")
		return
	}

	uin := runtime.Uin(ctx)

	ag, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		logs.ErrorContextf(ctx, "chat_agent.GetChatAgentByID(%d) failed: %v", req.Request.AgentID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_agent_info_failed" // 获取agent信息失败
		return
	}

	if !perm.HasManageAct(ctx, uin, ag.ID, foresttype.ResourceTypeAgent) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, ag.ID)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_no_permission_update_resource")) // 无权限更新此资源
		return
	}

	apk, err := apikey.GetAgentApikey(ag.ID, req.Request.ApikeyID)
	if err != nil {
		logs.ErrorContextf(ctx, "apikey.GetApiKeyByID(%d) failed: %v", req.Request.ApikeyID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_apikey_failed")) // 获取apikey失败
		return
	}

	if err = apikey.DeleteApiKey(ctx, apk.APIKey); err != nil {
		logs.ErrorContextf(ctx, "apikey.DeleteApiKey(%d) failed: %v", req.Request.ApikeyID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_delete_apikey_failed")) // 删除apikey失败
	}
}

// ListAgentAPIKey APIkey列表
// @Tags 轻应用APIKEY
// @Summary APIkey列表
// @Description APIkey列表
// @Router /account.ListAgentAPIKey [post]
// @Param user body ListAgentAPIKeyRequest true "入参"
// @Success 200 {object} ListAgentAPIKeyResponse "返回值"
func ListAgentAPIKey(ctx *gin.Context, req *ListAgentAPIKeyRequest, resp *ListAgentAPIKeyResponse) {
	if req.Valid(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "ListAgentAPIKey validate params failed")
		return
	}
	uin := runtime.Uin(ctx)

	req.Request.Uin = uin
	req.Request.CompanyID = runtime.CompanyID(ctx)
	if err := apikey.QueryAgentApiKeyList(ctx, req.Request.PageQuery, &resp.Response); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_query_apikey_list_failed" // 查询APIkey列表失败
		logs.WarnContextf(ctx, "QueryAgentApiKeyList failed ,err %s", err)
		return
	}
}

// SetAgentApiKeyStatus 设置apikey状态
// @Tags 轻应用APIKEY
// @Summary 设置apikey状态
// @Description 设置apikey状态
// @Router /account.SetAgentApiKeyStatus [post]
// @Param user body SetAgentApiKeyStatusRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func SetAgentApiKeyStatus(ctx *gin.Context, req *SetAgentApiKeyStatusRequest, resp *apiobj.BaseResponse) {
	if req.Valid(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "SetAgentApiKeyStatus validate params failed")
		return
	}
	uin := runtime.Uin(ctx)

	ag, err := chatagent.GetChatAgentByID(ctx, req.Request.AgentID)
	if err != nil {
		logs.ErrorContextf(ctx, "chat_agent.GetChatAgentByID(%d) failed: %v", req.Request.AgentID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_agent_info_failed" // 获取agent信息失败
		return
	}

	if !perm.HasManageAct(ctx, uin, ag.ID, foresttype.ResourceTypeAgent) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, ag.ID)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_no_permission_update_resource")) // 无权限更新此资源
		return
	}

	apk, err := apikey.GetAgentApikey(ag.ID, req.Request.ApikeyID)
	if err != nil {
		logs.ErrorContextf(ctx, "apikey.GetApiKeyByID(%d) failed: %v", req.Request.ApikeyID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_apikey_failed")) // 获取apikey失败
		return
	}

	if apk.Status != req.Request.Status {
		apk.Status = req.Request.Status
		if err := dbutil.Account().Save(apk).Error; err != nil {
			logs.ErrorContextf(ctx, "dbutil.AccountSave failed: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_update_apikey_status_failed")) // 修改apikey状态失败
		}
	}
}
