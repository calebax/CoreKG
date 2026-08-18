package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtohotwords"
	"github.com/insmtx/corekg/apps/kecore/services/svchotwords"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetHotWords 获取热词
// @Tags 热词管理
// @Summary 获取热词
// @Description 获取热词
// @Router /kecore.GetHotWords [post]
// @Param request body dtohotwords.GetHotWordsRequest true "request"
// @Success 200 {object} dtohotwords.GetHotWordsResponse "response"
func GetHotWords(ctx *gin.Context, req *dtohotwords.GetHotWordsRequest, resp *dtohotwords.GetHotWordsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetHotWords] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svchotwords.GetHotWords(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetHotWords] svchotwords.GetHotWords failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
