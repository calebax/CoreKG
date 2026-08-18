package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtostatistics"
	"github.com/insmtx/corekg/apps/kechat/services/devstatistics"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetAgentQuestionExcel 获取agent问题excel统计文件
// @Tags 统计
// @Summary 获取agent问题excel统计文件
// @Description 获取agent问题excel统计文件
// @Router /chat.GetAgentQuestionExcel [post]
// @Param request body dtostatistics.GetAgentQuestionExcelRequest true "request"
// @Success 200 {object} dtostatistics.GetAgentQuestionExcelResponse "response"
func GetAgentQuestionExcel(ctx *gin.Context, req *dtostatistics.GetAgentQuestionExcelRequest, resp *dtostatistics.GetAgentQuestionExcelResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetAgentQuestionExcel] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := devstatistics.GetAgentQuestionExcel(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAgentQuestionExcel] devstatistics.GetAgentQuestionExcel failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "生成失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetAgentQuestionCount 获取agent问题数量
// @Tags 统计
// @Summary 获取agent问题数量
// @Description 获取agent问题数量
// @Router /chat.GetAgentQuestionCount [post]
// @Param request body dtostatistics.GetAgentQuestionCountRequest true "request"
// @Success 200 {object} dtostatistics.GetAgentQuestionCountResponse "response"
func GetAgentQuestionCount(ctx *gin.Context, req *dtostatistics.GetAgentQuestionCountRequest, resp *dtostatistics.GetAgentQuestionCountResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetAgentQuestionCount] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := devstatistics.GetAgentQuestionCount(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAgentQuestionCount] svcstatistics.GetAgentQuestionCount failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "统计失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
