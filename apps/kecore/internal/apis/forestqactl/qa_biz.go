package forestqactl

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/keqa"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListSessionRequest apiobj.QueryRequest

func (req *ListSessionRequest) Validity(resp *ListSessionResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_orderby_field" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "created_at", "updated_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_single_value" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty_value" // 查询条件中的值不能为空
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_invalid_filter_field_data" // 查询条件中的字段不存在, {{.field}}
			return
		}
	}
}

type ListSessionResponse struct {
	apiobj.BaseResponse
	Response keqa.QueryQASessionListResponse
}

type ListSessionQARequest struct {
	apiobj.BaseRequest
	Request struct {
		SessionID uint `json:"session_id"`
	}
}

func (opt *ListSessionQARequest) Validity(resp *ListSessionQAResponse) {
	if opt.Request.SessionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
}

type ListSessionQAResponse struct {
	apiobj.BaseResponse
	Response keqa.QueryForestQAListResponse
}

type CreateForestSessionRequest struct {
	apiobj.BaseRequest
	Request struct {
		// SessionType 会话类型，excel_list：excel 文件列表，excel_sheet_list：excel sheet 列表, db_list：数据库列表，table_list：数据库表列表
		SessionType foresttype.KnownowQASessionType `json:"session_type"`
		IDS         types.UintArray                 `json:"ids"`
		LLMModelID  uint                            `json:"lm_model_id"`
	}
}

func (opt *CreateForestSessionRequest) Validity(resp *CreateForestSessionResponse) {
	if opt.Request.IDS.Len() == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_file" // 请选择知识森林文件
		return
	}
	if opt.Request.SessionType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_session_type" // 请输入会话类型
		return
	}
	if opt.Request.LLMModelID < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_model_id" // 非法模型id
		return
	}
}

type CreateForestSessionResponse struct {
	apiobj.BaseResponse
	Response *foresttype.KnownowQASession
}

type ModifySessionRequest struct {
	apiobj.BaseRequest
	Request struct {
		SessionID uint   `json:"session_id"`
		Name      string `json:"name"`
	}
}

func (opt *ModifySessionRequest) Validity(resp *ModifySessionResponse) {
	if opt.Request.SessionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_session_name" // 请输入会话名称
		return
	}
}

type ModifySessionResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type DeleteSessionRequest struct {
	apiobj.BaseRequest
	Request struct {
		SessionID uint `json:"session_id"`
	}
}

func (opt *DeleteSessionRequest) Validity(resp *DeleteSessionResponse) {
	if opt.Request.SessionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
}

type DeleteSessionResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type CreateForestQARequest struct {
	apiobj.BaseRequest
	Request struct {
		SessionID    uint     `json:"session_id"`
		Question     string   `json:"question"`
		ImageUrlList []string `json:"image_url_list"`
	}
}

func (opt *CreateForestQARequest) Validity(resp *CreateForestQAResponse) {
	if opt.Request.SessionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
	if opt.Request.Question == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_question" // 请输入问题
		return
	}
}

type CreateForestQAResponse struct {
	apiobj.BaseResponse
	Response *foresttype.KnownowForestQA
}

type ForestQAChatRequest struct {
	apiobj.BaseRequest
	Request struct {
		SessionID  uint `json:"session_id"`
		QuestionID uint `json:"question_id"`
	}
}

func (opt *ForestQAChatRequest) Validity(resp *ForestQAChatResponse) {
	if opt.Request.SessionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
	if opt.Request.QuestionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_question" // 请输入问题
		return
	}
}

type ForestQAChatResponse struct {
	apiobj.BaseResponse
	Response struct {
		Answer string              `json:"answer"`
		Status foresttype.QAStatus `json:"status"`
	}
}

type ChatGetMessageRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

func (opt *ChatGetMessageRequest) Validity(resp *ChatGetMessageResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
	// if opt.Request.QuestionID == 0 {
	// 	resp.Code = errcode.ErrCode_BadRequest
	// 	resp.Message = "请输入问题"
	// 	return
	// }
}

type ChatGetMessageResponse struct {
	apiobj.BaseResponse
}

type GetSessionInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

type GetSessionInfoResponse struct {
	apiobj.BaseResponse
	Response *keqa.SessionInfo
}

func (opt *GetSessionInfoRequest) Validity(resp *GetSessionInfoResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_session" // 请选择会话
		return
	}
}
