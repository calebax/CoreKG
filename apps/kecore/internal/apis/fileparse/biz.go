package fileparse

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type GetContentRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

type GetContentResponse struct {
	apiobj.BaseResponse
	Response struct {
		Content string                             `json:"content"`
		Status  foresttype.KnownowForestTaskStatus `json:"status"`
	}
}

func (req *GetContentRequest) Validity(response *GetContentResponse) {

}

type GetMindMapRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

type GetMindMapResponse struct {
	apiobj.BaseResponse
	Response struct {
		MindMap string                             `json:"mind_map"`
		Status  foresttype.KnownowForestTaskStatus `json:"status"`
	}
}

func (req *GetMindMapRequest) Validity(response *GetMindMapResponse) {

}

type GetAnalysisRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

type GetAnalysisResponse struct {
	apiobj.BaseResponse
	Response struct {
		Analysis string                             `json:"analysis"`
		Status   foresttype.KnownowForestTaskStatus `json:"status"`
	}
}

func (req *GetAnalysisRequest) Validity(response *GetAnalysisResponse) {
	if req.Request.FileID <= 0 {
		response.Code = errcode.ErrCode_BadRequest
		response.Message = "kecore_invalid_file_id" // 非法文件id
	}
}

type GetFileDescriptionRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

type GetFileDescriptionResponse struct {
	apiobj.BaseResponse
	Response struct {
		Abstract    string `json:"abstract"`
		MindMap     string `json:"mind_map"`
		Description string `json:"description"`
	}
}

func (req *GetFileDescriptionRequest) Validity(resp *GetFileDescriptionResponse) {
	if req.Request.FileID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_file_id" // 非法文件id
	}
}

type GetRecommendQuestionsRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

type GetRecommendQuestionsResponse struct {
	apiobj.BaseResponse
	Response struct {
		RecommendQuestions []string                           `json:"recommend_questions"`
		Status             foresttype.KnownowForestTaskStatus `json:"status"`
	}
}

func (req *GetRecommendQuestionsRequest) Validity(response *GetRecommendQuestionsResponse) {
	if req.Request.FileID <= 0 {
		response.Code = errcode.ErrCode_BadRequest
		response.Message = "kecore_invalid_file_id" // 非法文件id
	}
}
