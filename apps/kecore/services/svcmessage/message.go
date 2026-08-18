package svcmessage

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomessage"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func ListMessage(ctx *gin.Context, req *dtomessage.ListMessageRequest) (res *dtomessage.ListMessageResponse, err error) {
	res = &dtomessage.ListMessageResponse{}
	messageList, c, err := forest.NewKeUinMessageDao().GetPageListByCond(ctx, &forest.KeUinMessageCond{
		BaseCond: forest.BaseCond{
			Uin:       req.Request.Uin,
			Offset:    req.Request.Offset,
			Limit:     req.Request.Limit,
			OrderBy:   req.Request.OrderBy,
			BeginTime: req.Request.BeginTime,
			EndTime:   req.Request.EndTime,
		},
		Filters: req.Request.Filters,
	})
	if err != nil {
		return nil, err
	}

	res.Response.Data = make([]dtomessage.Message, len(messageList))
	for i, message := range messageList {
		res.Response.Data[i] = dtomessage.Message{
			ID:           message.ID,
			CreatedAt:    message.CreatedAt,
			Title:        message.Title,
			Content:      message.Content,
			TemplateType: message.TemplateType,
			RoutePath:    message.RoutePath,
			ReadStatus:   message.ReadStatus,
			ReadAt:       message.ReadAt,
		}
	}
	res.Response.QueryResponse.Total = c
	res.Response.QueryResponse.Limit = req.Request.Limit
	res.Response.QueryResponse.Offset = req.Request.Offset

	return res, nil
}

func SetMessageStatus(ctx *gin.Context, req *dtomessage.SetMessageStatusRequest) (res *dtomessage.SetMessageStatusResponse, err error) {
	res = &dtomessage.SetMessageStatusResponse{}
	readAt := time.Now()

	mes := &foresttype.KeUinMessage{
		ReadStatus: req.Request.Status,
	}

	if mes.ReadStatus == foresttype.MessageReadStatusRead {
		mes.ReadAt = &readAt
	}

	if err := forest.NewKeUinMessageDao().UpdateByID(ctx, req.Request.MessageID, mes); err != nil {
		logs.ErrorContextf(ctx, "[SetMessageStatus] update message status fail, message id: %d, status: %s, err: %v", req.Request.MessageID, req.Request.Status, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_set_message_status_fail"
		return res, err
	}
	return res, nil
}

func DeleteMessages(ctx *gin.Context, req *dtomessage.DeleteMessagesRequest) (res *dtomessage.DeleteMessagesResponse, err error) {
	res = &dtomessage.DeleteMessagesResponse{}

	if req.Request.DeleteAll {
		if err := forest.NewKeUinMessageDao().DeleteByUin(ctx, runtime.Uin(ctx)); err != nil {
			logs.ErrorContextf(ctx, "[DeleteMessages] delete all messages fail, uin: %d, err: %v", runtime.Uin(ctx), err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_delete_all_messages_fail"
			return res, err
		}
		return res, nil
	}

	if err := forest.NewKeUinMessageDao().DeleteByIDs(ctx, req.Request.MessageIDs); err != nil {
		logs.ErrorContextf(ctx, "[DeleteMessages] delete messages fail, message ids: %v, err: %v", req.Request.MessageIDs, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_delete_messages_fail"
		return res, err
	}

	return res, nil
}

func GetMessageCount(ctx *gin.Context, req *dtomessage.GetMessageCountRequest) (res *dtomessage.GetMessageCountResponse, err error) {
	res = &dtomessage.GetMessageCountResponse{}
	count, err := forest.NewKeUinMessageDao().CountByCond(ctx, &forest.KeUinMessageCond{
		BaseCond: forest.BaseCond{
			Uin: req.Request.Uin,
		},
		Filters: req.Request.Filters,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetMessageCount] get message count fail, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_get_message_count_fail"
		return res, err
	}
	res.Response.Count = count

	return res, nil
}
