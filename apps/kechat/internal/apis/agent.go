package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtoagent"
	"github.com/insmtx/corekg/apps/kechat/services/svcagent"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetLatestAgents 获取最近轻应用
// @Tags 轻应用
// @Summary 获取最近轻应用
// @Description 获取最近轻应用
// @Router /chat.GetLatestAgents [post]
// @Param request body dtoagent.GetLatestAgentsRequest true "request"
// @Success 200 {object} dtoagent.GetLatestAgentsResponse "response"
func GetLatestAgents(ctx *gin.Context, req *dtoagent.GetLatestAgentsRequest, resp *dtoagent.GetLatestAgentsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetLatestAgents] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcagent.GetLatestAgents(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetLatestAgents] svcagent.GetLatestAgents failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_query_agents_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
