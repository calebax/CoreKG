package chattype

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

const (
	TableNameChatModel             = "chat_model"
	TableNameChatSessions          = "chat_sessions"
	TableNameAgent                 = "chat_agent"
	TableNameAgentVersion          = "chat_agent_version"
	TableNameAgentCollect          = "chat_agent_collect"
	TableNameChatMigrate           = "chat_migrate"
	TableNameChatQuestionDbDataset = "chat_question_db_dataset"
	TableNameChatChartCanvas       = "chat_chart_canvas"
	TableNameChatChart             = "chat_chart"
	TableNameChatCozeMapping       = "chat_coze_mapping"
	TableNameChatRecentUsedModel   = "chat_recent_used_model"
	TableNameModelInstance         = "model_instance"
)

// InitModel init db tables
func InitDB() error {
	err := dbtools.InitModel(dbutil.Chat(),
		&ChatSession{},
		&ChatModel{},
		&ChatAgentVersion{},
		&ChatAgent{},
		&ChatAgentCollect{},
		&ChatMigrate{},
		&ChatCozeMapping{},
		&ChatQuestionDbDataset{},
		&ChatChartCanvas{},
		&ChatChart{},
		&ChatRecentUsedModel{},
		&CozeModelInstance{},
	)

	if err != nil {
		return err
	}
	return nil
}
