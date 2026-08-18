package qachatnodes

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/logs"
)

type intent struct {
	*baseHandler
}

func newIntent() *intent {
	return &intent{
		baseHandler: &baseHandler{},
	}
}

func (i *intent) NewReferencesIntentLambdaNode(ctx context.Context, input string, opts ...any) (output string, err error) {
	err = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
		req := &chattype.ChatRequestBody{
			Stream: false,
			Model:  chatagent.GetAgentI18nName(ctx, "", global.ChatAgentInetentionRecognition),
			ChatOptions: chattype.ChatOptions{
				Input: []chattype.Input{
					{Name: "input1", Value: state.QuestionEntity.Source.Question},
				},
			},
		}
		wrapper, err := chatclient.NewInternalChat(state.Ctx, state.QuestionEntity.Source.ReqID, "", 1, req)
		if err != nil {
			logs.ErrorContextf(ctx, "[NewReferencesIntentLambdaNode] failed to create internal chat: %v", err)
			return err
		}
		res, err := wrapper.AgentChatInternal(nil)
		if err != nil {
			logs.ErrorContextf(ctx, "[NewReferencesIntentLambdaNode] agent chat error: %v", err)
			return err
		}
		output = res.Content
		return nil
	})
	return output, nil
}
