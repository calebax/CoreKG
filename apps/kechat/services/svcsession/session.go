package svcsession

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtosession"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/ygpkg/yg-go/apis/errcode"
)

func MoveSession(ctx *gin.Context, req *dtosession.MoveSessionRequest) (res *dtosession.MoveSessionResponse, err error) {
	res = &dtosession.MoveSessionResponse{}
	s, err := chat.NewChatSessionsDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kechat_query_session_failed" // 查询会话失败
		return
	}
	if s.ID == 0 {
		res.Code = errcode.ErrCode_NotFound
		res.Message = "kechat_session_not_found" // 会话不存在
		return
	}

	err = chat.NewChatSessionsDao().UpdateMap(ctx, req.Request.ID, map[string]interface{}{
		"subject_id": req.Request.SubjectID,
	})
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kechat_update_session_failed" // 更新会话失败
		return
	}

	return res, nil
}

func ListFreeSessions(ctx *gin.Context, req *dtosession.ListFreeSessionsRequest) (res *dtosession.ListFreeSessionsResponse, err error) {
	res = &dtosession.ListFreeSessionsResponse{}
	return res, nil
}
