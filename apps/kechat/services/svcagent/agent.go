package svcagent

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtoagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func GetLatestAgents(ctx *gin.Context, req *dtoagent.GetLatestAgentsRequest) (res *dtoagent.GetLatestAgentsResponse, err error) {
	res = &dtoagent.GetLatestAgentsResponse{}
	//inject data
	req.Request.Uin = runtime.Uin(ctx)
	req.Request.CompanyID = runtime.CompanyID(ctx)
	//
	ags, pg, err := chatagent.GetLatestAgent(ctx, req.Request)
	if err != nil {
		logs.ErrorContextf(ctx, "GetLatestAgents(uin:%d) err:%v", runtime.Uin(ctx), err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kechat_query_agents_failed"
		return
	}
	res.Response.Data = ags
	res.Response.QueryResponse = *pg
	return res, nil
}
