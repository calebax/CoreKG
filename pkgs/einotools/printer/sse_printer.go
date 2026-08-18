package printer

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	"github.com/ygpkg/yg-go/apis/sseclient"
)

type SSEPrinter struct {
	agentRequest *models.AgentRequest
	eventResult  *models.EventResult
	sseClient    *sseclient.SSEClient
	writer       io.Writer
	cancelFn     func()
}

func NewSSEPrinter(agentRequest *models.AgentRequest, sseClient *sseclient.SSEClient, writer io.Writer) *SSEPrinter {
	return &SSEPrinter{
		agentRequest: agentRequest,
		eventResult:  models.NewEventResult(),
		sseClient:    sseClient,
		writer:       writer,
	}
}

func (p *SSEPrinter) SetCancelFunc(cancel func()) {
	p.cancelFn = cancel
}

func (p *SSEPrinter) Send(ctx context.Context, msgId string, msgType string, msg any, isFinal bool) {
	if len(msgId) == 0 {
		msgId = uuid.NewString()
	}

	if msgType == models.MsgTypeExecFlag {
		p.handleExecFlag(ctx, msg)
		return
	}
	if msgType == models.MsgTypeResult {
		p.handleResult(ctx, msg, isFinal)
		return
	}
	if msgType == models.MsgTypeKnowledgeSearch {
		p.handleKnowledgeSearch(ctx, msg, isFinal)
		return
	}
	if msgType == models.MsgTypeCustomize {
		p.handleKCustomize(ctx, msg)
		return
	}

	finish := msgType == models.MsgTypeResult

	agentResponse := &models.AgentResponse{
		MessageID:   msgId,
		MessageType: msgType,
		MessageTime: time.Now().UnixMilli(),
		IsFinal:     isFinal,
		Finish:      finish,
	}

	resultMap := utils.ConvertMsgResultMap(ctx, agentResponse, msgType, msg)
	if resultMap == nil {
		return
	}

	agentResponse.ResultMap = resultMap

	agentResponse.BuildPartialResult(p.eventResult)

	result := &models.WriteResult{
		Content: agentResponse,
		Flag:    models.FlagAgent,
	}

	stop, _ := p.sseClient.WriteMessage(ctx, p.writer, p.agentRequest.RequestID, result.String())
	if stop {
		p.triggerCancel()
	}
}

func (p *SSEPrinter) SendFinalMsg(ctx context.Context, msgType string, msg any) {
	p.Send(ctx, uuid.NewString(), msgType, msg, true)
}

func (p *SSEPrinter) Close(ctx context.Context) {
	p.sseClient.Close(ctx, p.agentRequest.RequestID)
}

func (p *SSEPrinter) triggerCancel() {
	if p.cancelFn != nil {
		p.cancelFn()
	}
}

// 处理执行标志消息
func (p *SSEPrinter) handleExecFlag(ctx context.Context, msg any) {
	result := &models.WriteResult{
		Flag: msg.(models.FlagAnswer),
	}
	p.writeAndCheckStop(ctx, result.String())
}

// 处理结果消息
func (p *SSEPrinter) handleResult(ctx context.Context, msg any, isFinal bool) {
	// 兼容以前正文格式
	result := &models.WriteResult{
		Content: msg.(string),
	}
	if isFinal {
		result.Flag = models.FlagFinalResult
	}
	p.writeAndCheckStop(ctx, result.String())
}

// 处理知识库检索消息
type SearchResult struct {
	ResultType string      `json:"result_type"`
	ResultData interface{} `json:"result_data"`
}

func (p *SSEPrinter) handleKnowledgeSearch(ctx context.Context, msg any, _ bool) {

	if _, ok := msg.(*models.ToolResponse); ok {
		return
	}

	var result SearchResult
	json.Unmarshal([]byte(msg.(string)), &result)

	switch result.ResultType {
	case "search_normal_result":
		if data, ok := result.ResultData.([]interface{}); ok {
			for _, reference := range data {
				result := &models.WriteResult{
					Reference: reference,
					Flag:      models.FlagFound,
				}
				p.writeAndCheckStop(ctx, result.String())
			}
		}
	case "search_qa_result":
		return
	}

}

// 通用的写入和停止检查逻辑
func (p *SSEPrinter) writeAndCheckStop(ctx context.Context, message string) {
	stop, _ := p.sseClient.WriteMessage(ctx, p.writer, p.agentRequest.RequestID, message)
	if stop {
		p.triggerCancel()
	}
}

func (p *SSEPrinter) handleKCustomize(ctx context.Context, msg any) {
	result := &models.WriteResult{
		Content: msg,
		Flag:    models.FlagCustomize,
	}
	stop, _ := p.sseClient.WriteMessage(ctx, p.writer, p.agentRequest.RequestID, result.String())
	if stop {
		p.triggerCancel()
	}
}
