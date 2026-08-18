package dtoprompt

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/prompt/model"
)

// CreatePromptRequest 创建 prompt 模板主记录+首版本
type CreatePromptRequest struct {
	apiobj.BaseRequest
	Request CreatePromptEmbedRequest
}

// CreatePromptEmbedRequest 创建 prompt 的内嵌请求参数
type CreatePromptEmbedRequest struct {
	App          string         `json:"app"`
	Group        string         `json:"group"`
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	Content      string         `json:"content"`
	VariableKeys []model.VarKey `json:"variable_keys"`
}

// Validity 校验 CreatePromptRequest 参数
func (opt *CreatePromptRequest) Validity(resp *CreatePromptResponse) {
	if opt.Request.App == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "app_is_empty"
		return
	}
	if opt.Request.Group == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "group_is_empty"
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "name_is_empty"
		return
	}
	if opt.Request.Code == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "code_is_empty"
		return
	}
	if opt.Request.Content == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "content_is_empty"
		return
	}
	if len(opt.Request.VariableKeys) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "variable_keys_is_empty"
		return
	}
}

// AddPromptVersionRequest 新增 prompt 模板版本
type AddPromptVersionRequest struct {
	apiobj.BaseRequest
	Request AddPromptVersionEmbedRequest
}

// AddPromptVersionEmbedRequest 新增版本的内嵌请求参数
type AddPromptVersionEmbedRequest struct {
	PromptID     uint           `json:"prompt_id"`
	Content      string         `json:"content"`
	VariableKeys []model.VarKey `json:"variable_keys"`
}

// Validity 校验 AddPromptVersionRequest 参数
func (opt *AddPromptVersionRequest) Validity(resp *AddPromptVersionResponse) {
	if opt.Request.PromptID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "prompt_id_is_empty"
		return
	}
	if opt.Request.Content == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "content_is_empty"
		return
	}
	if len(opt.Request.VariableKeys) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "variable_keys_is_empty"
		return
	}
}

// SwitchPromptVersionRequest 切换 prompt 模板生效版本
type SwitchPromptVersionRequest struct {
	apiobj.BaseRequest
	Request SwitchPromptVersionEmbedRequest
}

// SwitchPromptVersionEmbedRequest 切换版本的内嵌请求参数
type SwitchPromptVersionEmbedRequest struct {
	PromptID  uint `json:"prompt_id"`
	VersionID uint `json:"version_id"`
}

// Validity 校验 SwitchPromptVersionRequest 参数
func (opt *SwitchPromptVersionRequest) Validity(resp *SwitchPromptVersionResponse) {
	if opt.Request.PromptID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "prompt_id_is_empty"
		return
	}
	if opt.Request.VersionID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "version_id_is_empty"
		return
	}
}

// GetPromptDetailRequest 获取 prompt 模板详情+当前生效版本
type GetPromptDetailRequest struct {
	apiobj.BaseRequest
	Request GetPromptDetailEmbedRequest
}

// GetPromptDetailEmbedRequest 获取详情的内嵌请求参数
type GetPromptDetailEmbedRequest struct {
	ID uint `json:"id"`
}

// Validity 校验 GetPromptDetailRequest 参数
func (opt *GetPromptDetailRequest) Validity(resp *GetPromptDetailResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "id_is_empty"
		return
	}
}

// ListPromptVersionsRequest 获取 prompt 模板全部版本列表
type ListPromptVersionsRequest struct {
	apiobj.BaseRequest
	Request ListPromptVersionsEmbedRequest
}

// ListPromptVersionsEmbedRequest 版本列表的内嵌请求参数
type ListPromptVersionsEmbedRequest struct {
	PromptID uint `json:"prompt_id"`
	apiobj.PageQuery
}

// Validity 校验 ListPromptVersionsRequest 参数
func (opt *ListPromptVersionsRequest) Validity(resp *ListPromptVersionsResponse) {
	if opt.Request.PromptID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "prompt_id_is_empty"
		return
	}
}

// ListPromptsRequest 查询 prompt 模板列表
type ListPromptsRequest struct {
	apiobj.BaseRequest
	Request ListPromptsEmbedRequest
}

// ListPromptsEmbedRequest 模板列表的内嵌请求参数，支持按 app/group/code/name/status 等筛选
type ListPromptsEmbedRequest struct {
	App      string `json:"app"`
	Group    string `json:"group"`
	Code     string `json:"code"`
	NameLike string `json:"name_like"`
	Status   []int  `json:"status"`
	apiobj.PageQuery
}

// Validity 校验 ListPromptsRequest 参数
func (opt *ListPromptsRequest) Validity(resp *ListPromptsResponse) {
}

// EditPromptRequest 编辑 prompt 模板主记录
type EditPromptRequest struct {
	apiobj.BaseRequest
	Request EditPromptEmbedRequest
}

// EditPromptEmbedRequest 编辑模板的内嵌请求参数
type EditPromptEmbedRequest struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

// Validity 校验 EditPromptRequest 参数
func (opt *EditPromptRequest) Validity(resp *EditPromptResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "id_is_empty"
		return
	}
}

// DeletePromptRequest 删除 prompt 模板（软删）
type DeletePromptRequest struct {
	apiobj.BaseRequest
	Request DeletePromptEmbedRequest
}

// DeletePromptEmbedRequest 删除模板的内嵌请求参数
type DeletePromptEmbedRequest struct {
	ID uint `json:"id"`
}

// Validity 校验 DeletePromptRequest 参数
func (opt *DeletePromptRequest) Validity(resp *DeletePromptResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "id_is_empty"
		return
	}
}

// RenderPromptPreviewRequest 渲染预览：传入 prompt_value，返回渲染后文本
// 支持两种模式：草稿模式（传 content+variable_keys）和已保存模式（传 prompt_id+version_id）
type RenderPromptPreviewRequest struct {
	apiobj.BaseRequest
	Request RenderPromptPreviewEmbedRequest
}

// RenderPromptPreviewEmbedRequest 渲染预览的内嵌请求参数
type RenderPromptPreviewEmbedRequest struct {
	PromptID     uint           `json:"prompt_id"`
	VersionID    uint           `json:"version_id"`
	Content      string         `json:"content"`
	VariableKeys []model.VarKey `json:"variable_keys"`
	PromptValue  map[string]any `json:"prompt_value"`
}

// Validity 校验 RenderPromptPreviewRequest 参数
func (opt *RenderPromptPreviewRequest) Validity(resp *RenderPromptPreviewResponse) {
	if len(opt.Request.PromptValue) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "prompt_value_is_empty"
		return
	}
	// 草稿模式：content 非空时 variable_keys 也必须非空
	if opt.Request.Content != "" && len(opt.Request.VariableKeys) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "variable_keys_is_empty_for_draft_mode"
		return
	}
	// 已保存模式：content 为空时 prompt_id 必须非空
	if opt.Request.Content == "" && opt.Request.PromptID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "prompt_id_is_empty_for_saved_mode"
		return
	}
}
