package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtofilesession"
	"github.com/insmtx/corekg/apps/kechat/services/svcfilesession"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetFileSession 获取文件会话
// @Tags 文件会话
// @Summary 获取文件会话
// @Description 获取文件会话
// @Router /chat.GetFileSession [post]
// @Param request body dtofilesession.GetFileSessionRequest true "request"
// @Success 200 {object} dtofilesession.GetFileSessionResponse "response"
func GetFileSession(ctx *gin.Context, req *dtofilesession.GetFileSessionRequest, resp *dtofilesession.GetFileSessionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetFileSession] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcfilesession.GetFileSession(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetFileSession] svcfilesession.GetFileSession failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
