package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtodevtool"
	"github.com/insmtx/corekg/apps/kecore/services/svcdevtool"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// RewriteMarkdownURL 重写markdown中的url
// @Tags 开发工具
// @Summary 重写markdown中的url
// @Description 重写markdown中的url
// @Router /kecore.RewriteMarkdownURL [post]
// @Param request body dtodevtool.RewriteMarkdownURLRequest true "request"
// @Success 200 {object} dtodevtool.RewriteMarkdownURLResponse "response"
func RewriteMarkdownURL(ctx *gin.Context, req *dtodevtool.RewriteMarkdownURLRequest, resp *dtodevtool.RewriteMarkdownURLResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RewriteMarkdownURL] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svcdevtool.RewriteMarkdownURL(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[RewriteMarkdownURL] svcdevtool.RewriteMarkdownURL failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// StatAlgoMarkdown 统计算法markdown
// @Tags 开发工具
// @Summary 统计算法markdown
// @Description 统计算法markdown
// @Router /kecore.StatAlgoMarkdown [post]
// @Param request body dtodevtool.StatAlgoMarkdownRequest true "request"
// @Success 200 {object} dtodevtool.StatAlgoMarkdownResponse "response"
func StatAlgoMarkdown(ctx *gin.Context, req *dtodevtool.StatAlgoMarkdownRequest, resp *dtodevtool.StatAlgoMarkdownResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[StatAlgoMarkdown] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svcdevtool.StatAlgoMarkdown(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[StatAlgoMarkdown] svcdevtool.StatAlgoMarkdown failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
