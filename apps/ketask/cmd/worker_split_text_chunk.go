package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

func init() {
	rootCmd.AddCommand(splitTextChunkCmd())
}

func splitTextChunkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split_text_chunk",
		Short: "Split text into chunks for processing",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InfoContextf(cmd.Context(), "Starting Split text chunk worker with "+
				"worker ID: %s, task type: %s, base URL: %s, workersvr URL %s", workerID, taskType, baseURL, workerServerURL)
			ctx := lifecycle.Std().Context()
			loadConfig(ctx, configFile)

			for i := 0; i < routineSize; i++ {
				go splitTextChunkRoutine(ctx, i, esClient)
			}
			lifecycle.Std().WaitExit()
		},
	}
	withESFlags(cmd)
	return cmd
}

func splitTextChunkRoutine(ictx context.Context, idx int, escli *elasticsearch.Client) {
	ctx := logs.WithContextFields(ictx, "routine_idx", idx)
	logs.InfoContextf(ctx, "Split text chunk worker started with routine index: %d", idx)
	for {
		select {
		case <-lifecycle.Std().C():
			return
		default:
			//if err := checkURL(workerServerURL); err != nil {
			//	logs.WarnContextf(ctx, "Worker server URL(%s) check failed: %v", workerServerURL, err)
			//	time.Sleep(5 * time.Second) // Retry after a delay
			//	continue
			//}
			payload := &ragtask.TaskPayload{}
			taskId, err := GetPendingTask(ctx, payload)
			if err != nil {
				logs.ErrorContextf(ctx, "Failed to get pending task: %v", err)
				time.Sleep(5 * time.Second) // Retry after a delay
				continue
			}
			logs.InfoContextf(ctx, "Received task ID: %d, %v", taskId, payload)
			if taskId == 0 {
				time.Sleep(2 * time.Second)
				continue
			}
			logs.InfoContextf(ctx, "Received task ID: %d, %v", taskId, payload)
			newSplitTextChunkWorker(ctx, taskId, payload, escli).Run()
			time.Sleep(2 * time.Second)
		}
	}
}

type splitTextChunkWorker struct {
	ctx      context.Context
	taskID   uint
	payload  *ragtask.TaskPayload
	rootDir  string
	escli    *elasticsearch.Client
	nbCli    *nbgraph.NebulaCli
	chunkIDs []string

	content    string
	taskStatus task.TaskStatus
	taskErrMsg []string
	taskResult interface{}
}

const (
	splitByCharacter     = "。，.,;；"
	splitByCharacterOnly = false
	chunkTokenSize       = 500
)

func newSplitTextChunkWorker(ctx context.Context, taskID uint, payload *ragtask.TaskPayload, escli *elasticsearch.Client) *splitTextChunkWorker {
	taskKey := fmt.Sprintf("%v-%s", taskID, encryptor.UUID())
	return &splitTextChunkWorker{
		ctx:     ctx,
		taskID:  taskID,
		payload: payload,
		rootDir: filepath.Join(os.TempDir(), "yg_text2chunk", taskKey),
		escli:   escli,

		taskStatus: task.TaskStatusFail,
	}
}

func (w *splitTextChunkWorker) Run() (err error) {
	defer func() {
		if err != nil {
			w.taskErrMsg = append(w.taskErrMsg, err.Error())
			w.taskStatus = task.TaskStatusFail
		}
		if err := w.CallBack(); err != nil {
			logs.ErrorContextf(w.ctx, "Callback failed: %v", err)
		}
	}()

	//get md file's content from public url
	if err = w.PreRun(); err != nil {
		return err
	}

	//do request to algo-service and need algo-svc to store a direct path
	r, err := DoFormRequest(w.ctx, workerServerURL+"/split", &url.Values{
		"uin":        []string{strconv.FormatUint(uint64(w.payload.Uin), 10)},
		"company_id": []string{strconv.FormatUint(uint64(w.payload.CompanyID), 10)},
		"forest_id":  []string{strconv.FormatUint(uint64(w.payload.ForestID), 10)},
		"file_id":    []string{strconv.FormatUint(uint64(w.payload.FileID), 10)},
		"content":    []string{w.payload.FileURL},
		"es_index":   []string{w.payload.ESIndex},
		"file_ext":   []string{w.payload.FileExt},
		"llm_model":  []string{w.payload.LLM.ModelName},
		"llm_url":    []string{w.payload.LLM.BaseURL},
		"llm_key":    []string{w.payload.LLM.APIKEY},

		//optional
		//"split_by_character": []string{splitByCharacter},
		//"split_by_character_only": []string{"false"},
		//"chunk_token_size": []string{strconv.FormatUint(chunkTokenSize, 10)},
	})
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	//get chunk_ids by file_id
	chunkIDS, err := GetChunkIDsByFileID(w.escli, w.payload.ESIndex, w.payload.FileID)
	if err != nil {
		return err
	}
	if len(chunkIDS) == 0 {
		logs.WarnContextf(w.ctx, "chunkIDs is empty")
		w.taskStatus = task.TaskStatusSuccess
		return nil
	}
	logs.InfoContextf(w.ctx, "accept len of chunkIDs: %v", len(chunkIDS))
	w.chunkIDs = chunkIDS

	w.taskStatus = task.TaskStatusSuccess
	w.taskResult = r

	return nil
}

func (w *splitTextChunkWorker) PreRun() error {
	//*删除之前脏数据
	//*删除nebula数据
	cli := NewNBCli(&appCfg.Nebula)
	if err := DeleteFiles(context.TODO(), w.payload.ForestID, []uint{w.payload.FileID}, w.payload.ESIndex, cli); err != nil {
		logs.ErrorContextf(w.ctx, "Failed to delete nebula entities/relation: %v", err)
		err = nil
	}
	defer cli.Release()
	//*删除es数据
	err := DeleteFileReferencesFileChunk(w.ctx, w.escli, w.payload.ESIndex, []uint{w.payload.FileID})
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to delete file references: %v", err)
		return err
	}
	return err
}

func (w *splitTextChunkWorker) StoragePath() string {
	return filepath.Join(w.rootDir, "storage")
}

func (w *splitTextChunkWorker) Init() error {
	if err := os.MkdirAll(w.rootDir, 0755); err != nil {
		logs.ErrorContextf(w.ctx, "Failed to create root directory: %v", err)
		return err
	}
	logs.InfoContextf(w.ctx, "Initialized worker with root directory: %s", w.rootDir)

	if err := os.MkdirAll(w.StoragePath(), 0755); err != nil {
		logs.ErrorContextf(w.ctx, "Failed to create storage directory: %v", err)
		return err
	}

	return nil
}

func (w *splitTextChunkWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), w.taskResult)
}

func (w *splitTextChunkWorker) Close() error {
	if debug {
		return nil
	}
	if err := os.RemoveAll(w.rootDir); err != nil {
		logs.ErrorContextf(w.ctx, "Failed to remove root directory: %v", err)
		return err
	} else {
		logs.InfoContextf(w.ctx, "Removed root directory: %s", w.rootDir)
	}
	return nil
}

func (w *splitTextChunkWorker) OriginFilePath() (path string) {
	path = filepath.Join(w.rootDir, "origin.pdf")
	u, _ := url.Parse(w.payload.FileURL)
	if u == nil {
		return
	}
	path = filepath.Join(w.rootDir, "origin"+filepath.Ext(u.Path))
	return
}

func GetChunkIDsByFileID(es *elasticsearch.Client, index string, fileID uint) ([]string, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "references",
				"query": map[string]interface{}{
					"term": map[string]interface{}{
						"references.file_id": fileID,
					},
				},
			},
		},
		"_source": false,
	}

	queryJSON, _ := json.Marshal(query)

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewBuffer(queryJSON),
		Size:  esapi.IntPtr(1000), // 根据实际情况调整
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES错误: %s", res.String())
	}

	var response struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析失败: %w", err)
	}

	var docIDs []string
	for _, hit := range response.Hits.Hits {
		docIDs = append(docIDs, hit.ID)
	}

	return docIDs, nil

}

func DoFormRequest(ctx context.Context, serverUrl string, values *url.Values) (io.Reader, error) {
	encode := values.Encode()
	if len(encode) > 100 {
		encode = encode[:100]
	}
	logs.InfoContextf(ctx, "Send Form Request to %v with %v ...", serverUrl, encode)
	resp, err := http.PostForm(serverUrl, *values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "Received non-OK HTTP status: %s, body: %s", resp.Status, string(body))
		return nil, fmt.Errorf("received non-OK HTTP status: %s,, body: %s", resp.Status, string(body))
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(all), nil
}

// DeleteFileReferences 根据文件id删除对应的references
func DeleteFileReferences(ctx context.Context, cli *elasticsearch.Client, indexName string, fileIds []uint) error {
	sourceStr := `
      if (ctx._source.references != null) {
        ArrayList updated = new ArrayList();
        for (item in ctx._source.references) {
          if (!params.file_ids_to_remove.contains(item.file_id)) {
            updated.add(item);
          }
        }
        if (updated.isEmpty()) {
          ctx.op = 'delete'; 
        } else {
          ctx._source.references = updated; 
        }
      }
	`
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", esquery.BuildMap("must", []esquery.Map{
			esquery.BuildMap("term", esquery.BuildMap("type", "entity")),
			esquery.BuildMap("nested", esquery.BuildMap("path", "references", "query", esquery.BuildMap("terms", esquery.BuildMap("references.file_id", fileIds)))),
		}))).
		Set("script", esquery.BuildMap("source", sourceStr, "lang", "painless", "params", esquery.BuildMap("file_ids_to_remove", fileIds)))

	querybyte, err := query.BuildBytes()
	if err != nil {
		return err
	}
	logs.InfoContextf(ctx, "delete es query body: %s", string(querybyte))
	resp, err := cli.UpdateByQuery([]string{indexName}, cli.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)), cli.UpdateByQuery.WithContext(context.Background()))
	if err != nil {
		return err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es query failed: %s err: %s", resp.Status(), string(body))
	}
	return nil
}

// DeleteFileReferencesFileChunk 根据文件id删除对应的references及chunk
func DeleteFileReferencesFileChunk(ctx context.Context, cli *elasticsearch.Client, indexName string, fileIds []uint) error {
	sourceStr := `
      if (params.file_ids_to_remove.contains(ctx._source.file_id)) {
        ctx.op = 'delete';
      } else if (ctx._source.references != null) {
        ArrayList updated = new ArrayList();
        for (item in ctx._source.references) {
          if (!params.file_ids_to_remove.contains(item.file_id)) {
            updated.add(item);
          }
        }
        if (updated.isEmpty()) {
          ctx.op = 'delete';
        } else {
          ctx._source.references = updated;
        }
      }
	`
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", esquery.BuildMap("should", []esquery.Map{
			esquery.BuildMap("terms", esquery.BuildMap("file_id", fileIds)),
			esquery.BuildMap("nested",
				esquery.BuildMap("path", "references",
					"query", esquery.BuildMap("terms", esquery.BuildMap("references.file_id", fileIds)))),
		}))).
		Set("script",
			esquery.BuildMap(
				"source", sourceStr,
				"lang", "painless",
				"params",
				esquery.BuildMap("file_ids_to_remove", fileIds)))

	querybyte, err := query.BuildBytes()
	if err != nil {
		return err
	}
	logs.InfoContextf(ctx, "delete es query body: %s", string(querybyte))
	resp, err := cli.UpdateByQuery(
		[]string{indexName},
		cli.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		cli.UpdateByQuery.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es query failed: %s err: %v", resp.Status(), body)
	}
	return nil
}

// DeleteFiles will delete all reference data about a slice of files.
// This is the primary, more powerful function for handling batch deletions.
func DeleteFiles(ctx context.Context, forestID uint, fileIDs []uint, space string, cli *nbgraph.NebulaCli) error {
	var fileIDsToDelete []string
	for _, v := range fileIDs {
		fileIDsToDelete = append(fileIDsToDelete, strconv.FormatUint(uint64(v), 10))
	}
	if len(fileIDsToDelete) == 0 {
		logs.InfoContextf(ctx, "DeleteFiles: received an empty slice of file IDs, nothing to do.")
		return nil
	}

	// --- STAGE 1: READ ----
	// 捞取 forest 和 uin 下的所有相关节点
	lookupNql := fmt.Sprintf("USE %s; LOOKUP ON entities WHERE entities.forest_id == %d "+
		"YIELD id(vertex) AS vid, properties(vertex).file_id AS file_id",
		space, forestID)

	resp, err := cli.ExecuteAndCheck(lookupNql)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFiles: failed to lookup nodes: %v", err)
		return err
	}
	if resp.IsEmpty() {
		return nil
	}

	// --- STAGE 2: MODIFY ----

	// **将待删除的 file IDs 放入一个 set (map) 中**
	filesToDeleteSet := make(map[string]struct{}, len(fileIDsToDelete))
	for _, id := range fileIDsToDelete {
		filesToDeleteSet[id] = struct{}{}
	}

	colNames := resp.GetColNames()
	nameIndexMap := make(map[string]int, len(colNames))
	for i, name := range colNames {
		nameIndexMap[name] = i
	}
	vidIndex, _ := nameIndexMap["vid"]
	fileIDIndex, _ := nameIndexMap["file_id"]

	var nodesToDelete []string
	nodesToUpdate := make(map[string]string)

	for _, row := range resp.GetRows() {
		vid := string(row.Values[vidIndex].SVal)
		currentFileID := string(row.Values[fileIDIndex].SVal)

		currentIDs := strings.Split(currentFileID, "&&&")
		remainingIDs := make([]string, 0, len(currentIDs))

		// 筛选出不应被删除的ID
		for _, id := range currentIDs {
			if _, found := filesToDeleteSet[id]; !found {
				remainingIDs = append(remainingIDs, id)
			}
		}

		// 根据剩余ID的数量来决定操作
		if len(remainingIDs) == len(currentIDs) {
			// 如果没有任何ID被移除，说明此节点与本次删除无关，跳过
			continue
		}

		if len(remainingIDs) == 0 {
			// 如果所有关联的ID都被移除了，那么此节点需要被删除
			nodesToDelete = append(nodesToDelete, vid)
		} else {
			// 如果还有剩余的ID，那么此节点需要被更新
			nodesToUpdate[vid] = strings.Join(remainingIDs, "&&&")
		}
	}

	// --- STAGE 3: WRITE ----

	// Execute Deletions
	if len(nodesToDelete) > 0 {
		var quotedVidsBuilder strings.Builder
		for i, vid := range nodesToDelete {
			if i > 0 {
				quotedVidsBuilder.WriteString(", ")
			}
			fmt.Fprintf(&quotedVidsBuilder, "\"%s\"", escapeVidForNQL(vid))
		}

		deleteNql := fmt.Sprintf("USE %s; DELETE VERTEX %s WITH EDGE;", space, quotedVidsBuilder.String())
		if _, err := cli.ExecuteAndCheck(deleteNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to delete nodes: %v,nql: %v", err, deleteNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully deleted %d nodes.", len(nodesToDelete))
	}

	// Execute Updates
	if len(nodesToUpdate) > 0 {
		var multiStatementBuilder strings.Builder
		fmt.Fprintf(&multiStatementBuilder, "USE %s;", space)
		for vid, newFileID := range nodesToUpdate {
			fmt.Fprintf(&multiStatementBuilder, " UPDATE VERTEX \"%s\" SET entities.file_id = \"%s\";",
				escapeVidForNQL(vid), escapeVidForNQL(newFileID))
		}
		updateNql := multiStatementBuilder.String()
		if _, err := cli.ExecuteAndCheck(updateNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to update nodes with batch statements: %v ,nql: %v", err, updateNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully sent batch update for %d nodes.", len(nodesToUpdate))
	}

	return nil
}

// escapeVidForNQL prepares a string to be safely embedded as a VID in an nGQL query.
// It escapes special characters like backslashes, quotes, and newlines.
func escapeVidForNQL(vid string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\", // 1. 先将 \ 替换为 \\
		"\"", "\\\"", // 2. 再将 " 替换为 \"
		"\n", "\\n", // 3. 将换行符 LF 替换为 \n 两个字符
		"\r", "\\r", // 4. 将回车符 CR 替换为 \r 两个字符
		"\t", "\\t", // 5. 将制表符 Tab 替换为 \t 两个字符
	)
	return replacer.Replace(vid)
}

// DeleteFilesEntAndRel will delete all reference data about a slice of files which own entities/relationship type.
// This is the primary, more powerful function for handling batch deletions.
func DeleteFilesEntAndRel(ctx context.Context, forestID uint, fileIDs []uint, space string) error {
	var fileIDsToDelete []string
	for _, v := range fileIDs {
		fileIDsToDelete = append(fileIDsToDelete, strconv.FormatUint(uint64(v), 10))
	}
	if len(fileIDsToDelete) == 0 {
		logs.InfoContextf(ctx, "DeleteFiles: received an empty slice of file IDs, nothing to do.")
		return nil
	}

	cli, err := nbgraph.NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFiles: failed to get nebula client: %v", err)
		return err
	}
	defer cli.Release()

	// --- STAGE 1: READ ----
	// 捞取 forest 和 uin 下的所有相关节点
	lookupNql := fmt.Sprintf("USE %s; LOOKUP ON entities WHERE entities.forest_id == %d "+
		"AND entities.type IN [\"entity\"] "+
		"YIELD id(vertex) AS vid, properties(vertex).file_id AS file_id",
		space, forestID)

	resp, err := cli.ExecuteAndCheck(lookupNql)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFiles: failed to lookup nodes: %v", err)
		return err
	}
	if resp.IsEmpty() {
		return nil
	}

	// --- STAGE 2: MODIFY ----
	// **将待删除的 file IDs 放入一个 set (map) 中**
	filesToDeleteSet := make(map[string]struct{}, len(fileIDsToDelete))
	for _, id := range fileIDsToDelete {
		filesToDeleteSet[id] = struct{}{}
	}

	colNames := resp.GetColNames()
	nameIndexMap := make(map[string]int, len(colNames))
	for i, name := range colNames {
		nameIndexMap[name] = i
	}
	vidIndex, _ := nameIndexMap["vid"]
	fileIDIndex, _ := nameIndexMap["file_id"]

	var nodesToDelete []string
	nodesToUpdate := make(map[string]string)

	for _, row := range resp.GetRows() {
		vid := string(row.Values[vidIndex].SVal)
		currentFileID := string(row.Values[fileIDIndex].SVal)

		currentIDs := strings.Split(currentFileID, "&&&")
		remainingIDs := make([]string, 0, len(currentIDs))

		// 筛选出不应被删除的ID
		for _, id := range currentIDs {
			if _, found := filesToDeleteSet[id]; !found {
				remainingIDs = append(remainingIDs, id)
			}
		}

		// 根据剩余ID的数量来决定操作
		if len(remainingIDs) == len(currentIDs) {
			// 如果没有任何ID被移除，说明此节点与本次删除无关，跳过
			continue
		}

		if len(remainingIDs) == 0 {
			// 如果所有关联的ID都被移除了，那么此节点需要被删除
			nodesToDelete = append(nodesToDelete, vid)
		} else {
			// 如果还有剩余的ID，那么此节点需要被更新
			nodesToUpdate[vid] = strings.Join(remainingIDs, "&&&")
		}
	}

	// --- STAGE 3: WRITE ----

	// Execute Deletions
	if len(nodesToDelete) > 0 {
		var quotedVidsBuilder strings.Builder
		for i, vid := range nodesToDelete {
			if i > 0 {
				quotedVidsBuilder.WriteString(", ")
			}
			fmt.Fprintf(&quotedVidsBuilder, "\"%s\"", escapeVidForNQL(vid))
		}

		deleteNql := fmt.Sprintf("USE %s; DELETE VERTEX %s WITH EDGE;", space, quotedVidsBuilder.String())
		if _, err := cli.ExecuteAndCheck(deleteNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to delete nodes: %v,nql: %v", err, deleteNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully deleted %d nodes.", len(nodesToDelete))
	}

	// Execute Updates
	if len(nodesToUpdate) > 0 {
		var multiStatementBuilder strings.Builder
		fmt.Fprintf(&multiStatementBuilder, "USE %s;", space)
		for vid, newFileID := range nodesToUpdate {
			fmt.Fprintf(&multiStatementBuilder, " UPDATE VERTEX \"%s\" SET entities.file_id = \"%s\";",
				escapeVidForNQL(vid), escapeVidForNQL(newFileID))
		}
		updateNql := multiStatementBuilder.String()
		if _, err := cli.ExecuteAndCheck(updateNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to update nodes with batch statements: %v ,nql: %v", err, updateNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully sent batch update for %d nodes.", len(nodesToUpdate))
	}

	return nil
}
