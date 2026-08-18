package essearch

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// ChatSubQuestion schedule chat with
func ChatSubQuestion(ctx context.Context, question string) ([]esquery.Map, error) {
	var esMap []esquery.Map
	subQi, err := subquestion(ctx, question)
	if err != nil {
		return nil, err
	}

	for _, q := range subQi {
		esMap = append(esMap, esquery.BuildMap("match", esquery.BuildMap("description", q)))
	}
	return esMap, nil
}

// subquestion split subquestion search all related chunk
func subquestion(ctx context.Context, question string) ([]string, error) {
	subquestion, err := ChatWithCommandAgent(question)
	if err != nil {
		return nil, err
	}
	var subs *Subquestion
	if err = json.Unmarshal([]byte(subquestion), &subs); err != nil {
		logs.ErrorContextf(ctx, "unmarshal subquestion error: %v , question: %s, subquestion: %s", err, question, subquestion)
		return []string{question}, nil
	}

	//push original question
	subs.Subquery = append(subs.Subquery, question)
	return subs.Subquery, nil
}

// Subquestion
type Subquestion struct {
	Subquery []string `json:"subquerys"`
}

func ChatWithCommandAgent(question string) (string, error) {
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(context.Background(), "", global.ChatAgentSubQuestionChat),
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: question},
			},
		},
	}
	// TODO reqid @shikefan
	ctx := &gin.Context{}
	w, err := chatclient.NewInternalChat(ctx, "essearch", "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return "", err
	}
	return res.Content, nil
}
