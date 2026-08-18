package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/apikey"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// CreateCozePlugin 创建coze插件
// @Tags coze
// @Summary 创建coze插件
// @Description 创建coze插件
// @Router /chat.CreateCozePlugin [post]
// @Param request body CreateCozePluginRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func CreateCozePlugin(ctx *gin.Context, req *CreateCozePluginRequest, resp *apiobj.BaseResponse) {
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	agentInfo, err := chatagent.GetAgentDetail(ctx, req.Request.AgentID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取基本信息失败: %v", err)
		logs.ErrorContextf(ctx, "GetAgentInfo failed ,err %s", err)
		return
	}
	if agentInfo.PublishStatus == chattype.StatusDraft {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "轻应用未发布，暂不支持同步"
		return
	}

	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
		return
	}

	sessionKey := runtime.LoginStatus(ctx).Token
	space, _, err := coze.GetSpaceAPI(ctx, cozeUrl, sessionKey)
	if err != nil || space == "" {
		logs.ErrorContextf(ctx, "get coze space error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "get coze space error"
		return
	}
	var key string
	key, err = apikey.ListAPIKeyAgentID(ctx, companyID, req.Request.AgentID, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "list coze api key error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "list coze api key error"
		return
	}
	if key == "" {
		key, err = coze.CreateAPIKey(ctx, req.Request.AgentID, uin, companyID)
		if err != nil {
			logs.ErrorContextf(ctx, "create coze api key error, %s", err.Error())
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "create coze api key error"
			return
		}
	}
	pluginID, err := coze.CreatePlugin(ctx, key, space, agentInfo.ShowName, cozeUrl, sessionKey)
	if err != nil {
		logs.ErrorContextf(ctx, "create coze plugin error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "create coze plugin error"
		return
	}
	cozeAPI, err := coze.CreateCozeAPI(ctx, pluginID, cozeUrl, sessionKey)
	if err != nil {
		logs.ErrorContextf(ctx, "create coze api error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "create coze api error"
		return
	}

	err = coze.UpdateCozeAPI(ctx, pluginID, cozeAPI, cozeUrl, sessionKey, len(agentInfo.Params))
	if err != nil {
		logs.ErrorContextf(ctx, "update coze api error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "update coze api error"
		return
	}
	err = coze.DebugCozeAPI(ctx, pluginID, cozeAPI, agentInfo.Name, sessionKey, agentInfo.Params, cozeUrl)
	if err != nil {
		logs.ErrorContextf(ctx, "debug coze api error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "debug coze api error"
		return
	}
	err = coze.PublishPlugin(ctx, pluginID, sessionKey, cozeUrl)
	if err != nil {
		logs.ErrorContextf(ctx, "publish coze plugin error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "publish coze plugin error"
		return
	}
	cozeMapping := chattype.ChatCozeMapping{
		Uin:      uin,
		Type:     chattype.ChatTypeAgentApp,
		CoreKGID: req.Request.AgentID,
		CozeID:   pluginID,
	}
	err = chattype.CreateCozeMapping(ctx, &cozeMapping)
	if err != nil {
		logs.ErrorContextf(ctx, "create coze mapping error, %s", err.Error())
	}
}

// DeleteCozeMappingByCozeID 删除coze映射
// @Tags coze
// @Summary 删除coze映射
// @Description 删除coze映射
// @Router /chat.DeleteCozeMappingByCozeID [post]
// @Param request body DeleteCozeMappingByCozeIDRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func DeleteCozeMappingByCozeID(ctx *gin.Context, req *DeleteCozeMappingByCozeIDRequest, resp *apiobj.BaseResponse) {
	err := coze.DeleteCozeMappingByCozeID(ctx, req.Request.CozeID, req.Request.ChatType)
	if err != nil {
		logs.ErrorContextf(ctx, "delete coze mapping error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "delete coze mapping error"
		return
	}
}

// GetAgentMapping 获取智能体映射关系
// @Tags coze
// @Summary 获取智能体映射关系
// @Description 获取智能体映射关系
// @Router /chat.GetAgentMapping [post]
// @Param request body GetAgentMappingRequest true "入参"
// @Success 200 {object} GetAgentMappingResponse
func GetAgentMapping(ctx *gin.Context, req *GetAgentMappingRequest, resp *GetAgentMappingResponse) {
	agentMap, err := coze.GetAgentMapping(ctx, req.Request.WorkflowID, chattype.ChatTypeWorkflow)
	if err != nil {
		logs.ErrorContextf(ctx, "delete coze mapping error, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "delete coze mapping error"
		return
	}
	resp.Response.AgentID = agentMap.CoreKGID
}
