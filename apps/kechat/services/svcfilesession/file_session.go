package svcfilesession

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtofilesession"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func GetFileSession(ctx *gin.Context, req *dtofilesession.GetFileSessionRequest) (res *dtofilesession.GetFileSessionResponse, err error) {
	res = &dtofilesession.GetFileSessionResponse{}
	uin := runtime.Uin(ctx)
	proj, err := forest.NewKeProjectDao().GetByCond(ctx, &forest.KeProjectCond{
		Uin:             uin, //只筛选单文档类型
		ProjectTypeList: []foresttype.ProjectType{foresttype.ProjectTypeForestQA},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetFileSession] forest.NewKeProjectDao().CountByCond failed, err: %v", err)
		return res, err
	}

	sess, err := chat.NewChatSessionsDao().GetByCond(ctx, &chat.ChatSessionsCond{
		FileID:    req.Request.FileID,
		SubjectID: proj.ID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetFileSession] chat.NewChatSessionsDao().CountByCond failed, err: %v", err)
		return res, err
	}
	if sess.ID <= 0 {
		logs.DebugContextf(ctx, "Not found chat session, file_id:%d, subject_id:%d", req.Request.FileID, proj.ID)
		//does not need to do anything just return and let frontend to NewQuestion with NewChatSession
	} else {
		res.Response.Session = sess
	}

	return res, nil
}
