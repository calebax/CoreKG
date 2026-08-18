package keqa

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/apps/kesearch/models/reranksearch"
	"github.com/ygpkg/yg-go/logs"
)

// HandelSearchReference 处理搜索引用
func HandelSearchReference(ctx context.Context, forestIDs, fileIDs []uint, esIndex string, question string) (*SearchReferenceWrapper, error) {
	searchWrapper, err := essearch.NewEsSearchWrapper(ctx, esIndex, question, forestIDs, fileIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "NewEsSearchWrapper error: %v", err)
		return nil, err
	}
	wrapper := &SearchReferenceWrapper{
		ctx:           ctx,
		ForestIDs:     forestIDs,
		FileIDs:       fileIDs,
		EsIndex:       esIndex,
		SearchWrapper: searchWrapper,
		Question:      question,
	}

	return wrapper, nil
}

type SearchReferenceWrapper struct {
	ctx context.Context
	// Forest   *foresttype.KnownowForest
	ForestIDs     []uint
	FileIDs       []uint
	EsIndex       string
	SearchWrapper *essearch.EsSearchWrapper

	Files    []*foresttype.KnownowForestFile
	FilesMap map[uint]*foresttype.KnownowForestFile
	Question string
}

// PreSearchQuestionChunk 预搜索问题
func (wrapper *SearchReferenceWrapper) PreSearchQuestionChunk() (*essearch.SearchResult, []*foresttype.KnownowForestFile, error) {

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
	files, err := forest.ListForestEnableFile(fileids)
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
func (wrapper *SearchReferenceWrapper) SupSearchQuestionChunk(preResult *essearch.SearchResult) (chattype.QueryReferenceList, error) {
	searchResult, err := wrapper.SearchWrapper.SearchChunkSequence(preResult)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "SearchChunkSequence error: %v", err)
		return nil, err
	}
	logs.InfoContextf(wrapper.ctx, "SearchChunkSequence result: %v", searchResult.Hits.Hits)

	refFileMap := map[uint]*chattype.QueryReference{}
	userMap := map[uint]*accounttype.User{}
	for _, hit := range searchResult.Hits.Hits {
		file, exists := wrapper.FilesMap[hit.Source.FileID]
		if !exists {
			logs.WarnContextf(wrapper.ctx, "FileID %d not found in fileMap", hit.Source.FileID)
			continue
		}
		userEntity, exists := userMap[file.Uin]
		if !exists {
			u, err := user.GetUserByUin(wrapper.ctx, file.Uin)
			if err != nil {
				logs.WarnContextf(wrapper.ctx, "GetUserByUin error: %v", err)
				continue
			}
			userEntity = u
			userMap[file.Uin] = userEntity
		}

		refChunk := chattype.QueryReferenceChunk{
			// ChunkID:  innerhit.ChunkID,
			Sequence: hit.Source.Sequence,
			Content:  strings.Join(hit.Source.TitleLevel, "\n") + "\n" + hit.Source.Description,
			ImageURL: hit.Source.ImageUrl,
		}
		ref, exists := refFileMap[hit.Source.FileID]
		if exists {
			ref.ChunkList = append(ref.ChunkList, refChunk)
		} else {
			ref := &chattype.QueryReference{
				AvatarURL:      userEntity.AvatarURL,
				UserName:       userEntity.Name,
				CreatedAt:      file.CreatedAt,
				Uin:            file.Uin,
				FileID:         hit.Source.FileID,
				ForestID:       hit.Source.ForestID,
				FileName:       file.Name,
				DataSourceType: chattype.DataSourceTypeDC,
				ChunkList:      chattype.QueryReferenceChunkList{refChunk},
			}
			refFileMap[hit.Source.FileID] = ref
		}
	}
	retList := chattype.QueryReferenceList{}
	for _, ref := range refFileMap {
		ref.ChunkList.DeduplicateByContent()
		sort.Sort(ref.ChunkList)
		retList = append(retList, ref)
	}
	sort.Sort(retList)
	return retList, nil
}

// SearchFileDescriptions 搜索总结性问题
func (wrapper *SearchReferenceWrapper) SearchFileDescriptions(question string) (*essearch.SearchResult, []*foresttype.KnownowForestFile, error) {
	searchResult, err := wrapper.SearchWrapper.DescriptionSearch()
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "SearchQuestionChunk error: %v", err)
		return nil, nil, err
	}
	logs.InfoContextf(wrapper.ctx, "SearchQuestionChunk result: %v", searchResult.Hits.Hits)

	fileids := []uint{}
	for _, hit := range searchResult.Hits.Hits {
		fileids = append(fileids, hit.Source.FileID)
	}
	files, err := forest.ListForestEnableFile(fileids)
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
func (wrapper *SearchReferenceWrapper) SupQuestionChunk(preResult *essearch.SearchResult) chattype.QueryReferenceList {
	refFileMap := map[uint]*chattype.QueryReference{}
	userMap := map[uint]*accounttype.User{}
	for _, hit := range preResult.Hits.Hits {
		file, exists := wrapper.FilesMap[hit.Source.FileID]
		if !exists {
			logs.WarnContextf(wrapper.ctx, "FileID %d not found in fileMap", hit.Source.FileID)
			continue
		}
		userEntity, exists := userMap[file.Uin]
		if !exists {
			u, err := user.GetUserByUin(wrapper.ctx, file.Uin)
			if err != nil {
				logs.WarnContextf(wrapper.ctx, "GetUserByUin error: %v", err)
				continue
			}
			userEntity = u
			userMap[file.Uin] = userEntity
		}
		refChunk := chattype.QueryReferenceChunk{
			// ChunkID:  innerhit.ChunkID,
			Type:     hit.Source.Type,
			Sequence: hit.Source.Sequence,
			Content:  hit.Source.Description,
			ImageURL: hit.Source.ImageUrl,
		}
		ref, exists := refFileMap[hit.Source.FileID]
		if exists {
			ref.ChunkList = append(ref.ChunkList, refChunk)
		} else {
			ref := &chattype.QueryReference{
				AvatarURL:      userEntity.AvatarURL,
				UserName:       userEntity.Name,
				CreatedAt:      file.CreatedAt,
				Uin:            file.Uin,
				FileID:         hit.Source.FileID,
				ForestID:       hit.Source.ForestID,
				FileName:       file.Name,
				DataSourceType: chattype.DataSourceTypeDC,
				ChunkList:      chattype.QueryReferenceChunkList{refChunk},
			}
			refFileMap[hit.Source.FileID] = ref
		}
	}
	retList := chattype.QueryReferenceList{}
	for _, ref := range refFileMap {
		sort.Sort(ref.ChunkList)
		retList = append(retList, ref)
	}
	sort.Sort(retList)
	return retList
}

func (wrapper *SearchReferenceWrapper) concurrentSearchWithWaitGroup() (*essearch.SearchResult, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var searchResult, titleRes *essearch.SearchResult
	var err1, err2 error

	// 启动第一个 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := wrapper.SearchWrapper.SearchQuestionChunk()
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
		result, err := wrapper.SearchWrapper.SearchChunkWithTitle()
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

func (wrapper *SearchReferenceWrapper) RerankSearchQuestionChunk(searchConf *reranksearch.SearchConfig) (chattype.QueryReferenceList, error) {
	w, err := reranksearch.NewRerankSearchWrapper(wrapper.ctx,
		wrapper.EsIndex,
		wrapper.Question, wrapper.ForestIDs, wrapper.FileIDs, searchConf, nil)
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "NewRerankSearchWrapper error: %v", err)
		return nil, err
	}
	res, err := w.RerankSearchChunk()
	if err != nil {
		logs.ErrorContextf(wrapper.ctx, "RerankSearchChunk error: %v", err)
		return nil, err
	}
	return res, nil
}
