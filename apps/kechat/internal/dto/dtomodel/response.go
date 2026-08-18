package dtomodel

import "github.com/ygpkg/yg-go/apis/apiobj"

// BindCozeModelResponse 绑定 coze model 的响应
type BindCozeModelResponse struct {
	apiobj.BaseResponse
	Response struct {
		CozeModelID uint `json:"coze_model_id"`
	}
}
