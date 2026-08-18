package svcchat

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
)

// GetForestChatHistory 获取chat历史记录
func GetForestChatHistory(ctx context.Context, session *chattype.ChatSession) ([]*schema.Message, error) {
	quesrions, err := chatquestion.ListSessionQuestionsByUin(ctx, session.Uin, session.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestChatHistory ListSessionQuestions error: %v", err)
		return nil, err
	}
	msgs := []*schema.Message{}
	for _, qa := range quesrions {
		if qa.Source.Status != chattype.QuestionStatusAnswered {
			continue
		}
		msgs = append(msgs, &schema.Message{
			Role:    schema.User,
			Content: qa.Source.Question,
		})
		msgs = append(msgs, &schema.Message{
			Role:    schema.Assistant,
			Content: qa.Source.Answer,
		})
	}
	return msgs, nil
}
