package printer

import (
	"context"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/ygpkg/yg-go/logs"
)

type LogPrinter struct {
	agentRequest *models.AgentRequest
	cancelFn     func()
}

func NewLogPrinter(agentRequest *models.AgentRequest) *LogPrinter {
	return &LogPrinter{agentRequest: agentRequest}
}

func (p *LogPrinter) Send(ctx context.Context, msgId string, msgType string, msg any, isFinal bool) {
	logs.InfoContextf(ctx, "%s: %s, %s, %v, %v", p.agentRequest.RequestID, msgId, msgType, msg, isFinal)
}

func (p *LogPrinter) SendFinalMsg(ctx context.Context, msgType string, msg any) {
	p.Send(ctx, uuid.NewString(), msgType, msg, true)
}

func (p *LogPrinter) SetCancelFunc(cancel func()) {
	p.cancelFn = cancel
}

func (p *LogPrinter) Close(ctx context.Context) {
	logs.InfoContextf(ctx, "LogPrinter closed")
}
