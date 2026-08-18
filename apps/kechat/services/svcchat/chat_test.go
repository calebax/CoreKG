package svcchat

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtochat"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chatsession"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func TestChatQuestionStream(t *testing.T) {
	testutils.Initialize(testutils.AppNameKechat)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))

	questionID := "_80g7p4BjMOTYrjomD4p"
	t.Logf("using question_id=%s", questionID)

	streamReq := &dtochat.ChatQuestionStreamRequest{
		Request: dtochat.ChatQuestionStreamEmbedRequest{
			QuestionID: questionID,
		},
	}
	streamResp, err := ChatQuestionStream(ctx, streamReq)
	assert.Nil(t, err)
	assert.NotNil(t, streamResp)
	t.Log("streamResp:", logs.JSON(streamResp))
}

func TestChatQuestionStreamMultiRun(t *testing.T) {
	testutils.Initialize(testutils.AppNameKechat)
	defer testutils.Close()

	const totalRuns = 5
	const question = "分析预算和结算的工程量差异项"

	for i := 0; i < totalRuns; i++ {
		run := i + 1
		t.Logf("")
		t.Logf("================================================================")
		t.Logf("========== [MULTI] RUN %d/%d START ===========================", run, totalRuns)
		t.Logf("================================================================")

	ctx := testutils.NewCtx(testutils.WithUin(384))
	ctx.Set(constants.CtxKeyLang, "")

	uin := runtime.Uin(ctx)
		companyID := runtime.CompanyID(ctx)
		reqID := runtime.RequestID(ctx)

		session, err := createSession(ctx, uin, companyID)
		if err != nil {
			t.Fatalf("[RUN %d] createSession failed: %v", run, err)
		}
		t.Logf("[RUN %d] session_id=%d created", run, session.ID)

		qID, err := createQuestion(ctx, session.ID, uin, companyID, reqID, question)
		if err != nil {
			t.Fatalf("[RUN %d] createQuestion failed: %v", run, err)
		}
		t.Logf("[RUN %d] question_id=%s created", run, qID)

		streamReq := &dtochat.ChatQuestionStreamRequest{
			Request: dtochat.ChatQuestionStreamEmbedRequest{
				QuestionID: qID,
			},
		}
		streamResp, err := ChatQuestionStream(ctx, streamReq)
		if err != nil {
			t.Logf("[RUN %d] ChatQuestionStream error: %v", run, err)
		} else {
			assert.NotNil(t, streamResp)
			t.Logf("[RUN %d] ChatQuestionStream completed, resp=%s", run, logs.JSON(streamResp))
		}

		t.Logf("========== [MULTI] RUN %d/%d END =============================", run, totalRuns)
	}
}

func createSession(ctx *gin.Context, uin, companyID uint) (*chattype.ChatSession, error) {
	sess := &chattype.ChatSession{
		CompanyID:    companyID,
		Uin:          uin,
		ResourceType: chattype.ResourceTypeFileList,
		BaseType:     chattype.ResourceQASessionBaseTypeStandard,
		ModelID:      1,
		ModelName:    "deepseek-v3",
		EsIndex:      "ke_0",
		FileIDList:   types.NewUintArray([]uint{11537, 11538}),
	}
	if err := chatsession.CreateSession(ctx, sess); err != nil {
		return nil, fmt.Errorf("chatsession.CreateSession: %w", err)
	}
	return sess, nil
}

func createQuestion(ctx *gin.Context, sessionID uint, uin, companyID uint, reqID, question string) (string, error) {
	ques := &chattype.ChatQuestion{
		Source: &chattype.Question{
			CompanyID:  companyID,
			Uin:        uin,
			ReqID:      reqID,
			Question:   question,
			Status:     chattype.QuestionStatusPending,
			SessionID:  sessionID,
			ModelID:    1,
		},
	}
	if err := chatquestion.CreateQuestion(ctx, ques); err != nil {
		return "", fmt.Errorf("chatquestion.CreateQuestion: %w", err)
	}
	return ques.ID, nil
}
