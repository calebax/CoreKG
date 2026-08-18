package modes

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/kechat/chat/modelhelper"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	"github.com/ygpkg/yg-go/logs"
)

func getForestChatHistory(ctx context.Context, session *chattype.ChatSession) ([]schema.Message, error) {
	quesrions, err := chatquestion.ListSessionQuestionsByUin(ctx, session.Uin, session.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestChatHistory ListSessionQuestions error: %v", err)
		return nil, err
	}
	msgs := []schema.Message{}
	for _, qa := range quesrions {
		if qa.Source.Status != chattype.QuestionStatusAnswered {
			continue
		}
		source := qa.Source
		userMessage := schema.Message{
			Role:    schema.User,
			Content: source.Question,
		}

		if source.Extra != nil && source.Extra.Input != nil {
			inputParts := []schema.MessageInputPart{}
			for _, attachment := range source.Extra.Input.Attachments {
				url := attachment.Url
				if attachment.MdUrl != "" {
					url = attachment.MdUrl
				}
				inputParts = append(inputParts, schema.MessageInputPart{
					Type: schema.ChatMessagePartTypeFileURL,
					File: &schema.MessageInputFile{
						Name: attachment.Name,
						MessagePartCommon: schema.MessagePartCommon{
							// TODO 当前type 与 MIMEType不匹配，后面看情况调整
							MIMEType: attachment.Type,
							URL:      &url,
						},
					},
				})
			}
			if len(inputParts) > 0 {
				// prepend text part
				textPart := schema.MessageInputPart{
					Type: schema.ChatMessagePartTypeText,
					Text: source.Question,
				}
				userMessage.UserInputMultiContent = append([]schema.MessageInputPart{textPart}, inputParts...)
			}
		}
		msgs = append(msgs, userMessage)
		msgs = append(msgs, schema.Message{
			Role:    schema.Assistant,
			Content: source.Answer,
		})
	}
	return msgs, nil
}

type chatModelOptions = modelhelper.ToolCallingChatModelOptions

func ptrFloat32(v float32) *float32 {
	return &v
}

func newToolCallingChatModel(
	ctx context.Context,
	chatModel *chattype.ChatModel,
	options chatModelOptions,
) (model.ToolCallingChatModel, error) {
	return modelhelper.NewToolCallingChatModel(ctx, chatModel, options)
}

// sendQuestionRewriteMessage 发送工具消息
func sendQuestionRewriteMessage(ctx context.Context, msgPrinter printer.Printer, rewrite string) (*models.WriteResult, error) {
	res := &models.AgentResponse{
		MessageID:   uuid.NewString(),
		MessageType: models.MsgTypeQuestionRewrite,
		MessageTime: time.Now().UnixMilli(),
		TaskThought: rewrite,
		IsFinal:     true,
		Finish:      true,
	}
	msgPrinter.Send(ctx, "", models.MsgTypeCustomize, res, true)
	return &models.WriteResult{
		Content: res,
		Flag:    models.FlagCustomize,
	}, nil
}
