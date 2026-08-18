package apis

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

// NewChatQuestionRequest 新建问题请求
type NewChatQuestionRequest struct {
	apiobj.BaseRequest

	Request struct {
		// ChatSessionID 群ID
		SessionID uint `json:"session_id"`
		// Question 问题
		Question     string   `json:"question"`
		ImageUrlList []string `json:"image_url_list"`

		// Options 行为开关（如是否联网搜索）
		Options *ChatOptions `json:"options,omitempty"`

		// Input 用户输入上下文资源（附件、文件、URL、文本块等）
		Input *ChatInput `json:"input,omitempty"`
	}
}

type ChatOptions struct {
	EnableWebSearch bool `json:"enable_web_search,omitempty"`
}

type ChatInput struct {
	Attachments []AttachmentInfo `json:"attachments,omitempty"`
}

type AttachmentInfo struct {
	URL   string `json:"url"`
	MdURL string `json:"md_url"`
	Type  string `json:"type"` // image / pdf / doc / csv / url / audio ...
	Name  string `json:"name,omitempty"`
}

func (req *NewChatQuestionRequest) Validity(resp *NewChatQuestionResponse) {
	if req.Request.SessionID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_session_id" // 参数错误
		return
	}
}

// NewChatQuestionResponse 新建问题返回
type NewChatQuestionResponse struct {
	apiobj.BaseResponse

	Response struct {
		// QuestionID 问题ID
		QuestionID string `json:"question_id"`
	}
}

// NewChatQuestionRequest 新建问题请求
type GetQuestionInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		QuestionID string `json:"question_id"`
	}
}

func (req *GetQuestionInfoRequest) Validity(resp *GetQuestionInfoResponse) {
	if req.Request.QuestionID == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_session_id" // 参数错误
		return
	}
}

// NewChatQuestionResponse 新建问题返回
type GetQuestionInfoResponse struct {
	apiobj.BaseResponse

	Response struct {
		Question Question `json:"question"`
	}
}
