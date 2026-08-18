package svcchat

import (
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochat"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/services/svcchat/agent"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ExpansionQuestion 问题扩写
func ExpansionQuestion(ctx *gin.Context, req *dtochat.ExpansionQuestionRequest) (res *dtochat.ExpansionQuestionResponse, err error) {
	res = &dtochat.ExpansionQuestionResponse{}
	// 搜索摘要
	wrapper, err := essearch.NewEsSearchWrapper(ctx, "ke_0", req.Request.Question, []uint{}, req.Request.FileIDS.Slice())
	if err != nil {
		logs.ErrorContextf(ctx, "ExpansionQuestion NewEsSearchWrapper error: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "创建搜索包装器失败"
		return res, nil
	}
	searchResult, err := wrapper.SummarizeSearch()
	if err != nil {
		logs.ErrorContextf(ctx, "ExpansionQuestion SummarizeSearch error: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "搜索摘要失败"
		return res, nil
	}
	var desc []string
	for i, hit := range searchResult.Hits.Hits {
		desc = append(desc, hit.Source.Description)
		if i >= 10 {
			break
		}
	}

	var historyList historyList
	if req.Request.SessionID > 0 {
		chats, err := chatquestion.ListSessionQuestionsByUin(ctx, runtime.Uin(ctx), req.Request.SessionID)
		if err != nil {
			logs.ErrorContextf(ctx, "ListSessionChats error: %v", err)
			res.Code = errcode.ErrCode_InternalError
			res.Message = "获取会话问题失败" // 服务器错误
			return res, nil
		}
		for _, v := range chats {
			historyList = append(historyList, history{
				Question: v.Source.Question,
				Answer:   v.Source.Answer,
			})
		}
	}
	model, err := chatmodel.GetModelByID(ctx, 1)
	if err != nil {
		logs.ErrorContextf(ctx, "GetModelByID err:%v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取模型信息失败"
		return res, nil
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  model.APIKey,
		Model:   model.ModelName,
		BaseURL: strings.TrimSuffix(model.ModelUrl, "/chat/completions"),
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewChatModel err:%v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "创建聊天模型失败"
		return res, nil
	}

	newquestion, err := agent.ExecuteExpansionAgent(ctx, chatModel, req.Request.Question, logs.JSON(historyList), logs.JSON(desc))
	if err != nil {
		logs.ErrorContextf(ctx, "ExpansionQuestion ExecuteExpansionAgent error: %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "扩写问题失败"
		return res, nil
	}
	res.Response.ExpandedQuestion = newquestion

	return res, nil
}

type history struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type historyList []history
