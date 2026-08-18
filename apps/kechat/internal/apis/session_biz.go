package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// NewChatSessionRequest 新建群聊请求
type NewChatSessionRequest struct {
	apiobj.BaseRequest
	Request NewChatSessionEmbedRequest
}

type SourceFrom string

var (
	SourceFromFile  SourceFrom = "file"
	SourceFromAgent SourceFrom = "agent"
)

type NewChatSessionEmbedRequest struct {
	// Input 输入参数
	Input chattype.InputList `json:"input"`
	// ResourceID 资源ID
	ResourceID uint `json:"resource_id"`
	// 模型id
	ModelID uint `json:"model_id"`
	// BaseType session 基础类型，agent：轻应用，standard：标准知识库，model：模型问答，excel：excel，mysql：mysql
	BaseType chattype.ResourceQASessionBaseType `json:"base_type"`
	// ResourceType 资源类型，excel_list：excel 文件列表，excel_sheet_list：excel sheet 列表, db_list：数据库列表，db_table_list：数据库表列表
	ResourceType chattype.ResourceType `json:"resource_type"`
	// 各种资源的id数组
	IDS types.UintArray `json:"ids"`
	// Names 资源名称
	Names types.StringArray `json:"names"`
	// 项目id
	ProjectID uint `json:"project_id"`
	//SourceFrom 会话来源
	SourceFrom SourceFrom `json:"source_from"`
	// PromptKey 提示词key
	PromptKey string `json:"prompt_key"`
}

func (req *NewChatSessionRequest) Validity(ctx *gin.Context, resp *NewChatSessionResponse) {
	// 特殊处理
	if (req.Request.ResourceType == chattype.ResourceTypeExcelList || req.Request.ResourceType == chattype.ResourceTypeFileList) &&
		req.Request.ModelID == 0 {
		modelList := &chatmodel.QueryModelListResponse{}
		err := chatmodel.QueryModelList(ctx, apiobj.PageQuery{CompanyID: runtime.CompanyID(ctx)}, modelList)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kechat_get_model_failed" // 获取模型失败
			logs.ErrorContextf(ctx, "NewChatSessionRequest Validity QueryModelList err: %v", err)
			return
		}
		if len(modelList.Data) == 0 {
			logs.ErrorContextf(ctx, "NewChatSessionRequest Validity QueryModelList no model")
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_model_empty" // 模型为空
			return
		}
		req.Request.ModelID = modelList.Data[0].ID
		req.Request.ResourceID = modelList.Data[0].ID
	}

	if req.Request.ResourceID == 0 || req.Request.ResourceType == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	if req.Request.ResourceType == chattype.ResourceTypeModel {
		req.Request.ModelID = req.Request.ResourceID
	}
	//if req.Request.ModelID == 0 {
	//	resp.Code = errcode.ErrCode_BadRequest
	//	resp.Message = "kechat_no_model_specified" // 未指定对话模型
	//	return
	//}

	switch req.Request.PromptKey {
	case "", "normal", "concise", "study", "explanation", "formal":
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "未知提示词模式" // 未指定对话模型
		return
	}
}

// NewChatSessionResponse 新建群聊返回
type NewChatSessionResponse struct {
	apiobj.BaseResponse
	Response struct {
		*chattype.ChatSession `json:",inline"`
	}
}

// ListChatSessionRequest 获取群聊列表返回
type ListChatSessionRequest struct {
	apiobj.BaseResponse
	Request struct {
		apiobj.PageQuery
		ProjectID int `json:"project_id"`
	}
}

func (req *ListChatSessionRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "name", "created_at", "updated_at",
			"name desc", "created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_orderby_not_supported" // orderBy字段不支持
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "name", "resource_id", "resource_type":
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kechat_filter_value_empty" // 查询条件中的值不能为空
				return
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kechat_filter_field_invalid" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"field": v.Field,
			}
			return
		}
	}
}

// ListChatSessionResponse 获取群聊列表返回
type ListChatSessionResponse struct {
	apiobj.BaseResponse
	Response chatsession.QueryListChatSessionsResponse
}

// ListSessionChatsResponse 获取群聊消息列表返回
type ListSessionChatsResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data []*Question
	}
}

type Question struct {
	*chattype.ChatQuestion
	Msg []*models.WriteResult `json:"msg"`
}

// SetTopChatSessionResponse
type SetTopChatSessionResponse struct {
	apiobj.BaseResponse
}

// NewChatSessionResponse 新建群聊返回
type GetSessionInfoResponse struct {
	apiobj.BaseResponse
	Response *chatsession.SessionInfo
}

// WorkflowTestRunRequest 工作流试运行接口返回
type WorkflowTestRunRequest struct {
	apiobj.BaseRequest
	Request struct {
		WorkflowID string                      `json:"coze_workflow_id"`
		SpaceID    string                      `json:"coze_space_id"`
		Input      []coze.WorkflowTestRunField `json:"input"`
	} `json:"request"`
}

// WorkflowTestRunResponse 工作流试运行接口返回
type WorkflowTestRunResponse struct {
	apiobj.BaseResponse
	Response struct {
		Output string `json:"output"`
	}
}

// UpdateWorkflowVersionRequest 更新工作流版本
type UpdateWorkflowVersionRequest struct {
	apiobj.BaseRequest
	Request struct {
		WorkflowID string `json:"workflow_id"`
		Version    string `json:"version"`
	} `json:"request"`
}

type UpdateChatAgentInputRequest struct {
	apiobj.BaseRequest
	Request struct {
		AgentID uint                `json:"agent_id"`
		Params  chattype.ParamsList `json:"input"`
	} `json:"request"`
}
