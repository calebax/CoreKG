package apis

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// ListFileQARequest 请求结构体
type ListFileQARequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		FileID   uint `json:"file_id"`
	}
}

// Validity 校验请求参数
func (opt *ListFileQARequest) Validity(resp *ListFileQAResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_file" // 请选择文件
		return
	}
}

// ListFileQAResponse 响应结构体
type ListFileQAResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data []*chattype.ChatQuestion
	}
}

// FileChatRequest 请求结构体
type FileChatRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		FileID   uint   `json:"file_id"`
		Question string `json:"question"`
	}
}

// Validity 校验请求参数
func (opt *FileChatRequest) Validity(resp *FileChatResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_file" // 请选择文件
		return
	}
	if opt.Request.Question == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_enter_content" // 请输入内容
		return
	}
}

// FileChatResponse 响应结构体
type FileChatResponse struct {
	apiobj.BaseResponse
	Response struct {
		Answer string                  `json:"answer"`
		Status chattype.QuestionStatus `json:"status"`
	}
}

// DeleteFileQARequest 请求结构体
type DeleteFileQARequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		FileID   uint `json:"file_id"`
	}
}

// Validity 校验请求参数
func (opt *DeleteFileQARequest) Validity(resp *DeleteFileQAResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_file" // 请选择文件
		return
	}
}

// DeleteFileQAResponse 响应结构体
type DeleteFileQAResponse struct {
	apiobj.BaseResponse
}
