package fileqactl

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListFileQARequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		FileID   uint `json:"file_id"`
	}
}

func (opt *ListFileQARequest) Validity(resp *ListFileQAResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_file" // 请选择文件
		return
	}
}

type ListFileQAResponse struct {
	apiobj.BaseResponse
	Response keqa.QueryForestQAListResponse
}

type DeleteFileQARequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		FileID   uint `json:"file_id"`
	}
}

func (opt *DeleteFileQARequest) Validity(resp *DeleteFileQAResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_file" // 请选择文件
		return
	}
}

type DeleteFileQAResponse struct {
	apiobj.BaseResponse
}

type FileChatRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		FileID   uint   `json:"file_id"`
		Question string `json:"question"`
	}
}

func (opt *FileChatRequest) Validity(resp *FileChatResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_file" // 请选择文件
		return
	}
	if opt.Request.Question == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_content" // 请输入内容
		return
	}
}

type FileChatResponse struct {
	apiobj.BaseResponse
	Response struct {
		Answer string              `json:"answer"`
		Status foresttype.QAStatus `json:"status"`
	}
}
