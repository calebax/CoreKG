package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestPrintOption(t *testing.T) {
	opt := ragtask.TaskPayload{
		CommonPayload: task.CommonPayload{
			TaskType: "test",
			Timeout:  1000,
		},
		FileID: 1,

		FileURL:    "test",
		SubjectID:  1,
		UploadPath: "test",
		CompanyID:  1,
		ForestID:   1,
		Uin:        1,
		Bucket:     "test",

		VideoExtractext: &ragtask.VideoExtractextOption{
			Threshold:            0.5,
			FrameIntervalSeconds: 1,
		},
		ES: &ragtask.ESIndexOption{
			IndexName: "test",
			Addr:      "test:8080",
			Username:  "test",
			Password:  "<PASSWORD>",
		},
		Graph: &ragtask.GraphDBOption{
			Mode:     "test",
			Name:     "test",
			Addr:     "test:8080",
			Username: "test",
			Password: "<PASSWORD>",
		},
		Embedding: &config.LLMModelConfig{
			APIKEY:    "test",
			BaseURL:   "test",
			ModelName: "test",
		},
		VLLM: &config.LLMModelConfig{
			APIKEY:    "test",
			BaseURL:   "test",
			ModelName: "test",
		},
		LLM: &config.LLMModelConfig{
			APIKEY:    "test",
			BaseURL:   "test",
			ModelName: "test",
		},
	}
	data, _ := json.Marshal(opt)
	t.Error(string(data))
}

// BatchResult 批次结果
type BatchResult struct {
	BatchIndex     int
	StartFileIndex int
	EndFileIndex   int
	ProcessedFiles []*foresttype.KnownowForestFile
	Success        bool
	Error          error
}

// ProcessResult 处理结果
type ProcessResult struct {
	TotalFiles       int
	SuccessCount     int
	FailedBatches    []BatchResult
	LastSuccessIndex int
}

const (
	apiUrlStr = "https://api.example.com/v3/chat.Agent/chat/completions"
	apiKeyStr = "xxx"
)

func InitLoad() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}

	//this func call will init global var esClient
	NewEsCli(config.ESConfig{
		Addresses:  []string{"http://example.com:53082"},
		Username:   "elastic",
		Password:   "CHANGE_ME_PASSWORD",
		MaxRetries: 3,
	})

	//init fs storager
	if err := fs.InitForestStorage(); err != nil {
		panic(err)
	}
	invalidIDsMux = sync.Mutex{}
}
func MigrateToEsWithBatch(ctx context.Context, batchSize, workerCount int) (*ProcessResult, error) {
	logs.InfoContextf(ctx, "Call migrate es api, "+
		"this api will migrate file's description include analysis(now it called abstract),"+
		" mindmap , short-description to elastic search")

	// get the ready file list
	var fileList []*foresttype.KnownowForestFile
	if err := dbutil.Knownow().
		Where("analysis_status = ?", foresttype.TaskStatusSuccess).
		Where("mindmap_status = ?", foresttype.TaskStatusSuccess).
		Find(&fileList).
		Error; err != nil {
		return nil, fmt.Errorf("query files error: %v", err)
	}

	logs.InfoContextf(ctx, "Found %d files to migrate with %d workers", len(fileList), workerCount)

	result := &ProcessResult{
		TotalFiles:       len(fileList),
		LastSuccessIndex: -1,
	}

	// 计算每个worker处理的批次范围
	totalBatches := (len(fileList) + batchSize - 1) / batchSize
	batchesPerWorker := (totalBatches + workerCount - 1) / workerCount

	// 结果收集
	type WorkerResult struct {
		WorkerID      int
		SuccessCount  int
		InvalidCount  int
		FailedBatches []BatchResult
		LastIndex     int
		Error         error
	}

	results := make(chan WorkerResult, workerCount)
	var wg sync.WaitGroup

	// 启动多个worker
	for workerID := 0; workerID < workerCount; workerID++ {
		wg.Add(1)
		go func(wid int) {
			defer wg.Done()

			// 计算当前worker的批次范围
			startBatch := wid * batchesPerWorker
			endBatch := (wid + 1) * batchesPerWorker
			if endBatch > totalBatches {
				endBatch = totalBatches
			}

			workerResult := WorkerResult{
				WorkerID:  wid,
				LastIndex: -1,
			}

			logs.InfoContextf(ctx, "Worker %d processing batches %d-%d", wid, startBatch, endBatch-1)

			// 处理分配给当前worker的批次
			for batchIndex := startBatch; batchIndex < endBatch; batchIndex++ {
				// 计算文件索引范围
				startFileIndex := batchIndex * batchSize
				endFileIndex := startFileIndex + batchSize
				if endFileIndex > len(fileList) {
					endFileIndex = len(fileList) // 防止越界
				}

				batch := fileList[startFileIndex:endFileIndex]

				logs.InfoContextf(ctx, "Worker %d processing batch %d: files %d-%d (%d files)",
					wid, batchIndex, startFileIndex, endFileIndex-1, len(batch))

				batchResult := BatchResult{
					BatchIndex:     batchIndex,
					StartFileIndex: startFileIndex,
					EndFileIndex:   endFileIndex - 1,
					ProcessedFiles: batch,
				}

				// 调用修改后的processBatch，获取准确的统计
				validCount, invalidCount, err := processBatch(ctx, batch, batchIndex)
				if err != nil {
					batchResult.Success = false
					batchResult.Error = err
					workerResult.FailedBatches = append(workerResult.FailedBatches, batchResult)

					err := fmt.Errorf("worker %d batch %d failed at file index %d-%d: %v",
						wid, batchIndex, startFileIndex, endFileIndex-1, err)
					logs.ErrorContextf(ctx, err.Error())
					workerResult.Error = err
					results <- workerResult
					return
				}

				batchResult.Success = true

				// 使用准确的统计数据
				workerResult.SuccessCount += validCount
				workerResult.InvalidCount += invalidCount
				workerResult.LastIndex = endFileIndex - 1

				logs.InfoContextf(ctx, "Worker %d batch %d completed: Valid=%d, Invalid=%d",
					wid, batchIndex, validCount, invalidCount)
			}

			logs.InfoContextf(ctx, "Worker %d completed: Valid=%d, Invalid=%d",
				wid, workerResult.SuccessCount, workerResult.InvalidCount)
			results <- workerResult
		}(workerID)
	}

	// 等待所有worker完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	totalValid := 0
	totalInvalid := 0
	maxLastIndex := -1

	for workerResult := range results {
		if workerResult.Error != nil {
			// 如果有worker失败，返回错误
			result.SuccessCount = totalValid
			result.LastSuccessIndex = maxLastIndex
			for _, failed := range workerResult.FailedBatches {
				result.FailedBatches = append(result.FailedBatches, failed)
			}
			return result, workerResult.Error
		}

		totalValid += workerResult.SuccessCount
		totalInvalid += workerResult.InvalidCount
		if workerResult.LastIndex > maxLastIndex {
			maxLastIndex = workerResult.LastIndex
		}
	}

	result.SuccessCount = totalValid
	result.LastSuccessIndex = maxLastIndex

	// 获取最终的无效文件总数（所有worker的汇总）
	finalInvalidCount := getInvalidIDsCount()

	logs.InfoContextf(ctx, "Migration completed: Valid=%d, Invalid=%d, Total=%d (%.1f%% valid)",
		totalValid, finalInvalidCount, len(fileList),
		float64(totalValid)/float64(len(fileList))*100)

	return result, nil
}
func processBatch(ctx context.Context, files []*foresttype.KnownowForestFile, batchIndex int) (int, int, error) {
	var buf bytes.Buffer
	var processedDocs []ProcessedDoc
	var batchInvalidIDs []uint // 当前批次的无效文件ID
	validCount := 0
	for fileIndex, f := range files {
		doc, docID, err := buildDocument(ctx, f)
		if err != nil {
			if errors.Is(err, ErrInValidFile) {
				batchInvalidIDs = append(batchInvalidIDs, f.ID)
				logs.InfoContextf(ctx, "Batch %d: Invalid file - FileID: %d", batchIndex, f.ID)
				continue
			}
			// 返回统计信息，即使出错也要记录已处理的无效文件
			addInvalidFileIDs(batchInvalidIDs)
			return validCount, len(batchInvalidIDs), fmt.Errorf("build document failed for file %v (index %d): %v", f.ID, fileIndex, err)
		}

		validCount++
		processedDocs = append(processedDocs, ProcessedDoc{
			FileID:    f.ID,
			DocID:     docID,
			FileIndex: fileIndex,
		})
		meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`, "ke_0", docID)
		buf.WriteString(meta + "\n")

		docBytes, err := json.Marshal(doc)
		if err != nil {
			addInvalidFileIDs(batchInvalidIDs)
			return validCount, len(batchInvalidIDs), fmt.Errorf("marshal document failed for file %v: %v", f.ID, err)
		}
		buf.Write(docBytes)
		buf.WriteString("\n")

		logs.InfoContextf(ctx, "Prepared document for file %d (batch %d, file %d)", f.ID, batchIndex, fileIndex)
	}

	invalidCount := len(batchInvalidIDs)

	// 添加无效文件ID到全局列表
	addInvalidFileIDs(batchInvalidIDs)

	// 记录批次统计
	logs.InfoContextf(ctx, "Batch %d processed: Valid=%d, Invalid=%d, Total=%d",
		batchIndex, validCount, invalidCount, len(files))

	if len(batchInvalidIDs) > 0 {
		logs.InfoContextf(ctx, "Batch %d invalid FileIDs: %v", batchIndex, batchInvalidIDs)
	}

	// 如果没有有效文件，跳过ES插入
	if validCount == 0 {
		logs.InfoContextf(ctx, "Batch %d: No valid files, skipping ES insert", batchIndex)
		return validCount, invalidCount, nil
	}
	// ES插入逻辑
	resp, err := esClient.Bulk(
		bytes.NewReader(buf.Bytes()),
		esClient.Bulk.WithContext(ctx),
		esClient.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return validCount, invalidCount, fmt.Errorf("bulk request failed for batch %d: %v", batchIndex, err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return validCount, invalidCount, fmt.Errorf("bulk response error for batch %d: %s", batchIndex, resp.Status())
	}

	if err := checkBulkResponse(ctx, resp.Body, processedDocs, batchIndex); err != nil {
		return validCount, invalidCount, err
	}

	logs.InfoContextf(ctx, "Batch %d ES insert successful: %d documents", batchIndex, validCount)
	return validCount, invalidCount, nil
}

type ProcessedDoc struct {
	FileID    uint
	DocID     string
	FileIndex int
}

func buildDocument(ctx context.Context, f *foresttype.KnownowForestFile) (*ragtypes.FileDescription, string, error) {
	var (
		absByte, mindByte []byte
		absStr            string
	)

	absByte, err := fs.GetAnalysisContent(f)
	if err != nil {
		logs.ErrorContextf(ctx, "GetAnalysisContent failed for file %d: %v", f.ID, err)
		return nil, "", ErrInValidFile
	}
	absStr = string(absByte)

	mindByte, err = fs.GetFileGraph(f)
	if err != nil {
		logs.ErrorContextf(ctx, "GetFileGraph failed for file %d: %v", f.ID, err)
		return nil, "", ErrInValidFile
	}

	desc, err := doAgentRequest(ctx, map[string]string{
		"input1": absStr,
	}, apiUrlStr, apiKeyStr, "TXV9iwl")
	if err != nil {
		logs.ErrorContextf(ctx, "doAgentRequest failed for file %d: %v", f.ID, err)
		return nil, "", fmt.Errorf("doAgentRequest failed: %v", err)
	}

	eb, err := essearch.GetEmbedding(desc)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmbedding failed for file %d: %v", f.ID, err)
		return nil, "", fmt.Errorf("GetEmbedding failed: %v", err)
	}
	now := time.Now()
	docID := uuid.NewString()
	doc := &ragtypes.FileDescription{
		Common: ragtypes.Common{
			CreatedAt: now,
			UpdatedAt: now,
			ForestID:  f.ForestID,
			CompanyID: f.CompanyID,
			Uin:       f.Uin,
			Type:      ragtypes.ChunkTypeFileDescription,
			Version:   "",
		},
		FileID:      f.ID,
		Description: desc,
		Embedding:   eb,
		MindMap:     string(mindByte),
		Abstract:    absStr,
	}
	logs.InfoContextf(ctx, "Successfully built document for file %d, docID: %s", f.ID, docID)
	return doc, docID, nil
}
func checkBulkResponse(ctx context.Context, body io.Reader, docs []ProcessedDoc, batchIndex int) error {
	var bulkResp struct {
		Errors bool                     `json:"errors"`
		Items  []map[string]interface{} `json:"items"`
	}
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		logs.ErrorContextf(ctx, "Read response body failed for batch %d: %v", batchIndex, err)
		return fmt.Errorf("read response body failed: %v", err)
	}
	if err := json.Unmarshal(bodyBytes, &bulkResp); err != nil {
		logs.ErrorContextf(ctx, "Parse response failed for batch %d: %v", batchIndex, err)
		return fmt.Errorf("parse response failed: %v", err)
	}
	if !bulkResp.Errors {
		logs.ErrorContextf(ctx, "Batch %d: all items indexed successfully", batchIndex)
		return nil
	}

	var failedItems []string
	for i, item := range bulkResp.Items {
		for action, details := range item {
			if det, ok := details.(map[string]interface{}); ok {
				if errInfo, exists := det["error"]; exists {
					docInfo := "unknown"
					if i < len(docs) {
						docInfo = fmt.Sprintf("FileID:%d, DocID:%s", docs[i].FileID, docs[i].DocID)
					}
					errorMsg := fmt.Sprintf("batch[%d] item[%d] %s %s: %v", batchIndex, i, action, docInfo, errInfo)
					failedItems = append(failedItems, errorMsg)
					logs.ErrorContextf(ctx, "Bulk error: %s", errorMsg)
				}
			}
		}
	}
	if len(failedItems) > 0 {
		return fmt.Errorf("bulk operation has errors: %s", strings.Join(failedItems, "; "))
	}
	return nil
}

var (
	InvalidIDs     []uint
	invalidIDsMux  sync.Mutex
	ErrInValidFile = errors.New("invalid file")
)

// 线程安全添加无效文件ID
func addInvalidFileIDs(ids []uint) {
	if len(ids) == 0 {
		return
	}
	invalidIDsMux.Lock()
	defer invalidIDsMux.Unlock()
	InvalidIDs = append(InvalidIDs, ids...)
}

// 线程安全获取无效文件数量
func getInvalidIDsCount() int {
	invalidIDsMux.Lock()
	defer invalidIDsMux.Unlock()
	return len(InvalidIDs)
}

// 线程安全获取无效文件列表副本
func getInvalidIDsCopy() []uint {
	invalidIDsMux.Lock()
	defer invalidIDsMux.Unlock()
	result := make([]uint, len(InvalidIDs))
	copy(result, InvalidIDs)
	return result
}

func TestMigrateToEs(t *testing.T) {
	InitLoad()
	ctx := context.Background()

	invalidIDsMux.Lock()
	InvalidIDs = InvalidIDs[:0]
	invalidIDsMux.Unlock()

	result, err := MigrateToEsWithBatch(ctx, 50, 5)
	if err != nil {
		t.Logf("Migration failed: %v", err)
		if result != nil {
			t.Logf("Last successful index: %d", result.LastSuccessIndex)
			t.Logf("Success count: %d/%d", result.SuccessCount, result.TotalFiles)
			for _, failed := range result.FailedBatches {
				t.Logf("Failed batch %d (files %d-%d): %v",
					failed.BatchIndex, failed.StartFileIndex, failed.EndFileIndex, failed.Error)
			}
		}
		t.Fatal(err)
	}

	invalidFileIDs := getInvalidIDsCopy()

	t.Logf("Migration completed successfully: Valid=%d, Invalid=%d, Total=%d",
		result.SuccessCount, len(invalidFileIDs), result.TotalFiles)

	if len(invalidFileIDs) > 0 {
		t.Logf("Invalid file IDs (%d total): %v", len(invalidFileIDs), invalidFileIDs)
	}

	expectedTotal := result.SuccessCount + len(invalidFileIDs)
	if expectedTotal != result.TotalFiles {
		t.Errorf("Statistics mismatch: Valid(%d) + Invalid(%d) = %d, but Total is %d",
			result.SuccessCount, len(invalidFileIDs), expectedTotal, result.TotalFiles)
	} else {
		t.Logf("✅ Statistics verification passed")
	}
}

//=============================================================================================

const (
	AgentAbstract       = "NHfdA8m"
	AgentChunkAnalysis  = "f3sKLEa"
	AgentAbsChunkMerge  = "KLgmgsX"
	AgentMindChunkMerge = "HHFTCNR"

	ChunkSize    = 1000
	MaxWorkers   = 50
	MaxTokenSize = 10 * (20 << 10)
)

//func TestChunkAnalysis(t *testing.T) {
//	ctx := context.TODO()
//	var content []byte
//
//	for i := 0; i < 1; i++ {
//		f, err := os.ReadFile("/home/zoe/Downloads/civil_code.md")
//		if err != nil {
//			t.Fatal(err)
//		}
//		content = append(content, f...)
//	}
//
//	fStr := string(content)
//	loadConfig(ctx, "/home/zoe/CodeSpace/roc/clients/task_worker/config/test.yaml")
//
//	logs.InfoContextf(ctx, "content length %v", len(fStr))
//	abs, mind, desc, eb, err := wa.Do(fStr)
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("===========================================DONE===================================================")
//	fmt.Println(abs)
//	fmt.Println(mind)
//	fmt.Println(desc)
//	fmt.Println(len(eb))
//}

func TestDoAgentReqAbs(t *testing.T) {
	ctx := context.TODO()
	var content []byte

	for i := 0; i < 1; i++ {
		f, err := os.ReadFile("/home/zoe/Downloads/civil_code.md")
		if err != nil {
			t.Fatal(err)
		}
		content = append(content, f...)
	}

	fStr := string(content)
	loadConfig(ctx, "/home/zoe/CodeSpace/roc/clients/task_worker/config/test.yaml")

	logs.InfoContextf(ctx, "content length %v", len(fStr))
	resp, err := doAgentRequest(ctx, map[string]string{
		"input1": fStr,
	}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDAbstract],
		appCfg.Agent.Pool[MDAbsChunk], appCfg.Agent.Pool[MDMergeAbstract])
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}

func TestDoAgentReqMind(t *testing.T) {
	ctx := context.TODO()
	var content []byte

	for i := 0; i < 1; i++ {
		f, err := os.ReadFile("/home/zoe/Downloads/civil_code.md")
		if err != nil {
			t.Fatal(err)
		}
		content = append(content, f...)
	}

	fStr := string(content)

	loadConfig(ctx, "/home/zoe/CodeSpace/roc/clients/task_worker/config/test.yaml")

	titles := ExtractMarkdownTitles(fStr)
	c := strings.Join(titles, "\n")
	c += c
	c += c
	c += c
	c += c
	c += c
	c += c

	logs.InfoContextf(ctx, "content length %v", len(c))
	resp, err := doAgentRequest(ctx, map[string]string{
		"input1": c,
	}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDMindmap],
		appCfg.Agent.Pool[MDMindChunk], appCfg.Agent.Pool[MDMergeMindmap])
	if err != nil {
		t.Fatal(err)
	}

	//extract json code block
	jsonCode := ExtractCode("json", resp)

	//process resp with embedded uuid
	embeddedUuid, err := ProcessEmbeddedUuid(jsonCode)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp)
	fmt.Println(embeddedUuid)
}
