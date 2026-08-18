package qachatnodes

import (
	"github.com/cloudwego/eino/compose"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
)

func init() {
	err := compose.RegisterSerializableType[State]("ChatAgentState")
	if err != nil {
		panic(err)
	}
}

func NewState(ctx *gin.Context) *State {
	return &State{
		Ctx:                     ctx,
		Records:                 make(nodebase.RecordList, 0),
		SessionEntity:           &chattype.ChatSession{},
		QuestionEntity:          &chattype.ChatQuestion{},
		ModelEntity:             &chattype.ChatModel{},
		QuestionDbDatasetEntity: &chattype.ChatQuestionDbDataset{},
	}
}

type State struct {
	Ctx *gin.Context
	// Records 自定义 kv 集合
	Records nodebase.RecordList `json:"records,omitempty"`
	// UserInput 用户输入
	UserInput string `json:"user_input,omitempty"`
	// HistoryContext 历史上下文
	HistoryContext string `json:"history_context,omitempty"`
	// SessionEntity 会话实体
	SessionEntity *chattype.ChatSession `json:"session_entity,omitempty"`
	// QuestionEntity 问题实体
	QuestionEntity *chattype.ChatQuestion `json:"question_entity,omitempty"`
	// ModelEntity 模型实体
	ModelEntity *chattype.ChatModel `json:"model_entity,omitempty"`
	// QuestionDbDataset 问题数据集
	QuestionDbDatasetEntity *chattype.ChatQuestionDbDataset `json:"question_db_dataset,omitempty"`

	// Goto 下一步流转的节点
	Goto               string                       `json:"goto,omitempty"`
	QueryReferenceList *chattype.QueryReferenceList `json:"query_reference_list,omitempty"`
	ChatReferenceList  *chattype.ChatReferenceList  `json:"chat_reference_list,omitempty"`
	// QAPairs 问答对
	QAPairs []QAPair `json:"qa_pairs,omitempty"`
}

type QAPair struct {
	ReqID string `json:"req_id,omitempty"`
	// QAQuestion 问答对问题
	QAQuestion string `json:"qa_question,omitempty"`
	// QAAnswer 问答对答案
	QAAnswer string `json:"qa_answer,omitempty"`
	// Answer 答案
	Answer string                  `json:"answer,omitempty"`
	Status chattype.QuestionStatus `json:"status,omitempty"`
}
