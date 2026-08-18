package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtosession"
	"github.com/insmtx/corekg/apps/kechat/services/svcsession"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// MoveSession 移动项目
// @Tags 项目
// @Summary 移动项目
// @Description 移动项目
// @Router /kecore.MoveSession [post]
// @Param request body dtosession.MoveSessionRequest true "request"
// @Success 200 {object} dtosession.MoveSessionResponse "response"
func MoveSession(ctx *gin.Context, req *dtosession.MoveSessionRequest, resp *dtosession.MoveSessionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[MoveSession] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcsession.MoveSession(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[MoveSession] svcsession.MoveSession failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_session_failed" // 更新会话失败
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListFreeSessions 列出空闲会话
// @Tags 会话
// @Summary 列出空闲会话
// @Description 列出空闲会话
// @Router /kechat.ListFreeSessions [post]
// @Param request body dtosession.ListFreeSessionsRequest true "request"
// @Success 200 {object} dtosession.ListFreeSessionsResponse "response"
func ListFreeSessions(ctx *gin.Context, req *dtosession.ListFreeSessionsRequest, resp *dtosession.ListFreeSessionsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListFreeSessions] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svcsession.ListFreeSessions(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListFreeSessions] svcsession.ListFreeSessions failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
