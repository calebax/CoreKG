package core

import (
	"strconv"
	"strings"

	"github.com/insmtx/corekg/apps/kechat/chat/modelhelper"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
)

const (
	ExtraKeyPrinter             = "printer"
	ExtraKeySummarySystemPrompt = "summary_system_prompt"
	ExtraKeyEnableReference     = "enable_reference"
)

type ChatContext struct {
	Session      *chattype.ChatSession
	Question     *chattype.ChatQuestion
	Model        *chattype.ChatModel
	ModelOptions ChatModelOptions
	Extra        map[string]any
}

type ChatModelOptions = modelhelper.ToolCallingChatModelOptions

func GetPrinter(extra map[string]any) printer.Printer {
	if len(extra) == 0 {
		return nil
	}
	value, ok := extra[ExtraKeyPrinter]
	if !ok || value == nil {
		return nil
	}
	p, ok := value.(printer.Printer)
	if !ok {
		return nil
	}
	return p
}

func GetSummarySystemPrompt(extra map[string]any) string {
	if len(extra) == 0 {
		return ""
	}
	value, ok := extra[ExtraKeySummarySystemPrompt]
	if !ok || value == nil {
		return ""
	}
	prompt, ok := value.(string)
	if !ok {
		return ""
	}
	return prompt
}

func GetEnableReference(extra map[string]any) bool {
	if len(extra) == 0 {
		return true
	}
	value, ok := extra[ExtraKeyEnableReference]
	if !ok || value == nil {
		return true
	}
	switch v := value.(type) {
	case bool:
		return v
	case *bool:
		if v == nil {
			return true
		}
		return *v
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return true
		}
		return parsed
	default:
		return true
	}
}
