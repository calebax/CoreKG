package dtomodel

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

const bindCozeModelToken = "bind_coze_model"

// BindCozeModelRequest 绑定 coze model 的请求
type BindCozeModelRequest struct {
	apiobj.BaseRequest
	Request struct {
		ModelID     uint   `json:"model_id"`
		CozeModelID uint   `json:"coze_model_id"`
		Token       string `json:"token"`
	}
}

func (opt *BindCozeModelRequest) Validity(resp *BindCozeModelResponse) {
	if opt.Request.ModelID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_model_id_required"
		return
	}
	if opt.Request.Token == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_token"
		return
	}
	if opt.Request.Token != bindCozeModelToken {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_token"
		return
	}
}
