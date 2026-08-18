package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomessage"
	"github.com/insmtx/corekg/apps/kecore/services/svcmessage"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ListMessage 获取消息记录
// @Tags 消息记录
// @Summary 获取消息记录
// @Description 获取消息记录
// @Router /forest.ListMessage [post]
// @Param request body dtomessage.ListMessageRequest true "request"
// @Success 200 {object} dtomessage.ListMessageResponse "response"
func ListMessage(ctx *gin.Context, req *dtomessage.ListMessageRequest, resp *dtomessage.ListMessageResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListMessage] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	req.Request.Uin = runtime.Uin(ctx)
	res, err := svcmessage.ListMessage(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListMessage] svcmessage.ListMessage failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_message_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// SetMessageStatus 设置消息状态
// @Tags 消息状态
// @Summary 设置消息状态
// @Description 设置消息状态
// @Router /forest.SetMessageStatus [post]
// @Param request body dtomessage.SetMessageStatusRequest true "request"
// @Success 200 {object} dtomessage.SetMessageStatusResponse "response"
func SetMessageStatus(ctx *gin.Context, req *dtomessage.SetMessageStatusRequest, resp *dtomessage.SetMessageStatusResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SetMessageStatus] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcmessage.SetMessageStatus(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SetMessageStatus] svcmessage.SetMessageStatus failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_set_message_status_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteMessages 删除消息
// @Tags 消息
// @Summary 删除消息
// @Description 删除消息
// @Router /forest.DeleteMessages [post]
// @Param request body dtomessage.DeleteMessagesRequest true "request"
// @Success 200 {object} dtomessage.DeleteMessagesResponse "response"
func DeleteMessages(ctx *gin.Context, req *dtomessage.DeleteMessagesRequest, resp *dtomessage.DeleteMessagesResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteMessages] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcmessage.DeleteMessages(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteMessages] svcmessage.DeleteMessages failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_messages_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetMessageCount 获取消息数量
// @Tags 消息
// @Summary 获取消息数量
// @Description 获取消息数量
// @Router /forest.GetMessageCount [post]
// @Param request body dtomessage.GetMessageCountRequest true "request"
// @Success 200 {object} dtomessage.GetMessageCountResponse "response"
func GetMessageCount(ctx *gin.Context, req *dtomessage.GetMessageCountRequest, resp *dtomessage.GetMessageCountResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetMessageCount] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	req.Request.Uin = runtime.Uin(ctx)
	res, err := svcmessage.GetMessageCount(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetMessageCount] svcmessage.GetMessageCount failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_message_count_fail"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
