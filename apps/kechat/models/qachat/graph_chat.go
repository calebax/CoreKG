package qachat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/keqa"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
)

func (w *ChatWapper) ForestGraphChat() error {
	history, err := GetForestChatHistory(w.ctx, w.session)
	if err != nil {
		return err
	}
	if len(w.question.Source.ImageUrlList) > 0 {
		// 解析多模态
		w.question.Source.ImageContent = fmt.Sprintf("\n用户上传了%v张图片，根据描述信息找出与问题相关的可用信息并用于对话中，每条图片的具体描述信息如下：\n", len(w.question.Source.ImageUrlList))
		for i, url := range w.question.Source.ImageUrlList {
			res, err := keqa.DoImageParseRequest(w.ctx, url)
			if err != nil {
				logs.ErrorContextf(w.ctx, "[ForestChat] Failed to parse image: %v", err)
			}
			w.question.Source.ImageContent += fmt.Sprintf("第%v张\n%v\n", i+1, res)
		}
	}
	// 写入searching
	keqa.WriteFlag(w.ctx, w.question.Source.ReqID, llmchat.FlagSearching)
	wrapper, err := keqa.HandelSearchReference(w.ctx, w.session.ForestIDList.Slice(), w.session.FileIDList.Slice(), w.session.EsIndex, w.question.Source.Question+w.question.Source.ImageContent)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to HandelSearchReference: %v", err)
		return err
	}
	// 查找问答对
	fqa, err := wrapper.SearchWrapper.FindFQAByQuestion()
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to FindFQAByQuestion: %v", err)
		return err
	}
	if len(fqa.Hits.Hits) != 0 {
		logs.InfoContextf(w.ctx, "[ForestChat] FindFQAByQuestion result: %v", len(fqa.Hits.Hits))
		keqa.WriteSearchQA(w.ctx, fqa.Hits.Hits[0].Source.QAAnswer, w.question.Source.ReqID)
		w.question.Source.Answer = fqa.Hits.Hits[0].Source.QAAnswer
		w.question.Source.Status = chattype.QuestionStatusAnswered
		return nil
	}
	// 开始意图识别，正常搜索还是总结性问题
	intention, err := keqa.IntentionRecognition(w.ctx, w.question)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to IntentionRecognition: %v", err)
		return err
	}
	forestChatWapper := keqa.NewForestWrapper(w.ctx, w.question, wrapper, history)
	desc := false
	var g errgroup.Group
	switch intention {
	case "C", "c":
		_, err = forestChatWapper.DescriptionChat(true)
		desc = true
	default:
		_, err = forestChatWapper.GraphRerankChat(&g, true)
	}
	if err != nil {
		logs.ErrorContextf(w.ctx, "[ForestChat] Failed to DefaultChat: %v", err)
		return err
	}
	res, err := forestChatWapper.PromptChat(w.model.ID, desc, w.session.PromptMode)
	if res != nil {
		w.question.Source.Answer = res.Content
		w.question.Source.Reasoning = res.Reasoning
		w.question.Source.ReasoningSeconds = res.ReasoningTime
		w.question.Source.CostSeconds = res.CostSeconds
		w.question.Source.OutToken = res.Usage.CompletionTokens
		w.question.Source.CacheHitToken = res.Usage.PromptCacheHitTokens
		w.question.Source.CacheMissToken = res.Usage.PromptCacheMissTokens
		w.question.Source.TotalTokens = res.Usage.TotalTokens
		w.question.Source.Status = chattype.QuestionStatusAnswered
	}
	if err := g.Wait(); err == nil && !desc {
		if len(*w.question.Source.QueryReferenceList) < 1 {
			return nil
		}
		refItem, err := parseReferences(w.question.Source.Answer)
		if err != nil {
			logs.ErrorContextf(w.ctx, "[ForestChat] Failed to parseReferences: %v", err)
			return err
		}
		refchunks := getChunkIDsByReferenceItems(*w.question.Source.QueryReferenceList, refItem)
		graphInfo, err := graph.GetForestGraph(w.ctx, (*w.question.Source.QueryReferenceList)[0].ForestID)
		if err != nil {
			logs.ErrorContextf(w.ctx, "ForestGraphChat.GetForestGraph err:%v")
			return err
		}
		chunkGraph, err := graph.SearchGraphWithChunkIDs(w.ctx, graphInfo, refchunks)
		if err != nil {
			logs.ErrorContextf(w.ctx, "ForestGraphChat.SearchGraphWithChunkIDs err:%v")
			return err
		}
		w.question.Source.GraphChatReference = chunkGraph
	}
	return err
}

// ReferenceItem 表示一个解析出的引用项
type ReferenceItem struct {
	FileID   uint
	Sequence int
}

// ParseReferences 解析文本中的 {Reference ...} 模式并去重
func parseReferences(text string) ([]ReferenceItem, error) {
	// 正则表达式匹配 {Reference ...} 内的内容
	re := regexp.MustCompile(`\{Reference\s+(.+?)\}`)
	matches := re.FindAllStringSubmatch(text, -1)
	// 使用 map 来去重，key 为 FileID:ChunkID 的组合
	uniqueItems := make(map[string]ReferenceItem)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		content := match[1]
		content = strings.TrimSpace(content)
		// 修复：使用更精确的正则表达式直接匹配 §数字[数字,数字,...] 模式
		itemRe := regexp.MustCompile(`§(\d+)\[([^\]]+)\]`)
		itemMatches := itemRe.FindAllStringSubmatch(content, -1)
		for _, itemMatch := range itemMatches {
			if len(itemMatch) == 3 {
				fileIDStr := itemMatch[1]
				chunkIDsStr := itemMatch[2]

				fileID, err := strconv.Atoi(fileIDStr)
				if err != nil {
					return nil, fmt.Errorf("invalid file ID '%s': %v", fileIDStr, err)
				}
				// 解析 [] 中的 chunk IDs
				chunkIDStrs := strings.Split(chunkIDsStr, ",")
				for _, chunkIDStr := range chunkIDStrs {
					chunkIDStr = strings.TrimSpace(chunkIDStr)
					if chunkIDStr == "" {
						continue
					}
					chunkID, err := strconv.Atoi(chunkIDStr)
					if err != nil {
						return nil, fmt.Errorf("invalid chunk ID '%s': %v", chunkIDStr, err)
					}

					// 创建唯一标识符
					key := fmt.Sprintf("%d:%d", fileID, chunkID)
					uniqueItems[key] = ReferenceItem{
						FileID:   uint(fileID),
						Sequence: chunkID,
					}
				}
			}
		}
	}
	// 将 map 中的值转换为切片
	results := make([]ReferenceItem, 0, len(uniqueItems))
	for _, item := range uniqueItems {
		results = append(results, item)
	}
	// 可选：对结果进行排序，使输出更有序
	// 按 FileID 排序，如果 FileID 相同则按 ChunkID 排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].FileID > results[j].FileID ||
				(results[i].FileID == results[j].FileID && results[i].Sequence > results[j].Sequence) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results, nil
}

// 根据 ReferenceItem 数组获取对应的 ChunkID
func getChunkIDsByReferenceItems(refList chattype.QueryReferenceList, refItems []ReferenceItem) []string {
	var result []string

	// 创建一个 map 来快速查找 QueryReference，以 FileID 为 key
	refMap := make(map[uint]*chattype.QueryReference)
	for _, ref := range refList {
		refMap[ref.FileID] = ref
	}

	// 遍历 ReferenceItem 数组
	for _, item := range refItems {
		// 通过 FileID 找到对应的 QueryReference
		ref, exists := refMap[item.FileID]
		if !exists {
			continue // 如果没有找到对应的 FileID，跳过
		}

		// 在 ChunkList 中查找匹配的 Sequence
		for _, chunk := range ref.ChunkList {
			if chunk.Sequence == item.Sequence {
				result = append(result, chunk.ChunkID)
				break // 找到匹配的就跳出内层循环
			}
		}
	}

	return result
}
