package qachatnodes

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/connectors"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func TestToolsNode(t *testing.T) {
	ctx := &gin.Context{}
	dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
	})
	// TODO 在main中初始化
	err := connectors.InitProviders(ctx, "account", "pkl_connect_providers")
	if err != nil {
		// logs.Errorf("InitProviders error: %v", err)
		return
	}
	state := NewState(ctx)
	state.UserInput = "IN-48311029"
	state.QuestionEntity = &chattype.ChatQuestion{
		Source: &chattype.Question{
			Uin:      330,
			Question: "你好，看看。我查询的内容是什么,",
		},
	}
	state.SessionEntity = &chattype.ChatSession{
		Model: gorm.Model{
			ID: 5493,
		},
		Uin:     330,
		ModelID: 1,
		// ExternalTokenIDList: types.NewUintArray([]uint{2}),
	}
	genFunc := func(c context.Context) *State {
		return state
	}
	r, err := ExternalDataToolsBuilder[nodebase.RecordList, nodebase.RecordList](ctx, genFunc)
	if err != nil {
		logs.ErrorContextf(ctx, "r: %v, err: %v", r, err)
		return
	}
	state.Records.Add(&nodebase.Record{
		Key:   RecordKeyUseGmail,
		Value: RecordKeyUseGmail,
	})
	state.Records.Add(&nodebase.Record{
		Key:   RecordKeyUseTest,
		Value: RecordKeyUseTest,
	})
	state.Records.Add(&nodebase.Record{
		Key:   RecordKeyUseSlack,
		Value: RecordKeyUseGmail,
	})

	out, err := r.Invoke(ctx, state.Records)
	if err != nil {
		logs.ErrorContextf(ctx, "err: %v", err)
		return
	}
	for _, v := range out {
		logs.InfoContextf(ctx, "out: %+v", v)
	}

}

func TestGmailChat(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
	})
	ctx := &gin.Context{}
	chatquestion.InitHistoryESClient(ctx)
	// TODO 在main中初始化
	err := connectors.InitProviders(ctx, "account", "pkl_connect_providers")
	if err != nil {
		// logs.Errorf("InitProviders error: %v", err)
		return
	}
	state := NewState(ctx)
	state.UserInput = "IN-48311029"
	state.QuestionEntity = &chattype.ChatQuestion{
		Source: &chattype.Question{
			Uin:      671,
			Question: "IN-",
		},
	}
	state.SessionEntity = &chattype.ChatSession{
		Model: gorm.Model{
			ID: 5317,
		},
		Uin:     671,
		ModelID: 1,
	}
	genFunc := func(c context.Context) *State {
		return state
	}
	r, err := ExternalChatBuilder[nodebase.RecordList, nodebase.RecordList](ctx, genFunc)
	if err != nil {
		logs.ErrorContextf(ctx, "r: %v, err: %v", r, err)
		return
	}
	state.Records.Add(&nodebase.Record{
		Key:   RecordKeyUseGmail,
		Value: RecordKeyUseGmail,
	})

	out, err := r.Invoke(ctx, state.Records)
	if err != nil {
		logs.ErrorContextf(ctx, "err: %v", err)
		return
	}
	for _, v := range out {
		logs.InfoContextf(ctx, "out: %+v", v)
	}
}
