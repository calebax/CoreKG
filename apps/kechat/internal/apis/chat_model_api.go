package apis

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtomodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/services/svcmodel"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ListModel 模型列表
// @Tags 模型管理
// @Summary 模型列表
// @Description 模型列表
// @Router /chat.ListModel [post]
// @Param request body ListModelRequest true "入参"
// @Success 200 {object} ListModelResponse
func ListModel(ctx *gin.Context, req *ListModelRequest, resp *ListModelResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	req.Request.Uin = runtime.Uin(ctx)
	err := chatmodel.QueryModelList(ctx, req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_list_failed" // 获取模型列表失败
		logs.ErrorContextf(ctx, "ListSettingGroup failed ,err %s", err)
		return
	}
}

// CreateModel 创建模型
// @Tags 模型管理
// @Summary 创建模型
// @Description 创建模型
// @Router /chat.CreateModel [post]
// @Param request body CreateModelRequest true "入参"
// @Success 200 {object} CreateModelResponse
func CreateModel(ctx *gin.Context, req *CreateModelRequest, resp *CreateModelResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	model := &chattype.ChatModel{
		Uin:           runtime.Uin(ctx),
		CompanyID:     runtime.CompanyID(ctx),
		ShowName:      req.Request.ShowName,
		ModelName:     req.Request.ModelName,
		ModelUrl:      req.Request.ModelUrl,
		APIKey:        req.Request.APIKey,
		ModelProvider: req.Request.ModelProvider,
	}
	model.HeadURL = chatmodel.GetProviderHeadURL(model.ModelProvider)

	err := chatmodel.CreateModel(ctx, model)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_model_failed" // 创建模型失败
		logs.ErrorContextf(ctx, "CreateModel failed ,err %s", err)
		return
	}

	if _, err := svcmodel.SyncCozeModelInstance(ctx, model.ID, model.CozeModelID); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_model_failed" // 创建模型失败
		logs.ErrorContextf(ctx, "CreateModel sync coze model failed ,err %s", err)
		return
	}
}

// DeleteModel 删除模型
// @Tags 模型管理
// @Summary 删除模型
// @Description 删除模型
// @Router /chat.DeleteModel [post]
// @Param request body DeleteModelRequest true "入参"
// @Success 200 {object} DeleteModelResponse
func DeleteModel(ctx *gin.Context, req *DeleteModelRequest, resp *DeleteModelResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	err := svcmodel.DeleteModel(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_delete_model_failed" // 删除模型失败
		logs.ErrorContextf(ctx, "DeleteChatModel failed ,err %s", err)
		return
	}
}

// UpdateModel 修改模型
// @Tags 模型管理
// @Summary 修改模型
// @Description 修改模型
// @Router /chat.UpdateModel [post]
// @Param request body UpdateModelRequest true "入参"
// @Success 200 {object} UpdateModelResponse
func UpdateModel(ctx *gin.Context, req *UpdateModelRequest, resp *UpdateModelResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	model, err := chatmodel.GetModelByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed" // 查询模型失败
		logs.ErrorContextf(ctx, "UpdateChatModel GetModelByID failed ,err %s", err)
		return
	}
	model.ShowName = req.Request.ShowName
	model.ModelName = req.Request.ModelName
	model.APIKey = req.Request.APIKey
	model.ModelUrl = req.Request.ModelUrl
	model.ModelProvider = req.Request.ModelProvider
	err = chatmodel.UpdateModel(ctx, model)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_model_failed" // 更新模型失败
		logs.ErrorContextf(ctx, "UpdateChatModel failed ,err %s", err)
		return
	}

	if _, err := svcmodel.SyncCozeModelInstance(ctx, model.ID, model.CozeModelID); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_model_failed" // 更新模型失败
		logs.ErrorContextf(ctx, "UpdateChatModel sync coze model failed ,err %s", err)
		return
	}
}

// BindCozeModel 绑定或重新生成 coze model
// @Tags 模型管理
// @Summary 绑定 coze 模型
// @Description 绑定 coze 模型
// @Router /chat.BindCozeModel [post]
// @Param request body dtomodel.BindCozeModelRequest true "入参"
// @Success 200 {object} dtomodel.BindCozeModelResponse
func BindCozeModel(ctx *gin.Context, req *dtomodel.BindCozeModelRequest, resp *dtomodel.BindCozeModelResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}

	newCozeID, err := svcmodel.BindCozeModel(ctx, req.Request.ModelID, req.Request.CozeModelID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_model_failed"
		logs.ErrorContextf(ctx, "BindCozeModel failed ,err %s", err)
		return
	}
	resp.Response.CozeModelID = newCozeID
}

// GetModelDetail 获取模型详情
// @Tags 模型管理
// @Summary 获取模型详情
// @Description 获取模型详情
// @Router /chat.GetModelDetail [post]
// @Param request body GetModelDetailRequest true "入参"
// @Success 200 {object} GetModelDetailResponse
func GetModelDetail(ctx *gin.Context, req *GetModelDetailRequest, resp *GetModelDetailResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	model, err := chatmodel.GetModelByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed" // 查询模型失败
		logs.ErrorContextf(ctx, "UpdateChatModel GetModelByID failed ,err %s", err)
		return
	}
	resp.Response.Data = model
}

// ModelTest 模型测试
// @Tags 模型管理
// @Summary 模型测试
// @Description 模型测试
// @Router /chat.ModelTest [post]
// @Param request body ModelTestRequest true "入参"
// @Success 200 {object} ModelTestResponse
func ModelTest(ctx *gin.Context, req *ModelTestRequest, resp *ModelTestResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.ErrorContextf(ctx, "kechat_validate_params_failed") // 参数校验失败
		return
	}
	body := map[string]interface{}{
		"messages": []map[string]interface{}{
			{"role": "user", "content": "hello"},
		},
		"model":      req.Request.ModelName,
		"max_tokens": 2,
		"response_format": map[string]interface{}{
			"type": "text",
		},
		"stream": false,
	}
	// 将请求体转换为 JSON
	jsonPayload, err := json.Marshal(body)
	if err != nil {
		resp.Message = "kechat_internal_error" // 内部错误
		logs.WarnContextf(ctx, "[chat] [ModelTest] Failed to convert request body to JSON: %s", err.Error())
		return
	}
	// 创建 HTTP
	request, err := http.NewRequest("POST", req.Request.ModelUrl, strings.NewReader(string(jsonPayload)))
	if err != nil {
		resp.Message = "kechat_model_request_failed" // 模型请求失败
		logs.WarnContextf(ctx, "[chat] [ModelTest] Failed to create HTTP request: %s", err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if req.Request.APIKey != "" {
		request.Header.Set("Authorization", "Bearer "+req.Request.APIKey)
	}
	// 创建 HTTP 客户端
	client := &http.Client{}

	// 发送请求
	response, err := client.Do(request)
	if err != nil {
		resp.Message = "kechat_send_model_request_failed" // 发送模型请求失败
		logs.WarnContextf(ctx, "[aigc] [ModelTest] Failed to request DeepSeek: %s", err.Error())
		return
	}

	// 检查响应状态码
	if response.StatusCode != http.StatusOK {
		resp.Message = "kechat_check_model_status_failed" // 检查模型状态失败
		logs.WarnContextf(ctx, "[chat] [ModelTest] Response status code %v, token %s is invalid", response.StatusCode, req.Request.APIKey)
		return
	}
	resp.Response.Pass = true
}
