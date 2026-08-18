package apis

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kechat/models/qachat"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

// CreateAgent 创建指令型机器人
// Deprecated:this api func was deprecated try use CreateAgentApp to instead it
// @Tags 机器人
// @Summary 创建指令型机器人
// @Description 创建指令型机器人
// @Router /chat.CreateAgent [post]
// @Param request body CreateAgentRequest true "入参"
// @Success 200 {object} CreateAgentResponse
func CreateAgent(ctx *gin.Context, req *CreateAgentRequest, resp *CreateAgentResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreateAgent validate params failed")
		return
	}
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)

	req.Request.Uin = uin
	req.Request.CompanyID = companyID

	// 创建指令型机器人
	err := chatagent.CreateAgent(ctx, &req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_create_agent_failed" // 创建指令型机器人失败
		logs.ErrorContextf(ctx, "CreateAgent failed ,err %s", err)
		return
	}
}

// UpdateAgent 编辑指令型机器人
// Deprecated:this api func was deprecated try to use perm.UpdateAgentWithPerm to instead it
// @Tags 机器人
// @Summary 编辑指令型机器人
// @Description 编辑指令型机器人
// @Router /chat.UpdateAgent [post]
// @Param request body UpdateAgentRequest true "入参"
// @Success 200 {object} UpdateAgentResponse
func UpdateAgent(ctx *gin.Context, req *UpdateAgentRequest, resp *UpdateAgentResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "UpdateCommandAgent validate params failed")
		return
	}

	err := chatagent.UpdateAgent(ctx, &req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_update_agent_failed" // 编辑指令型机器人失败
		logs.ErrorContextf(ctx, "UpdateCommandAgent failed, err %s", err)
		return
	}
}

// GetAgentDetail 获取指令型机器人详情
// Deprecated:this api func was deprecated try to use perm.GetAgentWithPerm to instead it
// @Tags 机器人
// @Summary 获取指令型机器人详情
// @Description 获取指令型机器人详情
// @Router /chat.GetAgentDetail [post]
// @Param request body GetAgentDetailRequest true "入参"
// @Success 200 {object} GetAgentDetailResponse
func GetAgentDetail(ctx *gin.Context, req *GetAgentDetailRequest, resp *GetAgentDetailResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "GetAgentDetail validate params failed")
		return
	}

	// 获取机器人信息
	agentInfo, err := chatagent.GetChatAgentWithPermByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_detail_failed" // 获取基本信息失败
		logs.ErrorContextf(ctx, "GetAgentDetail failed ,err %s", err)
		return
	}

	// 获取机器人类型额外信息
	promptAgent, err := chatagent.GetAgentTypeByID(ctx, agentInfo.Version)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_agent_type_failed" // 获取指令类型信息失败
		logs.ErrorContextf(ctx, "GetAgentDetail failed ,err %s", err)
		return
	}

	resp.Response.AgentWithPerm = *agentInfo
	resp.Response.AgentItemInfo = *promptAgent
}

// TestAgent 测试指令型机器人
// @Tags 机器人
// @Summary 测试指令型机器人
// @Description 测试指令型机器人
// @Router /chat.TestAgent [post]
// @Param request body TestAgentRequest true "入参"
func TestAgent(ctx *gin.Context) {
	var resp AgentChatResponse
	reqBodyData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] TestAgent ioutil.ReadAll failed, %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_read_request_body_failed" // 读取请求体失败
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	req := &TestAgentRequest{}
	if err := json.Unmarshal(reqBodyData, req); err != nil {
		logs.ErrorContextf(ctx, "[chat] TestAgent BindJSON failed, %s", err.Error())
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_bind_json_failed" // 绑定 JSON 失败
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	if req.Validity(ctx, &resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "TestAgent validate params failed")
		return
	}

	// 获取模型信息
	model, err := chatmodel.GetModelByID(ctx, req.Request.ChatModelIDs.Slice()[0])
	if err != nil {
		logs.ErrorContextf(ctx, "TestAgent GetModelByID failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_get_model_failed" // 获取模型失败
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	chatBody := &chattype.ChatRequestBody{
		Model:      req.Request.agentInfo.Name,
		LLMModelID: model.ID,
		Stream:     true,
		ChatOptions: chattype.ChatOptions{
			Input: req.Request.Input,
		},
	}

	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:    runtime.CompanyID(ctx),
			Uin:          runtime.Uin(ctx),
			ReqID:        runtime.RequestID(ctx),
			Status:       chattype.QuestionStatusPending,
			BaseAgentID:  req.Request.agentInfo.ID,
			AgentVersion: req.Request.agentInfo.Version,
			ModelID:      model.ID,
			UserInput:    chatBody,
			AgentName:    req.Request.agentInfo.Name,
			Question:     req.Request.Question,
		},
	}
	wrapper := qachat.NewChatAPIWrapper(ctx, ques, model, req.Request.agentInfo)
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(ctx, ques.Source.ReqID)
	sseClient.SetHeaders(ctx.Writer)
	_, err = wrapper.AgentChatInternal(func(chunk *chattype.ChatStreamResponseBody) error {
		writeResult := llmchat.WriteResult{
			ReasoningContent: chunk.Choices[0].Delta.ReasoningContent,
			Content:          chunk.Choices[0].Delta.Content,
		}
		if stoped, err := sseClient.WriteMessage(ctx, ctx.Writer, ques.Source.ReqID, writeResult.String()); err != nil {
			logs.ErrorContextf(ctx, "AgentChat write response error: %s", err)
			return err
		} else if stoped {
			logs.InfoContextf(ctx, "AgentChat stream Stoped by client")
			return nil
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "TestAgent AgentAPIChat failed ,err %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kechat_chat_failed"
		ctx.JSON(http.StatusBadRequest, resp)
		// return
		ques.Source.Status = chattype.QuestionStatusError
	}

}
