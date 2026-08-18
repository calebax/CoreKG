package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ListChatModelRequest 模型列表请求
type ListModelRequest apiobj.QueryRequest

func (req *ListModelRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_offset_limit" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_invalid_order_by_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "created_at", "updated_at", "show_name":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_field_value_required" // 查询条件中的值不能为空
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_filter_field_not_exist" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// 模型列表响应
type ListModelResponse struct {
	apiobj.BaseResponse
	Response chatmodel.QueryModelListResponse
}

// CreateModelRequest 创建用户自建模型
type CreateModelRequest struct {
	apiobj.BaseRequest
	Request struct {
		ShowName      string `json:"show_name"`      // 模型显示名称
		ModelName     string `json:"model_name"`     // 模型名称
		APIKey        string `json:"api_key"`        // API密钥
		ModelUrl      string `json:"model_url"`      // 模型URL
		ModelProvider string `json:"model_provider"` // 模型供应商
	}
}

func (req *CreateModelRequest) Validity(resp *apiobj.BaseResponse) {
	if len(req.Request.ShowName) > 32 || len(req.Request.ShowName) < 2 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_name_required_length" // 模型名称合法长度为2～32
		return
	}

	if !chatmodel.ValidModel(req.Request.ShowName) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_name" // 非法模型名称
		return
	}

	if req.Request.ModelUrl == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_url_required" // 模型URL不能为空
		return
	}

	if len(req.Request.ModelName) > 32 || len(req.Request.ModelName) < 2 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_name_length" // 基础模型合法长度为2~32
		return
	}

	if !chatmodel.ValidModel(req.Request.ModelName) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_model_name" // 非法基础模型
		return
	}

	if req.Request.ModelProvider == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_provider_required" // 模型供应商不能为空
		return
	}
}

// CreateModelResponse 创建用户自建模型
type CreateModelResponse struct {
	apiobj.BaseResponse
}

// DeleteModelRequest 删除用户自建模型
type DeleteModelRequest apiobj.DetailIdRequest

func (req *DeleteModelRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_id_required" // 模型ID不能为空
		return
	}
	if version.DeployMode() != "" && req.Request.ID == 1 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_default_model_delete_forbidden" // 默认模型无法删除
	}
}

// DeleteModelResponse DeleteModelRequest
type DeleteModelResponse struct {
	apiobj.BaseResponse
}

// UpdateModelRequest 更新用户自建模型
type UpdateModelRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID            uint   `json:"id"`             // 模型ID
		ShowName      string `json:"show_name"`      // 模型显示名称
		ModelName     string `json:"model_name"`     // 模型名称
		APIKey        string `json:"api_key"`        // API密钥
		ModelUrl      string `json:"model_url"`      // 模型URL
		ModelProvider string `json:"model_provider"` // 模型供应商
	}
}

func (req *UpdateModelRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_id_required" // 模型ID不能为空
		return
	}
	if req.Request.ShowName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_name_required" // 模型显示名称不能为空
		return
	}
	if req.Request.ModelName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_required" // 模型不能为空
		return
	}
	if !chatmodel.ValidModel(req.Request.ModelName) {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_model_name" // 非法基础模型
		return
	}

	if req.Request.ModelUrl == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_url_required" // 模型URL不能为空
		return
	}
	if req.Request.ModelProvider == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_provider_required" // 模型供应商不能为空
		return
	}
}

// UpdateModelResponse
type UpdateModelResponse struct {
	apiobj.BaseResponse
}

// GetModelDetailRequest 获取模型详情请求
type GetModelDetailRequest apiobj.DetailIdRequest

func (req *GetModelDetailRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_id_required" // 模型ID不能为空
		return
	}
}

// GetModelDetailResponse
type GetModelDetailResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data *chattype.ChatModel `json:"data"`
	}
}

// ModelTestRequest 模型测试请求
type ModelTestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ModelName string `json:"model_name"` // 模型名称
		APIKey    string `json:"api_key"`    // API密钥
		ModelUrl  string `json:"model_url"`  // 模型URL
	}
}

func (req *ModelTestRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ModelName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_required" // 模型不能为空
		return
	}
	if req.Request.ModelUrl == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_url_required" // 模型URL不能为空
		return
	}
}

// ModelTestResponse 模型测试响应
type ModelTestResponse struct {
	apiobj.BaseResponse
	Response struct {
		Pass bool `json:"pass"` // 测试是否通过
	}
}
