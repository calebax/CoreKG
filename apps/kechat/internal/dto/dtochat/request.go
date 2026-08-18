package dtochat

import (
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ChatQuestionStreamRequest struct {
	apiobj.BaseRequest
	Request ChatQuestionStreamEmbedRequest
}

type ChatQuestionStreamEmbedRequest struct {
	// QuestionID 问题id
	QuestionID string `json:"question_id" validate:"required"`
}

func (opt *ChatQuestionStreamRequest) Validity(resp *ChatQuestionStreamResponse) {
	if opt.Request.QuestionID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_select_session" // 请选择会话
		return
	}
}

type ExpansionQuestionRequest struct {
	apiobj.BaseRequest
	Request ExpansionQuestionEmbedRequest `json:"request"`
}
type ExpansionQuestionEmbedRequest struct {
	FileIDS   types.UintArray `json:"file_ids"`
	Question  string          `json:"question"`
	SessionID uint            `json:"session_id"`
}

func (opt *ExpansionQuestionRequest) Validity(resp *ExpansionQuestionResponse) {
	if len(opt.Request.FileIDS) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "资源id不能为空" // 扩写ID列表不能为空
		return
	}
	if opt.Request.Question == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "扩写问题不能为空" // 扩写问题不能为空
		return
	}
}
