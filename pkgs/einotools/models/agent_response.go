package models

import (
	"encoding/json"
)

type FlagAnswer string

const (
	FlagAgent       FlagAnswer = "agent"
	FlagFinalResult FlagAnswer = "result"
	FlagFound       FlagAnswer = "found"
	FlagSearching   FlagAnswer = "searching"

	FlagCustomize FlagAnswer = "customize"
)

type WriteResult struct {
	ReasoningContent string     `json:"reasoning_content"`
	ReasoningSeconds int        `json:"reasoning_seconds"`
	Content          any        `json:"content"`
	Flag             FlagAnswer `json:"flag"`
	Reference        any        `json:"reference"`
}

func (w WriteResult) String() string {
	jsonData, err := json.Marshal(w)
	if err != nil {
		return "\n"
	}
	return string(jsonData) + "\n"
}

type AgentResponse struct {
	TaskID       string         `json:"task_id"`
	TaskOrder    int            `json:"task_order"`
	MessageID    string         `json:"message_id"`
	MessageOrder int            `json:"message_order"`
	MessageTime  int64          `json:"message_time"`
	MessageType  string         `json:"message_type"`
	ResultMap    map[string]any `json:"result_map"`
	TaskThought  string         `json:"task_thought"`
	Finish       bool           `json:"finish"`
	IsFinal      bool           `json:"is_final"`
}

const (
	MsgTypeTaskThought     = "task_thought"
	MsgTypeResult          = "result"
	MsgTypeFileView        = "file"
	MsgTypeCodeExec        = "code"
	MsgTypeChartGenerate   = "chart"
	MsgTypeExecFlag        = "exec_flag"
	MsgTypeKnowledgeSearch = "forest_search_tool"

	MsgTypeCustomize       = "customize"        // 自定义工具
	MsgTypeQuestionRewrite = "question_rewrite" // 问题改写
)

func (r *AgentResponse) BuildPartialResult(eventResult *EventResult) {
	r.TaskID = eventResult.GetTaskID()
	r.TaskOrder = eventResult.GetTaskOrder()
	r.MessageOrder = eventResult.GetAndIncrOrder(r.TaskID + ":" + r.MessageID)
}

type ToolResponse struct {
	ToolShell  string `json:"shell"`
	ToolResult any    `json:"tool_result"`
}
