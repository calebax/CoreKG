package dtoprompt

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/prompt/model"
)

// CreatePromptResponse 创建 prompt 模板的响应
type CreatePromptResponse struct {
	apiobj.BaseResponse
	Response CreatePromptEmbedResponse
}

// CreatePromptEmbedResponse 创建 prompt 的内嵌响应参数
type CreatePromptEmbedResponse struct {
	PromptID uint `json:"prompt_id"`
}

// AddPromptVersionResponse 新增版本的响应
type AddPromptVersionResponse struct {
	apiobj.BaseResponse
	Response AddPromptVersionEmbedResponse
}

// AddPromptVersionEmbedResponse 新增版本的内嵌响应参数
type AddPromptVersionEmbedResponse struct {
	VersionID uint `json:"version_id"`
}

// SwitchPromptVersionResponse 切换版本的响应
type SwitchPromptVersionResponse struct {
	apiobj.BaseResponse
	Response SwitchPromptVersionEmbedResponse
}

// SwitchPromptVersionEmbedResponse 切换版本的内嵌响应参数
type SwitchPromptVersionEmbedResponse struct{}

// GetPromptDetailResponse 获取详情的响应
type GetPromptDetailResponse struct {
	apiobj.BaseResponse
	Response GetPromptDetailEmbedResponse
}

// GetPromptDetailEmbedResponse 获取详情的内嵌响应参数
type GetPromptDetailEmbedResponse struct {
	Prompt  model.CorePrompt        `json:"prompt"`
	Version model.CorePromptVersion `json:"version"`
}

// ListPromptVersionsResponse 版本列表的响应
type ListPromptVersionsResponse struct {
	apiobj.BaseResponse
	Response ListPromptVersionsEmbedResponse
}

// ListPromptVersionsEmbedResponse 版本列表的内嵌响应参数
type ListPromptVersionsEmbedResponse struct {
	apiobj.QueryResponse
	Data model.CorePromptVersionList `json:"data"`
}

// ListPromptsResponse 模板列表的响应
type ListPromptsResponse struct {
	apiobj.BaseResponse
	Response ListPromptsEmbedResponse
}

// ListPromptsEmbedResponse 模板列表的内嵌响应参数
type ListPromptsEmbedResponse struct {
	apiobj.QueryResponse
	Data model.CorePromptList `json:"data"`
}

// EditPromptResponse 编辑模板的响应
type EditPromptResponse struct {
	apiobj.BaseResponse
	Response EditPromptEmbedResponse
}

// EditPromptEmbedResponse 编辑模板的内嵌响应参数
type EditPromptEmbedResponse struct{}

// DeletePromptResponse 删除模板的响应
type DeletePromptResponse struct {
	apiobj.BaseResponse
	Response DeletePromptEmbedResponse
}

// DeletePromptEmbedResponse 删除模板的内嵌响应参数
type DeletePromptEmbedResponse struct{}

// RenderPromptPreviewResponse 渲染预览的响应
type RenderPromptPreviewResponse struct {
	apiobj.BaseResponse
	Response RenderPromptPreviewEmbedResponse
}

// RenderPromptPreviewEmbedResponse 渲染预览的内嵌响应参数
type RenderPromptPreviewEmbedResponse struct {
	RenderedText string `json:"rendered_text"`
}
