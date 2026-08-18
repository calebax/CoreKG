package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatmigration"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// MigrateChatQuestion 迁移chat记录
// @Tags 迁移
// @Summary 迁移chat记录
// @Description 迁移chat记录
// @Router /chat.MigrateChatQuestion [post]
// @Param request body MigrateRequest true "入参"
// @Success 200 {object} MigrateResponse
func MigrateChatQuestion(ctx *gin.Context, req *MigrateRequest, resp *MigrateResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "MigrateChatQuestion validate Key: %s", req.Request.Key)
		return
	}
	err := chatmigration.MigrateChatQuestion(ctx)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_migrate_chat_question_failed" // 迁移chat记录失败
		logs.ErrorContextf(ctx, "MigrateChatQuestion error: %s", err.Error())
		return
	}
}

// MigrateForestChat 迁移forest记录
// @Tags 迁移
// @Summary 迁移forest记录
// @Description 迁移forest记录
// @Router /chat.MigrateForestChat [post]
// @Param request body MigrateRequest true "入参"
// @Success 200 {object} MigrateResponse
func MigrateForestChat(ctx *gin.Context, req *MigrateRequest, resp *MigrateResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "MigrateChatQuestion validate Key: %s", req.Request.Key)
		return
	}
	err := chatmigration.MigrateForestChat(ctx)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_migrate_forest_chat_failed" // 迁移forest记录失败
		logs.ErrorContextf(ctx, "MigrateForestChat error: %s", err.Error())
		return
	}
}

// MigrateRequest 请求结构体
type MigrateRequest struct {
	apiobj.BaseRequest
	Request struct {
		Key string `json:"key"`
	}
}

// Validity 验证有效性
func (req *MigrateRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Key != "ajhsdkajdssajkzb" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_key" // key错误
		return
	}
}

// MigrateResponse 响应结构体
type MigrateResponse struct {
	apiobj.BaseResponse
	Response struct{}
}
