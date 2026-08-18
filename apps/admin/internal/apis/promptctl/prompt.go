package promptctl

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtoprompt"
	"github.com/insmtx/corekg/apps/admin/services/svcprompt"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreatePrompt 创建 prompt 模板主记录+首版本
// @Tags Prompt管理
// @Summary 创建prompt模板
// @Description 创建prompt模板主记录+首版本
// @Router /admin.CreatePrompt [post]
// @Param request body dtoprompt.CreatePromptRequest true "request"
// @Success 200 {object} dtoprompt.CreatePromptResponse "response"
func CreatePrompt(ctx *gin.Context, req *dtoprompt.CreatePromptRequest, resp *dtoprompt.CreatePromptResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreatePrompt] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.CreatePrompt(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreatePrompt] svcprompt.CreatePrompt failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "create_prompt_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// AddPromptVersion 新增 prompt 模板版本
// @Tags Prompt管理
// @Summary 新增prompt版本
// @Description 新增prompt模板版本，含模板校验
// @Router /admin.AddPromptVersion [post]
// @Param request body dtoprompt.AddPromptVersionRequest true "request"
// @Success 200 {object} dtoprompt.AddPromptVersionResponse "response"
func AddPromptVersion(ctx *gin.Context, req *dtoprompt.AddPromptVersionRequest, resp *dtoprompt.AddPromptVersionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[AddPromptVersion] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.AddPromptVersion(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[AddPromptVersion] svcprompt.AddPromptVersion failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "add_prompt_version_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// SwitchPromptVersion 切换 prompt 模板生效版本
// @Tags Prompt管理
// @Summary 切换prompt版本
// @Description 切换主表latest_version_id
// @Router /admin.SwitchPromptVersion [post]
// @Param request body dtoprompt.SwitchPromptVersionRequest true "request"
// @Success 200 {object} dtoprompt.SwitchPromptVersionResponse "response"
func SwitchPromptVersion(ctx *gin.Context, req *dtoprompt.SwitchPromptVersionRequest, resp *dtoprompt.SwitchPromptVersionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[SwitchPromptVersion] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.SwitchPromptVersion(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[SwitchPromptVersion] svcprompt.SwitchPromptVersion failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "switch_prompt_version_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetPromptDetail 获取 prompt 模板详情
// @Tags Prompt管理
// @Summary 获取prompt详情
// @Description 获取模板详情+当前生效版本内容+variable_keys
// @Router /admin.GetPromptDetail [post]
// @Param request body dtoprompt.GetPromptDetailRequest true "request"
// @Success 200 {object} dtoprompt.GetPromptDetailResponse "response"
func GetPromptDetail(ctx *gin.Context, req *dtoprompt.GetPromptDetailRequest, resp *dtoprompt.GetPromptDetailResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetPromptDetail] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.GetPromptDetail(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetPromptDetail] svcprompt.GetPromptDetail failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "get_prompt_detail_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListPromptVersions 获取 prompt 模板版本列表
// @Tags Prompt管理
// @Summary 版本列表
// @Description 获取模板全部版本列表
// @Router /admin.ListPromptVersions [post]
// @Param request body dtoprompt.ListPromptVersionsRequest true "request"
// @Success 200 {object} dtoprompt.ListPromptVersionsResponse "response"
func ListPromptVersions(ctx *gin.Context, req *dtoprompt.ListPromptVersionsRequest, resp *dtoprompt.ListPromptVersionsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListPromptVersions] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.ListPromptVersions(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListPromptVersions] svcprompt.ListPromptVersions failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "list_prompt_versions_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListPrompts 查询 prompt 模板列表
// @Tags Prompt管理
// @Summary 模板列表
// @Description 查询模板列表按app/group/code等筛选
// @Router /admin.ListPrompts [post]
// @Param request body dtoprompt.ListPromptsRequest true "request"
// @Success 200 {object} dtoprompt.ListPromptsResponse "response"
func ListPrompts(ctx *gin.Context, req *dtoprompt.ListPromptsRequest, resp *dtoprompt.ListPromptsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListPrompts] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.ListPrompts(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListPrompts] svcprompt.ListPrompts failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "list_prompts_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// EditPrompt 编辑 prompt 模板主记录
// @Tags Prompt管理
// @Summary 编辑prompt
// @Description 编辑模板主记录(name/status等)
// @Router /admin.EditPrompt [post]
// @Param request body dtoprompt.EditPromptRequest true "request"
// @Success 200 {object} dtoprompt.EditPromptResponse "response"
func EditPrompt(ctx *gin.Context, req *dtoprompt.EditPromptRequest, resp *dtoprompt.EditPromptResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[EditPrompt] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.EditPrompt(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[EditPrompt] svcprompt.EditPrompt failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "edit_prompt_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeletePrompt 删除 prompt 模板（软删）
// @Tags Prompt管理
// @Summary 删除prompt
// @Description 删除模板(软删)
// @Router /admin.DeletePrompt [post]
// @Param request body dtoprompt.DeletePromptRequest true "request"
// @Success 200 {object} dtoprompt.DeletePromptResponse "response"
func DeletePrompt(ctx *gin.Context, req *dtoprompt.DeletePromptRequest, resp *dtoprompt.DeletePromptResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeletePrompt] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.DeletePrompt(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeletePrompt] svcprompt.DeletePrompt failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "delete_prompt_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// RenderPromptPreview 渲染预览 prompt 模板
// @Tags Prompt管理
// @Summary 渲染预览
// @Description 传入prompt_value返回渲染后文本，支持草稿模式和已保存模式
// @Router /admin.RenderPromptPreview [post]
// @Param request body dtoprompt.RenderPromptPreviewRequest true "request"
// @Success 200 {object} dtoprompt.RenderPromptPreviewResponse "response"
func RenderPromptPreview(ctx *gin.Context, req *dtoprompt.RenderPromptPreviewRequest, resp *dtoprompt.RenderPromptPreviewResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[RenderPromptPreview] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svcprompt.RenderPromptPreview(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[RenderPromptPreview] svcprompt.RenderPromptPreview failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "render_prompt_preview_failed"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}
