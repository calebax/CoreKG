package svcforestchat

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	agentmodels "github.com/insmtx/corekg/pkgs/einotools/models"
)

type NoopPrinter struct{}

func NewNoopPrinter() *NoopPrinter {
	return &NoopPrinter{}
}

func (p *NoopPrinter) Send(context.Context, string, string, any, bool) {}

func (p *NoopPrinter) SendFinalMsg(ctx context.Context, msgType string, msg any) {
	p.Send(ctx, uuid.NewString(), msgType, msg, true)
}

func (p *NoopPrinter) SetCancelFunc(func()) {}

func (p *NoopPrinter) Close(context.Context) {}

type OpenAIPrinter struct {
	ctx      *gin.Context
	id       string
	created  int64
	model    string
	started  bool
	closed   bool
	roleSent bool
	emitted  strings.Builder
	partials map[string]string
	cancelFn func()
}

func NewOpenAIPrinter(ctx *gin.Context, id string, created int64, model string) *OpenAIPrinter {
	return &OpenAIPrinter{
		ctx:      ctx,
		id:       id,
		created:  created,
		model:    model,
		partials: make(map[string]string),
	}
}

func (p *OpenAIPrinter) Send(ctx context.Context, msgID string, msgType string, msg any, isFinal bool) {
	if p.closed || msgType != agentmodels.MsgTypeResult {
		return
	}
	content, ok := msg.(string)
	if !ok || content == "" {
		return
	}
	if filtered, parsed := filterVisibleAnswer(content); parsed {
		content = filtered
	}
	if content == "" {
		return
	}
	delta := p.resolveDelta(msgID, content, isFinal)
	if delta == "" {
		return
	}
	_ = p.writeContent(delta)
}

func (p *OpenAIPrinter) SendFinalMsg(ctx context.Context, msgType string, msg any) {
	p.Send(ctx, uuid.NewString(), msgType, msg, true)
}

func (p *OpenAIPrinter) SetCancelFunc(cancel func()) {
	p.cancelFn = cancel
}

func (p *OpenAIPrinter) SetID(id string) {
	p.id = id
}

func (p *OpenAIPrinter) Close(context.Context) {
	_ = p.Finish("")
}

func (p *OpenAIPrinter) HasStarted() bool {
	return p.started
}

func (p *OpenAIPrinter) Finish(answer string) error {
	if p.closed {
		return nil
	}

	answer = normalizeAnswer(answer)
	if answer != "" {
		emitted := p.emitted.String()
		switch {
		case !p.started:
			if err := p.writeContent(answer); err != nil {
				return err
			}
		case strings.HasPrefix(answer, emitted):
			remainder := answer[len(emitted):]
			if remainder != "" {
				if err := p.writeContent(remainder); err != nil {
					return err
				}
			}
		case emitted == "":
			if err := p.writeContent(answer); err != nil {
				return err
			}
		}
	}

	if err := p.writeFinishChunk(); err != nil {
		return err
	}
	if _, err := p.ctx.Writer.Write([]byte("data: [DONE]\n\n")); err != nil {
		return err
	}
	p.flush()
	p.closed = true
	return nil
}

func (p *OpenAIPrinter) resolveDelta(msgID string, content string, isFinal bool) string {
	if msgID == "" {
		return content
	}
	if !isFinal {
		p.partials[msgID] += content
		return content
	}

	accumulated := p.partials[msgID]
	delete(p.partials, msgID)
	if accumulated != "" && strings.HasPrefix(content, accumulated) {
		return content[len(accumulated):]
	}
	return content
}

func (p *OpenAIPrinter) writeContent(content string) error {
	if content == "" {
		return nil
	}
	delta := dtokeapi.OpenAIChatDelta{Content: content}
	if !p.roleSent {
		delta.Role = "assistant"
		p.roleSent = true
	}
	if err := p.writeChunk(dtokeapi.OpenAIChatChunk{
		ID:      p.id,
		Object:  "chat.completion.chunk",
		Created: p.created,
		Model:   p.model,
		Choices: []dtokeapi.OpenAIChatChunkChoice{
			{
				Index:        0,
				Delta:        delta,
				FinishReason: nil,
			},
		},
	}); err != nil {
		return err
	}
	p.emitted.WriteString(content)
	return nil
}

func (p *OpenAIPrinter) writeFinishChunk() error {
	finishReason := "stop"
	return p.writeChunk(dtokeapi.OpenAIChatChunk{
		ID:      p.id,
		Object:  "chat.completion.chunk",
		Created: p.created,
		Model:   p.model,
		Choices: []dtokeapi.OpenAIChatChunkChoice{
			{
				Index:        0,
				Delta:        dtokeapi.OpenAIChatDelta{},
				FinishReason: &finishReason,
			},
		},
	})
}

func (p *OpenAIPrinter) writeChunk(chunk dtokeapi.OpenAIChatChunk) error {
	if err := p.ensureStarted(); err != nil {
		return err
	}
	payload, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	if _, err := p.ctx.Writer.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
		return err
	}
	p.flush()
	return nil
}

func (p *OpenAIPrinter) ensureStarted() error {
	if p.started {
		return nil
	}
	p.ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	p.ctx.Writer.Header().Set("Cache-Control", "no-cache")
	p.ctx.Writer.Header().Set("Connection", "keep-alive")
	p.ctx.Status(http.StatusOK)
	p.started = true
	return nil
}

func (p *OpenAIPrinter) flush() {
	if flusher, ok := p.ctx.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
