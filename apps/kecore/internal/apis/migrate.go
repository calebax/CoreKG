package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomigrate"
	"github.com/insmtx/corekg/apps/kecore/services/svcmigrate"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// MigrateInterface 业务迁移接口
// @Tags 迁移
// @Summary 业务迁移接口
// @Description 业务迁移接口
// @Router /forest.MigrateInterface [post]
// @Param request body dtomigrate.MigrateInterfaceRequest true "request"
// @Success 200 {object} dtomigrate.MigrateInterfaceResponse "response"
func MigrateInterface(ctx *gin.Context, req *dtomigrate.MigrateInterfaceRequest, resp *dtomigrate.MigrateInterfaceResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[MigrateInterface] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcmigrate.MigrateInterface(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[MigrateInterface] svcmigrate.MigrateInterface failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_migrate_interface_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
