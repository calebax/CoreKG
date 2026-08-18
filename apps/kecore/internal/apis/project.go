package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoproject"
	"github.com/insmtx/corekg/apps/kecore/services/svcproject"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetDefaultProject 获取默认项目
// @Tags 项目
// @Summary 获取默认项目
// @Description 获取默认项目
// @Router /forest.GetDefaultProject [post]
// @Param request body dtoproject.GetDefaultProjectRequest true "request"
// @Success 200 {object} dtoproject.GetDefaultProjectResponse "response"
func GetDefaultProject(ctx *gin.Context, req *dtoproject.GetDefaultProjectRequest, resp *dtoproject.GetDefaultProjectResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetDefaultProject] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.GetDefaultProject(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetDefaultProject] svcproject.GetDefaultProject failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_default_project_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListProjectItem 列出项目项
// @Tags 项目
// @Summary 列出项目项
// @Description 列出项目项
// @Router /forest.ListProjectItem [post]
// @Param request body dtoproject.ListProjectItemRequest true "request"
// @Success 200 {object} dtoproject.ListProjectItemResponse "response"
func ListProjectItem(ctx *gin.Context, req *dtoproject.ListProjectItemRequest, resp *dtoproject.ListProjectItemResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListProjectItem] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.ListProjectItem(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListProjectItem] svcproject.ListProjectItem failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_project_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetProjectItem 获取项目项
// @Tags 项目
// @Summary 获取项目项
// @Description 获取项目项
// @Router /forest.GetProjectItem [post]
// @Param request body dtoproject.GetProjectItemRequest true "request"
// @Success 200 {object} dtoproject.GetProjectItemResponse "response"
func GetProjectItem(ctx *gin.Context, req *dtoproject.GetProjectItemRequest, resp *dtoproject.GetProjectItemResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetProjectItem] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.GetProjectItem(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetProjectItem] svcproject.GetProjectItem failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_project_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
