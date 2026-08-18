package svcprompt

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtoprompt"
	"github.com/insmtx/corekg/apps/admin/models/prompt"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/prompt/model"
	"gorm.io/gorm"
)

// CreatePrompt 创建 prompt 模板主记录+首版本，事务内确保 CreatedUin 一致
func CreatePrompt(ctx *gin.Context, req *dtoprompt.CreatePromptRequest) (res *dtoprompt.CreatePromptResponse, err error) {
	res = &dtoprompt.CreatePromptResponse{}
	uin := runtime.Uin(ctx)

	_, err = model.ParseVariableKeys(ctx, req.Request.Content, req.Request.VariableKeys)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreatePrompt] ParseVariableKeys failed, err: %v", err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = err.Error()
		return res, nil
	}

	varKeysJSON, err := json.Marshal(req.Request.VariableKeys)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreatePrompt] marshal variable_keys failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "marshal_variable_keys_failed"
		return res, err
	}

	promptEntity := &model.CorePrompt{
		CompanyID:  runtime.CompanyID(ctx),
		Uin:        uin,
		App:        req.Request.App,
		Group:      req.Request.Group,
		Name:       req.Request.Name,
		Code:       req.Request.Code,
		Status:     model.PromptStatusEnabled,
		CreatedUin: uin,
		UpdatedUin: uin,
	}

	txErr := prompt.CoreDB().Transaction(func(tx *gorm.DB) error {
		promptDao := model.NewPromptDao(tx)
		if err := promptDao.Create(ctx, promptEntity); err != nil {
			return err
		}

		versionEntity := &model.CorePromptVersion{
			CompanyID:    runtime.CompanyID(ctx),
			Uin:          uin,
			PromptID:     promptEntity.ID,
			Content:      req.Request.Content,
			VariableKeys: json.RawMessage(varKeysJSON),
			CreatedUin:   uin,
			UpdatedUin:   uin,
		}
		versionDao := model.NewPromptVersionDao(tx)
		if err := versionDao.Create(ctx, versionEntity); err != nil {
			return err
		}

		if err := promptDao.UpdateMap(ctx, promptEntity.ID, map[string]interface{}{
			"latest_version_id": versionEntity.ID,
			"updated_uin":       uin,
		}); err != nil {
			return err
		}

		promptEntity.LatestVersionID = versionEntity.ID
		return nil
	})

	if txErr != nil {
		logs.ErrorContextf(ctx, "[CreatePrompt] transaction failed, err: %v", txErr)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "create_prompt_transaction_failed"
		return res, txErr
	}

	res.Response.PromptID = promptEntity.ID
	return res, nil
}

// AddPromptVersion 新增 prompt 模板版本，含模板校验，自动回写主表 latest_version_id
func AddPromptVersion(ctx *gin.Context, req *dtoprompt.AddPromptVersionRequest) (res *dtoprompt.AddPromptVersionResponse, err error) {
	res = &dtoprompt.AddPromptVersionResponse{}
	uin := runtime.Uin(ctx)

	_, err = model.ParseVariableKeys(ctx, req.Request.Content, req.Request.VariableKeys)
	if err != nil {
		logs.ErrorContextf(ctx, "[AddPromptVersion] ParseVariableKeys failed, err: %v", err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = err.Error()
		return res, nil
	}

	p, err := prompt.NewPromptDao().GetByID(ctx, req.Request.PromptID)
	if err != nil {
		logs.ErrorContextf(ctx, "[AddPromptVersion] GetByID failed, prompt_id: %d, err: %v", req.Request.PromptID, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "prompt_get_failed"
		return res, err
	}
	if p.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "prompt_not_found"
		return res, nil
	}

	varKeysJSON, err := json.Marshal(req.Request.VariableKeys)
	if err != nil {
		logs.ErrorContextf(ctx, "[AddPromptVersion] marshal variable_keys failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "marshal_variable_keys_failed"
		return res, err
	}

	versionEntity := &model.CorePromptVersion{
		CompanyID:    p.CompanyID,
		Uin:          uin,
		PromptID:     p.ID,
		Content:      req.Request.Content,
		VariableKeys: json.RawMessage(varKeysJSON),
		CreatedUin:   uin,
		UpdatedUin:   uin,
	}

	if err := prompt.NewPromptVersionDao().Create(ctx, versionEntity); err != nil {
		logs.ErrorContextf(ctx, "[AddPromptVersion] Insert version failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "insert_version_failed"
		return res, err
	}

	if err := prompt.NewPromptDao().UpdateMap(ctx, p.ID, map[string]interface{}{
		"latest_version_id": versionEntity.ID,
		"updated_uin":       uin,
	}); err != nil {
		logs.ErrorContextf(ctx, "[AddPromptVersion] UpdateMap prompt failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "update_prompt_failed"
		return res, err
	}

	res.Response.VersionID = versionEntity.ID
	return res, nil
}

// SwitchPromptVersion 切换 prompt 模板生效版本
func SwitchPromptVersion(ctx *gin.Context, req *dtoprompt.SwitchPromptVersionRequest) (res *dtoprompt.SwitchPromptVersionResponse, err error) {
	res = &dtoprompt.SwitchPromptVersionResponse{}
	uin := runtime.Uin(ctx)

	version, err := prompt.NewPromptVersionDao().GetByID(ctx, req.Request.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "[SwitchPromptVersion] GetByID version failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "version_get_failed"
		return res, err
	}
	if version.ID == 0 || version.PromptID != req.Request.PromptID {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "version_not_found_or_not_belong_to_prompt"
		return res, nil
	}

	if err := prompt.NewPromptDao().UpdateMap(ctx, req.Request.PromptID, map[string]interface{}{
		"latest_version_id": req.Request.VersionID,
		"updated_uin":       uin,
	}); err != nil {
		logs.ErrorContextf(ctx, "[SwitchPromptVersion] UpdateMap failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "switch_version_failed"
		return res, err
	}

	return res, nil
}

// GetPromptDetail 获取 prompt 模板详情+当前生效版本内容+variable_keys
func GetPromptDetail(ctx *gin.Context, req *dtoprompt.GetPromptDetailRequest) (res *dtoprompt.GetPromptDetailResponse, err error) {
	res = &dtoprompt.GetPromptDetailResponse{}

	p, err := prompt.NewPromptDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetPromptDetail] GetByID failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "prompt_get_failed"
		return res, err
	}
	if p.ID == 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "prompt_not_found"
		return res, nil
	}

	var version *model.CorePromptVersion
	if p.LatestVersionID > 0 {
		version, err = prompt.NewPromptVersionDao().GetByID(ctx, p.LatestVersionID)
		if err != nil {
			logs.ErrorContextf(ctx, "[GetPromptDetail] GetByID version failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "version_get_failed"
			return res, err
		}
	}

	if version == nil {
		version = &model.CorePromptVersion{}
	}

	res.Response.Prompt = *p
	res.Response.Version = *version
	return res, nil
}

// ListPromptVersions 获取 prompt 模板全部版本列表（分页）
func ListPromptVersions(ctx *gin.Context, req *dtoprompt.ListPromptVersionsRequest) (res *dtoprompt.ListPromptVersionsResponse, err error) {
	res = &dtoprompt.ListPromptVersionsResponse{}

	cond := &model.PromptVersionCond{
		PromptID: req.Request.PromptID,
		OrderBy:  req.Request.OrderBy,
		Offset:   req.Request.Offset,
		Limit:    req.Request.Limit,
	}

	versions, count, err := prompt.NewPromptVersionDao().GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListPromptVersions] GetPageListByCond failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "list_versions_failed"
		return res, err
	}

	res.Response.Data = versions
	res.Response.Total = count
	res.Response.Limit = req.Request.Limit
	res.Response.Offset = req.Request.Offset
	return res, nil
}

// ListPrompts 查询 prompt 模板列表（按 app/group/code/name/status 等筛选，分页）
func ListPrompts(ctx *gin.Context, req *dtoprompt.ListPromptsRequest) (res *dtoprompt.ListPromptsResponse, err error) {
	res = &dtoprompt.ListPromptsResponse{}

	var statuses []model.PromptStatus
	if len(req.Request.Status) > 0 {
		statuses = make([]model.PromptStatus, len(req.Request.Status))
		for i, s := range req.Request.Status {
			statuses[i] = model.PromptStatus(s)
		}
	}

	cond := &model.PromptCond{
		Group:    req.Request.Group,
		Code:     req.Request.Code,
		NameLike: req.Request.NameLike,
		Status:   statuses,
		OrderBy:  req.Request.OrderBy,
		Offset:   req.Request.Offset,
		Limit:    req.Request.Limit,
	}

	prompts, count, err := prompt.NewPromptDao().GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListPrompts] GetPageListByCond failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "list_prompts_failed"
		return res, err
	}

	res.Response.Data = prompts
	res.Response.Total = count
	res.Response.Limit = req.Request.Limit
	res.Response.Offset = req.Request.Offset
	return res, nil
}

// EditPrompt 编辑 prompt 模板主记录（name/status 等）
func EditPrompt(ctx *gin.Context, req *dtoprompt.EditPromptRequest) (res *dtoprompt.EditPromptResponse, err error) {
	res = &dtoprompt.EditPromptResponse{}
	uin := runtime.Uin(ctx)

	updateMap := map[string]interface{}{
		"updated_uin": uin,
	}
	if req.Request.Name != "" {
		updateMap["name"] = req.Request.Name
	}
	if req.Request.Status >= 0 {
		updateMap["status"] = req.Request.Status
	}

	if err := prompt.NewPromptDao().UpdateMap(ctx, req.Request.ID, updateMap); err != nil {
		logs.ErrorContextf(ctx, "[EditPrompt] UpdateMap failed, err: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "edit_prompt_failed"
		return res, err
	}

	return res, nil
}

// DeletePrompt 删除 prompt 模板（软删）
func DeletePrompt(ctx *gin.Context, req *dtoprompt.DeletePromptRequest) (res *dtoprompt.DeletePromptResponse, err error) {
	res = &dtoprompt.DeletePromptResponse{}

	if err := prompt.NewPromptDao().Delete(ctx, req.Request.ID); err != nil {
		logs.ErrorContextf(ctx, "[DeletePrompt] Delete failed, id: %d, err: %v", req.Request.ID, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "delete_prompt_failed"
		return res, err
	}

	return res, nil
}

// RenderPromptPreview 渲染预览：支持草稿模式（传 content+variable_keys）和已保存模式（传 prompt_id+version_id）
func RenderPromptPreview(ctx *gin.Context, req *dtoprompt.RenderPromptPreviewRequest) (res *dtoprompt.RenderPromptPreviewResponse, err error) {
	res = &dtoprompt.RenderPromptPreviewResponse{}

	var content string
	var variableKeys []model.VarKey

	if req.Request.Content != "" {
		// 草稿模式：直接使用传入的 content 和 variable_keys
		content = req.Request.Content
		variableKeys = req.Request.VariableKeys
	} else {
		// 已保存模式：从 DB 查版本
		p, err := prompt.NewPromptDao().GetByID(ctx, req.Request.PromptID)
		if err != nil {
			logs.ErrorContextf(ctx, "[RenderPromptPreview] GetByID prompt failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "prompt_get_failed"
			return res, err
		}
		if p.ID == 0 {
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "prompt_not_found"
			return res, nil
		}

		versionID := req.Request.VersionID
		if versionID == 0 {
			versionID = p.LatestVersionID
		}

		version, err := prompt.NewPromptVersionDao().GetByID(ctx, versionID)
		if err != nil {
			logs.ErrorContextf(ctx, "[RenderPromptPreview] GetByID version failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "version_get_failed"
			return res, err
		}
		if version.ID == 0 {
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "version_not_found"
			return res, nil
		}

		content = version.Content
		if err := json.Unmarshal(version.VariableKeys, &variableKeys); err != nil {
			logs.ErrorContextf(ctx, "[RenderPromptPreview] unmarshal variable_keys failed, err: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "unmarshal_variable_keys_failed"
			return res, err
		}
	}

	rendered, err := model.ValidateAndRender(ctx, content, variableKeys, req.Request.PromptValue)
	if err != nil {
		logs.ErrorContextf(ctx, "[RenderPromptPreview] ValidateAndRender failed, err: %v", err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = err.Error()
		return res, nil
	}

	res.Response.RenderedText = rendered
	return res, nil
}
