package devkeywords

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/insmtx/corekg/apps/kecore/models/forestkeywords"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/logs"
)

// ReplaceSynonymKeywords 替换同义词
func ReplaceSynonymKeywords(ctx context.Context, chatModel model.ToolCallingChatModel, CompanyID uint, text string) string {
	// 获取关键词
	var results []string
	ag := NewParticipleAgent(ctx, chatModel)
	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: ag,
	})
	iter := runner.Query(ctx, text)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			logs.ErrorContextf(ctx, "ReplaceSynonymKeywords iter.Next error: %v", event.Err)
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if m := event.Output.MessageOutput.Message; m != nil {
				if len(m.Content) > 0 {
					logs.InfoContextf(ctx, "ReplaceSynonymKeywords answer: %s", m.Content)
					results = append(results, strings.Split(m.Content, ",")...)
				}
			}
		}
	}
	logs.InfoContextf(ctx, "ReplaceSynonymKeywords results: %v", results)
	if len(results) == 0 {
		return text
	}
	// 获取所有的子词
	dao := forestkeywords.NewKeywordsDao()
	wordList, err := dao.GetListByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: CompanyID,
		},
		Words:         results,
		WordType:      foresttype.WordTypeSynonym,
		SubjectID:     -1,
		SubjectIDNot0: true,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "ReplaceSynonymKeywords GetListByCond fail, err: %v", err)
		return text
	}
	subjectIDs := make([]uint, 0, len(wordList))
	for _, v := range wordList {
		subjectIDs = append(subjectIDs, v.SubjectID)
	}
	logs.InfoContextf(ctx, "ReplaceSynonymKeywords subjectIDs: %v", subjectIDs)
	if len(subjectIDs) == 0 {
		return text
	}
	replaceMap := make(map[string]string)
	subjectList, err := dao.GetListByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			IDs: subjectIDs,
		},
		SubjectID: -1,
		// SubjectIDNot0: true,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "ReplaceSynonymKeywords GetListByCond fail, err: %v", err)
		return text
	}

	subjectListMap := subjectList.ToMap()
	for _, v := range wordList {
		if subject, ok := subjectListMap[v.SubjectID]; ok {
			replaceMap[v.Word] = subject.Word
		}
	}
	if len(replaceMap) == 0 {
		return text
	}
	logs.InfoContextf(ctx, "ReplaceSynonymKeywords subjectListMap: %v", replaceMap)
	retext, err := replaceByKeywords(text, replaceMap)
	if err != nil {
		logs.ErrorContextf(ctx, "ReplaceSynonymKeywords replaceByKeywords fail, err: %v", err)
		return text
	}

	return retext
}

func replaceByKeywords(text string, replaceMap map[string]string) (string, error) {
	keys := make([]string, 0, len(replaceMap))
	for k := range replaceMap {
		keys = append(keys, k)
	}

	// ⭐ 关键：长词优先，防止子串被先匹配
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	// 正则安全转义
	for i, k := range keys {
		keys[i] = regexp.QuoteMeta(k)
	}

	// (后倒车雷达|维修步骤)
	pattern := "(" + strings.Join(keys, "|") + ")"

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	result := re.ReplaceAllStringFunc(text, func(m string) string {
		if v, ok := replaceMap[m]; ok {
			return v
		}
		return m
	})

	return result, nil
}

// ReplaceMajorKeywords 替换专业名词
func ReplaceMajorKeywords(ctx context.Context, chatModel model.ToolCallingChatModel, CompanyID uint, text string) string {
	// 获取关键词
	var results []string
	ag := NewParticipleAgent(ctx, chatModel)
	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: ag,
	})
	iter := runner.Query(ctx, text)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			logs.ErrorContextf(ctx, "ReplaceMajorKeywords iter.Next error: %v", event.Err)
			continue
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if m := event.Output.MessageOutput.Message; m != nil {
				if len(m.Content) > 0 {
					logs.InfoContextf(ctx, "ReplaceMajorKeywords answer: %s", m.Content)
					results = append(results, strings.Split(m.Content, ",")...)
				}
			}
		}
	}
	logs.InfoContextf(ctx, "ReplaceMajorKeywords results: %v", results)
	if len(results) == 0 {
		return text
	}
	// 获取所有的子词
	dao := forestkeywords.NewKeywordsDao()
	wordList, err := dao.GetListByCond(ctx, &forestkeywords.KeywordsCond{
		BaseCond: forestkeywords.BaseCond{
			CompanyID: CompanyID,
		},
		Words:    results,
		WordType: foresttype.WordTypeMajor,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "ReplaceMajorKeywords GetListByCond fail, err: %v", err)
		return text
	}
	if len(wordList) == 0 {
		return text
	}

	replaceMap := make(map[string]string)
	for _, v := range wordList {
		replaceMap[v.Word] = v.Word + "(" + v.Description + ")"
	}
	logs.InfoContextf(ctx, "ReplaceMajorKeywords replaceMap: %v", replaceMap)

	retext, err := replaceByKeywords(text, replaceMap)
	if err != nil {
		logs.ErrorContextf(ctx, "ReplaceSynonymKeywords replaceByKeywords fail, err: %v", err)
		return text
	}

	return retext
}
