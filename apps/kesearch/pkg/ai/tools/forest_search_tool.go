package tools

import (
	"context"
	"encoding/json"
	"sort"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/logs"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kesearch/services/svcessearch"
)

type ForestSearchToolConfig struct {
	ToolName string `json:"tool_name"`
	ToolDesc string `json:"tool_desc"`

	Ctx       *gin.Context `json:"-"`
	EsIndex   string       `json:"es_index"`
	ForestIDs []uint       `json:"forest_ids"`
	FileIDs   []uint       `json:"file_ids"`
	// OriginalQuestion preserves the user's question before the ReAct model rewrites tool arguments.
	OriginalQuestion string `json:"original_question,omitempty"`

	ReferencesResult *SharedReferences `json:"references,omitempty"`
}

type SearchRequest struct {
	Question       string `json:"question" jsonschema:"required,description=用户提出的问题或需要检索问答的内容."`
	SearchStrategy string `json:"search_strategy" jsonschema:"required,description=用于处理查询的检索策略：common_questions-普通的提问或输入内容；knowledge_summary-用户明确要求对整个知识库进行信息归纳、整理、梳理、汇总或生成概览性报告时使用，关注从全局视角对知识内容进行结构化或摘要,enum=common_questions,enum=knowledge_summary"`
}

type SearchResult struct {
	ResultType string `json:"result_type" jsonschema:"description=检索结果类型,enum=search_qa_result,enum=search_normal_result,enum=knowledge_summary"`
	ResultData any    `json:"result_data" jsonschema:"description=检索结果数据"`
}

type forestSearchTool struct {
	conf *ForestSearchToolConfig
}

// 知识库检索 检索QA以及文档问答
func NewForestSearchTool(ctx context.Context, conf *ForestSearchToolConfig) (tool.InvokableTool, error) {
	if conf == nil {
		conf = &ForestSearchToolConfig{}
	}
	toolName := conf.ToolName
	toolDesc := conf.ToolDesc
	if toolName == "" {
		toolName = "forest_search_tool"
	}
	if toolDesc == "" {
		toolDesc = `用于检索内部文档及知识库内容的专业工具，提供最高事实可信度的信息检索服务。

本工具是默认首选工具，应优先于直接查看文件内容。

这是最高效的知识库文档检索方式，可快速检索已经完成解析、索引后的知识库文件，相比直接查看原始文件内容，能够以更少的调用次数、更低的上下文消耗，更快定位到目标信息。

本工具适用于：
- 在文件、文档、知识库中快速定位相关内容
- 根据关键词、主题、标题、术语进行搜索
- 查找可能存在于多个文件中的相关信息
- 在尚未确定具体文件或内容位置时进行初步定位
- 先缩小范围，再决定是否需要调用内容查看工具查看详细内容

强制使用原则：
- 只要目标信息的位置、文件、章节或范围尚不明确，必须优先使用本工具检索
- 不允许在没有检索定位的情况下直接查看文件内容
- 只有在检索结果已经明确指出具体文件和内容范围后，才允许调用内容查看工具
- 若检索结果已经足以回答问题，则不应继续调用内容查看工具
- 无论文件大小，优先先检索、再查看具体内容，通常比直接查看更高效
- 应尽量避免无目的地直接查看文件内容，以减少无效调用和上下文消耗
`
	}

	fst := &forestSearchTool{conf: conf}
	tl, err := toolutils.InferTool(toolName, toolDesc, fst.invoke)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (t *forestSearchTool) invoke(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] ForestSearchTool.invoke start: question=%s, original_question=%s, search_strategy=%s, forest_ids=%v, file_ids=%v, es_index=%s",
		req.Question, t.conf.OriginalQuestion, req.SearchStrategy, t.conf.ForestIDs, t.conf.FileIDs, t.conf.EsIndex)
	if len(t.conf.ForestIDs) == 0 && len(t.conf.FileIDs) == 0 {
		logs.WarnContextf(ctx, "[DEBUG][chunk-empty] ForestSearchTool.invoke: ForestIDs and FileIDs are both empty, returning empty chunk")
		return &SearchResult{
			ResultType: "search_normal_result",
			ResultData: chattype.QueryReferenceList{},
		}, nil
	}
	// 查找问答对
	searchQaResult, err := svcessearch.FindFQAByQuestion(ctx, t.conf.EsIndex, req.Question, t.conf.ForestIDs, t.conf.FileIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "[NewForestSearchTool] FindFQAByQuestion error: %v", err)
		return nil, err
	}
	if len(searchQaResult.Hits.Hits) != 0 {
		logs.InfoContextf(ctx, "[NewForestSearchTool] FindFQAByQuestion result: %v", len(searchQaResult.Hits.Hits))
		// TODO keqa.WriteSearchQA
		return &SearchResult{
			ResultType: "search_qa_result",
			ResultData: searchQaResult.Hits.Hits[0].Source.QAAnswer,
		}, nil
	}

	if req.SearchStrategy == "knowledge_summary" {
		// 知识库摘要
		searchResultStr, searchRes, err := svcessearch.SearchDescription(t.conf.Ctx, t.conf.EsIndex, req.Question, t.conf.ForestIDs, t.conf.FileIDs)
		if err != nil {
			logs.ErrorContextf(ctx, "[NewForestSearchTool] SearchDescription error: %v", err)
			return nil, err
		}

		// 若文件级描述/摘要检索不到任何内容（例如该文件未生成 file_description 文档），
		// 回退到普通 chunk 检索，用已入库的正文片段继续回答，避免直接返回空结果。
		if len(searchRes) != 0 && searchResultStr != "" && searchResultStr != "[]" {
			searchResult := SearchResult{
				ResultType: "knowledge_summary",
				ResultData: searchResultStr,
			}
			t.conf.ReferencesResult.Append(searchRes...)

			return &searchResult, nil
		}
		logs.InfoContextf(ctx, "[NewForestSearchTool] knowledge_summary empty (file_description missing), fallback to common_questions for file_ids=%v", t.conf.FileIDs)
	}

	// 正常检索
	// TODO 传入检索内容
	searchRes, err := svcessearch.RerankSearchQuestionChunk(ctx, t.conf.EsIndex, req.Question, t.conf.ForestIDs, t.conf.FileIDs, nil, t.conf.OriginalQuestion)
	if err != nil {
		logs.ErrorContextf(ctx, "[NewForestSearchTool] RerankSearchQuestionChunk error: %v", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] ForestSearchTool.invoke RerankSearchQuestionChunk result: chunk_count=%d", len(searchRes))

	searchResult := SearchResult{
		ResultType: "search_normal_result",
		ResultData: searchRes,
	}
	t.conf.ReferencesResult.Append(searchRes...)

	return &searchResult, nil
}

type SharedReferences struct {
	mu   sync.RWMutex
	Refs []*chattype.QueryReference
}

func (s *SharedReferences) Append(refs ...*chattype.QueryReference) {
	if len(refs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.Refs = append(s.Refs, refs...)
}

func (s *SharedReferences) Get() []*chattype.QueryReference {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Refs
}

func (s *SharedReferences) GetAggregated() []*chattype.QueryReference {
	refs := s.Get()
	if len(refs) == 0 {
		return refs
	}

	aggregatedMap := make(map[uint]*chattype.QueryReference)
	for _, ref := range refs {
		if existing, ok := aggregatedMap[ref.FileID]; ok {
			existing.ChunkList = append(existing.ChunkList, ref.ChunkList...)
		} else {
			newRef := *ref

			newChunkList := make(chattype.QueryReferenceChunkList, len(ref.ChunkList))
			copy(newChunkList, ref.ChunkList)
			newRef.ChunkList = newChunkList
			aggregatedMap[ref.FileID] = &newRef
		}
	}

	result := make([]*chattype.QueryReference, 0, len(aggregatedMap))
	for _, ref := range aggregatedMap {
		// ChunkList根据ChunkID去重
		uniqueChunks := make(map[string]struct{})
		deduplicatedList := make(chattype.QueryReferenceChunkList, 0, len(ref.ChunkList))
		for _, chunk := range ref.ChunkList {
			if _, exists := uniqueChunks[chunk.ChunkID]; !exists {
				uniqueChunks[chunk.ChunkID] = struct{}{}
				deduplicatedList = append(deduplicatedList, chunk)
			}
		}
		ref.ChunkList = deduplicatedList

		// 根据Sequence从小到大进行排序
		sort.Sort(ref.ChunkList)

		result = append(result, ref)
	}

	return result
}

func HasNonEmptyResult(searchResult string) bool {

	result := SearchResult{}
	err := json.Unmarshal([]byte(searchResult), &result)
	if err != nil {
		return false
	}
	v := result.ResultData

	if v == nil {
		return false
	}

	switch val := v.(type) {

	// 空字符串
	case string:
		return val != ""

	// JSON 数组 → []any
	case []any:
		return len(val) > 0

	// JSON 对象 → map
	case map[string]any:
		return len(val) > 0

	default:
		return true
	}
}
