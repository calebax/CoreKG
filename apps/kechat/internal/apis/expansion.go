package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochat"
	"github.com/insmtx/corekg/apps/kechat/services/svcchat"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ExpansionQuestion 问题扩写
// @Tags 问题扩写
// @Summary 问题扩写
// @Description 问题扩写
// @Router /chat.ExpansionQuestion [post]
// @Param request body dtochat.ExpansionQuestionRequest true "request"
// @Success 200 {object} dtochat.ExpansionQuestionResponse "response"
func ExpansionQuestion(ctx *gin.Context, req *dtochat.ExpansionQuestionRequest, resp *dtochat.ExpansionQuestionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ExpansionQuestion] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcchat.ExpansionQuestion(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ExpansionQuestion] svcchat.ExpansionQuestion failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "改写问题失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
