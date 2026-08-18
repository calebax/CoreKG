package svchotwords

import (
	"context"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtohotwords"
	"github.com/insmtx/corekg/apps/kecore/models/foresthotwords"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/services/svchotwords/agent"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// GetHotWords 获取热词
func GetHotWords(ctx *gin.Context, req *dtohotwords.GetHotWordsRequest) (res *dtohotwords.GetHotWordsResponse, err error) {
	res = &dtohotwords.GetHotWordsResponse{}
	dao := foresthotwords.NewForestHotWordDao()
	word, err := dao.GetByCond(ctx, &foresthotwords.ForestHotWordCond{
		BaseCond: foresthotwords.BaseCond{
			Uin:     runtime.Uin(ctx),
			OrderBy: []string{"created_at desc"},
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "GetHotWords GetByCond %v", err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "查询失败"
		return res, nil
	}
	if word != nil {
		res.Response.Words = word.HotWords
	}
	return res, nil
}

// GenerateUsersHotWords 生成用户热词
func GenerateUsersHotWords() (string, error) {
	ctx := context.Background()
	// 查询所有 uin
	var users []*accounttype.UserIdentification
	err := dbutil.Account().Table((&accounttype.UserIdentification{}).TableName()).
		WithContext(ctx).
		Find(&users).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[account] found failed, %s", err)
		return "", err
	}

	model, err := chatmodel.GetModelByID(ctx, 1)
	if err != nil {
		logs.ErrorContextf(ctx, "GetModelByID err:%v", err)
		return "", err
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  model.APIKey,
		Model:   model.ModelName,
		BaseURL: strings.TrimSuffix(model.ModelUrl, "/chat/completions"),
	})
	hotWordList := foresttype.ForestHotWordList{}
	// 获取所有 uin 所有问题
	for _, user := range users {
		if user.SubjectType == accounttype.SubjectTypeIndividual {
			continue
		}
		questions, err := chatquestion.ListUinSevenDayQuestions(ctx, user.ID)
		if err != nil {
			logs.ErrorContextf(ctx, "GenerateUsersHotWords ListUinSevenDayQuestions")
			continue
		}
		if len(questions) <= 0 {
			logs.WarnContextf(ctx, "no questions uin: %d", user.ID)
			continue
		}

		questionsList := []string{}

		for _, q := range questions {
			if q.Source.Question == "" {
				continue
			}
			questionsList = append(questionsList, q.Source.Question)
		}

		if len(questionsList) <= 0 {
			continue
		}
		// ai总结
		resWords, err := agent.ExecuteHotWordAgent(ctx, chatModel, questionsList)
		if err != nil {
			logs.ErrorContextf(ctx, "GenerateUsersHotWords ExecuteHotWordAgent %v", err)
			continue
		}
		if len(resWords) <= 1 {
			continue
		}
		hotWordList = append(hotWordList, foresttype.ForestHotWord{
			Uin:       user.ID,
			CompanyID: user.SubjectID,
			HotWords:  types.NewStringArray(resWords),
		})
	}
	logs.InfoContextf(ctx, "hotWordList count: %d", len(hotWordList))
	if len(hotWordList) <= 0 {
		logs.WarnContextf(ctx, "GenerateUsersHotWords no hot words")
		return "", nil
	}
	// 入库
	dao := foresthotwords.NewForestHotWordDao()
	err = dao.BatchInsert(ctx, hotWordList)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateUsersHotWords BatchInsert err: %v", err)
		return "", err
	}
	return "", nil
}
