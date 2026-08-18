package chat

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/insmtx/corekg/apps/kechat/chat/core"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/ygpkg/yg-go/apis/constants"
)

func TestChatWrapper_Run(t *testing.T) {

	testutils.Initialize(testutils.AppNameKechat)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(810))
	ctx.Set(constants.CtxKeyLang, "")

	questionID := "O3ioaZsBHGzjvqr_FTVr"

	questionEntity, err := chatquestion.GetQuetionByID(ctx, questionID)
	if err != nil {
		t.Fatalf("failed to get question, err: %v", err)
	}
	sessionEntity, err := chat.NewChatSessionsDao().GetByID(ctx, questionEntity.Source.SessionID)
	if err != nil {
		t.Fatalf("failed to get session, err: %v", err)
	}
	chatModelEntity, err := chat.NewChatModelDao().GetByID(ctx, sessionEntity.ModelID)
	if err != nil {
		t.Fatalf("failed to get model, err: %v", err)
	}

	wrapper := NewChatWrapper(ctx, &core.ChatContext{
		Session:  sessionEntity,
		Question: questionEntity,
		Model:    chatModelEntity,
	})
	result, err := wrapper.Run(ctx)
	if err != nil {
		t.Fatalf("failed to run wrapper, err: %v", err)
	}

	jsonData, _ := json.Marshal(result)
	log.Printf("Result JSON: %s\n", jsonData)
}
