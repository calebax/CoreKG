package keqa

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

// HandelSearchReference 处理搜索引用
func HandelSearchReference(ctx *gin.Context, forestIDs, fileIDs []uint, esIndex string, question string) (*searchReferenceWrapper, error) {
	searchWrapper, err := essearch.NewEsSearchWrapper(ctx, esIndex, question, forestIDs, fileIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "NewEsSearchWrapper error: %v", err)
		return nil, err
	}
	wrapper := &searchReferenceWrapper{
		ctx:           ctx,
		ForestIDs:     forestIDs,
		FileIDs:       fileIDs,
		EsIndex:       esIndex,
		searchWrapper: searchWrapper,
	}

	return wrapper, nil
}

type searchReferenceWrapper struct {
	ctx *gin.Context
	// Forest   *foresttype.KnownowForest
	ForestIDs     []uint
	FileIDs       []uint
	EsIndex       string
	searchWrapper *essearch.EsSearchWrapper

	Files    []*foresttype.KnownowForestFile
	FilesMap map[uint]*foresttype.KnownowForestFile
}

// PreSearchQuestionChunk 预搜索问题
func (wrapper *searchReferenceWrapper) PreSearchQuestionChunk() (*essearch.SearchResult, []*foresttype.KnownowForestFile, error) {

	searchResult, err := wrapper.concurrentSearchWithWaitGroup()
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "SearchQuestionChunk error: %v", err)
		return nil, nil, err
	}
	logs.InfoContextf(wrapper.ctx, "SearchQuestionChunk result: %v", searchResult.Hits.Hits)

	fileids := []uint{}
	for _, hit := range searchResult.Hits.Hits {
		fileids = append(fileids, hit.Source.FileID)
	}
	files, err := forest.ListForestFile(fileids)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ListForestFile error: %v", err)
		return searchResult, nil, err
	}
	wrapper.Files = files
	wrapper.FilesMap = map[uint]*foresttype.KnownowForestFile{}
	for _, file := range files {
		wrapper.FilesMap[file.ID] = file
	}
	return searchResult, files, nil
}

// SupSearchQuestionChunk 补充搜索问题
func (wrapper *searchReferenceWrapper) SupSearchQuestionChunk(preResult *essearch.SearchResult) (foresttype.ChatReferenceList, error) {
	searchResult, err := wrapper.searchWrapper.SearchChunkSequence(preResult)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "SearchChunkSequence error: %v", err)
		return nil, err
	}
	logs.InfoContextf(wrapper.ctx, "SearchChunkSequence result: %v", searchResult.Hits.Hits)

	refFileMap := map[uint]*foresttype.ChatReference{}
	for _, hit := range searchResult.Hits.Hits {
		file, exists := wrapper.FilesMap[hit.Source.FileID]
		if !exists {
			logs.WarnContextf(wrapper.ctx, "FileID %d not found in fileMap", hit.Source.FileID)
			continue
		}

		refChunk := foresttype.ChatReferenceChunk{
			// ChunkID:  innerhit.ChunkID,
			Sequence: hit.Source.Sequence,
			Content:  strings.Join(hit.Source.TitleLevel, "\n") + "\n" + hit.Source.Description,
			ImageURL: hit.Source.ImageUrl,
		}
		ref, exists := refFileMap[hit.Source.FileID]
		if exists {
			ref.ChunkList = append(ref.ChunkList, refChunk)
		} else {
			ref := &foresttype.ChatReference{
				FileID:         hit.Source.FileID,
				ForestID:       hit.Source.ForestID,
				Filename:       file.Name,
				DataSourceType: foresttype.DataSourceTypeDC,
				ChunkList:      foresttype.ChatReferenceChunkList{refChunk},
			}
			refFileMap[hit.Source.FileID] = ref
		}
	}
	retList := foresttype.ChatReferenceList{}
	for _, ref := range refFileMap {
		ref.ChunkList.DeduplicateByContent()
		sort.Sort(ref.ChunkList)
		retList = append(retList, ref)
	}
	sort.Sort(retList)
	return retList, nil
}

// SearchFileDescriptions 搜索总结性问题
func (wrapper *searchReferenceWrapper) SearchFileDescriptions(question string) (*essearch.SearchResult, []*foresttype.KnownowForestFile, error) {
	searchResult, err := wrapper.searchWrapper.DescriptionSearch()
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "SearchQuestionChunk error: %v", err)
		return nil, nil, err
	}
	logs.InfoContextf(wrapper.ctx, "SearchQuestionChunk result: %v", searchResult.Hits.Hits)

	fileids := []uint{}
	for _, hit := range searchResult.Hits.Hits {
		fileids = append(fileids, hit.Source.FileID)
	}
	files, err := forest.ListForestFile(fileids)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "ListForestFile error: %v", err)
		return searchResult, nil, err
	}
	wrapper.Files = files
	wrapper.FilesMap = map[uint]*foresttype.KnownowForestFile{}
	for _, file := range files {
		wrapper.FilesMap[file.ID] = file
	}
	return searchResult, files, nil
}

// SupQuestionChunk 转换文件形式
func (wrapper *searchReferenceWrapper) SupQuestionChunk(preResult *essearch.SearchResult) foresttype.ChatReferenceList {
	refFileMap := map[uint]*foresttype.ChatReference{}
	for _, hit := range preResult.Hits.Hits {
		file, exists := wrapper.FilesMap[hit.Source.FileID]
		if !exists {
			logs.WarnContextf(wrapper.ctx, "FileID %d not found in fileMap", hit.Source.FileID)
			continue
		}

		refChunk := foresttype.ChatReferenceChunk{
			// ChunkID:  innerhit.ChunkID,
			Sequence: hit.Source.Sequence,
			Content:  hit.Source.Description,
			ImageURL: hit.Source.ImageUrl,
		}
		ref, exists := refFileMap[hit.Source.FileID]
		if exists {
			ref.ChunkList = append(ref.ChunkList, refChunk)
		} else {
			ref := &foresttype.ChatReference{
				FileID:         hit.Source.FileID,
				ForestID:       hit.Source.ForestID,
				Filename:       file.Name,
				DataSourceType: foresttype.DataSourceTypeDC,
				ChunkList:      foresttype.ChatReferenceChunkList{refChunk},
			}
			refFileMap[hit.Source.FileID] = ref
		}
	}
	retList := foresttype.ChatReferenceList{}
	for _, ref := range refFileMap {
		sort.Sort(ref.ChunkList)
		retList = append(retList, ref)
	}
	sort.Sort(retList)
	return retList
}

func (wrapper *searchReferenceWrapper) concurrentSearchWithWaitGroup() (*essearch.SearchResult, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var searchResult, titleRes *essearch.SearchResult
	var err1, err2 error

	// 启动第一个 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := wrapper.searchWrapper.SearchQuestionChunk()
		mu.Lock()
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "SearchQuestionChunk error: %v", err)
			err1 = err
		} else {
			searchResult = result
			logs.InfoContextf(wrapper.ctx, "SearchQuestionChunk result: %v", searchResult.Hits.Hits)
		}
		mu.Unlock()
	}()

	// 启动第二个 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := wrapper.searchWrapper.SearchChunkWithTitle()
		mu.Lock()
		if err != nil {
			logs.ErrorContextf(wrapper.ctx, "SearchChunkWithTitle error: %v", err)
			err2 = err
		} else {
			titleRes = result
		}
		mu.Unlock()
	}()

	// 等待所有 goroutine 完成
	wg.Wait()

	// 检查错误
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("search error: question=%v, title=%v", err1, err2)
	}
	searchResult.Hits.Hits = append(searchResult.Hits.Hits, titleRes.Hits.Hits...)
	return searchResult, nil
}
