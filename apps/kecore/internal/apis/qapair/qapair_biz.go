package qapair

import (
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// CreateQAPairRequest 创建问答对请求
type CreateQAPairRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID    uint     `json:"forest_id"`
		Question    string   `json:"question"`
		SubQuestion []string `json:"sub_question"`
		Answer      string   `json:"answer"`
		Label       []string `json:"label"`
	}
}

// Validity 校验创建问答对请求
func (req *CreateQAPairRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id"
		return
	}
	if len(req.Request.Question) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_question_empty"
		return
	}
	if len(req.Request.Answer) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_answer_empty"
		return
	}
}

// DeleteQAPairRequest 删除问答对请求
type DeleteQAPairRequest struct {
	apiobj.BaseRequest
	Request struct {
		QuestionIDs []string `json:"question_ids"`
		ForestID    uint     `json:"forest_id"`
	}
}

// Validity 校验删除问答对请求
func (req *DeleteQAPairRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id"
		return
	}
	if len(req.Request.QuestionIDs) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_question_ids_empty"
		return
	}
}

// ListQAPairRequest 列表问答对请求
type ListQAPairRequest struct {
	apiobj.BaseRequest
	Request struct {
		apiobj.PageQuery
		ForestID uint `json:"forest_id"`
	}
}

// ListQAPairResponse 列表问答对响应
type ListQAPairResponse struct {
	apiobj.BaseResponse
	Response essearch.QueryFQAResponse
}

// Validity 校验列表问答对请求
func (req *ListQAPairRequest) Validity(resp *ListQAPairResponse) {
	if req.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id"
		return
	}
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_offset_limit"
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "updated_at", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_order_by_empty"
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "company_id", "qa_lable", "qa_main_id", "qa_question", "qa_answer", "uin":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_single_value"
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty_value"
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field_data"
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ModifyQAPairRequest 修改问答对请求
type ModifyQAPairRequest struct {
	apiobj.BaseRequest
	Request struct {
		*essearch.FQAItem
	}
}

// Validity 校验修改问答对请求
func (req *ModifyQAPairRequest) Validity(resp *apiobj.BaseResponse) {
	if err := uuid.Validate(req.Request.Main.ID); err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_question_id"
		return
	}
	for _, v := range req.Request.Child {
		if len(v.ID) > 0 {
			if err := uuid.Validate(v.ID); err != nil {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_invalid_sub_question_id"
				return
			}
		}
	}
}

// UploadQAPairResponse 上传问答对响应
type UploadQAPairResponse struct {
	apiobj.BaseResponse
	Response struct {
		// 总行数
		TotalLines uint `json:"total_lines"`
		// 有效行数
		ValidLines uint `json:"valid_lines"`
		// 问答对列表
		QAList []*keqa.PureQAItem `json:"qa_list"`
	}
}

// CommitQAPairRequest 提交问答对请求
type CommitQAPairRequest struct {
	apiobj.BaseRequest
	Request struct {
		// 知识库ID
		ForestID uint `json:"forest_id"`
		// 问答对列表
		QAList []*keqa.PureQAItem `json:"qa_list"`
	}
}

// Validity 校验提交问答对请求
func (r *CommitQAPairRequest) Validity(resp *apiobj.BaseResponse) {
	if r.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_forest_id"
	}
}
