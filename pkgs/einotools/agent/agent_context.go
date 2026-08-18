package agent

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	"github.com/insmtx/corekg/pkgs/einotools/tools"
)

type AgentContext struct {
	ModelRoleName       string
	SessionID           uint
	RequestID           string
	AgentRequest        *models.AgentRequest
	Query               string
	ProductFiles        []*models.File
	DateInfo            string
	SystemPrompt        string
	NextStepPrompt      string
	SummarySystemPrompt string
	ChatModel           model.ToolCallingChatModel
	Tools               []tools.ToolOption
	AvailableTools      []tool.BaseTool
	IsStream            bool
	MaxStep             int
	Printer             printer.Printer
	SaveChartFunc       func(string) (uint, error)
}
