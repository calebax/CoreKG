package projectctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoproject"
	"github.com/insmtx/corekg/apps/kecore/services/svcproject"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateProject 创建项目
// @Tags 知识库项目管理
// @Summary 创建项目
// @Description 创建项目
// @Router /forest.CreateProject [post]
// @Param request body dtoproject.CreateProjectRequest true "request"
// @Success 200 {object} dtoproject.CreateProjectResponse "response"
func CreateProject(ctx *gin.Context, req *dtoproject.CreateProjectRequest, resp *dtoproject.CreateProjectResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateProject] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.CreateProject(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateProject] svcproject.CreateProject failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_project_failed" // 创建失败
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetProjectInfo 获取项目详情
// Deprecated: 该接口已废弃，使用 /forest.GetProjectItem 接口
// @Tags 知识库项目管理
// @Summary 获取项目详情
// @Description 获取项目详情
// @Router /forest.GetProjectInfo [post]
// @Param request body dtoproject.GetProjectInfoRequest true "request"
// @Success 200 {object} dtoproject.GetProjectInfoResponse "response"
func GetProjectInfo(ctx *gin.Context, req *dtoproject.GetProjectInfoRequest, resp *dtoproject.GetProjectInfoResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetProjectInfo] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.GetProjectInfo(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetProjectInfo] svcproject.GetProjectInfo failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_project_failed" // 查询失败
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteProject 删除项目
// @Tags 知识库项目管理
// @Summary 删除项目
// @Description 删除项目
// @Router /forest.DeleteProject [post]
// @Param request body dtoproject.DeleteProjectRequest true "request"
// @Success 200 {object} dtoproject.DeleteProjectResponse "response"
func DeleteProject(ctx *gin.Context, req *dtoproject.DeleteProjectRequest, resp *dtoproject.DeleteProjectResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteProject] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.DeleteProject(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteProject] svcproject.DeleteProject(id:%v) failed, err: %v", req.Request.ID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_delete_project_failed" // 删除项目失败
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// RenameProject 项目重命名
// Deprecated: 该接口已废弃，使用 /forest.ListProjectItem 接口
// @Tags 知识库项目管理
// @Summary 项目重命名
// @Description 项目重命名
// @Router /forest.RenameProject [post]
// @Param request body dtoproject.RenameProjectRequest true "request"
// @Success 200 {object} dtoproject.RenameProjectResponse "response"
func RenameProject(ctx *gin.Context, req *dtoproject.RenameProjectRequest, resp *dtoproject.RenameProjectResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RenameProject] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.RenameProject(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[RenameProject] svcproject.RenameProject failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_rename_project_failed" // 项目重命名失败
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListProject 获取项目列表
// @Tags 知识库项目管理
// @Summary 获取项目列表
// @Description 获取项目列表
// @Router /forest.ListProject [post]
// @Param request body dtoproject.ListProjectRequest true "request"
// @Success 200 {object} dtoproject.ListProjectResponse "response"
func ListProject(ctx *gin.Context, req *dtoproject.ListProjectRequest, resp *dtoproject.ListProjectResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListProject] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcproject.ListProject(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListProject] svcproject.ListProject failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_list_project_failed" // 获取列表失败
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
