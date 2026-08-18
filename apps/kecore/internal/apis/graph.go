package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtograph"
	"github.com/insmtx/corekg/apps/kecore/services/svcgraph"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateForestGraph 获取默认图谱
// @Tags 知识图谱管理
// @Summary 获取默认图谱
// @Description 获取默认图谱
// @Router /forest.CreateForestGraph [post]
// @Param request body dtograph.CreateForestGraphRequest true "request"
// @Success 200 {object} dtograph.CreateForestGraphResponse "response"
func CreateForestGraph(ctx *gin.Context, req *dtograph.CreateForestGraphRequest, resp *dtograph.CreateForestGraphResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateForestGraph] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcgraph.CreateForestGraph(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateForestGraph] svcgraph.CreateForestGraph failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建或修改图谱失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
