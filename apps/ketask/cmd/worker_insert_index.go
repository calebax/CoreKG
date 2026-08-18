package main

import (
	"context"
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
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

func init() {
	rootCmd.AddCommand(InsertIndexCmd())
}

func InsertIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insert_index",
		Short: "insert es index and graph db  for processing",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InfoContextf(cmd.Context(), "Starting insert index graph worker with "+
				"worker ID: %s, task type: %s, base URL: %s, workersvr URL %s",
				workerID, taskType, baseURL, workerServerURL)
			ctx := lifecycle.Std().Context()
			loadConfig(ctx, configFile)

			for i := 0; i < routineSize; i++ {
				go insertIndexRoutine(ctx, i, esClient)
			}
			lifecycle.Std().WaitExit()
		},
	}
	withESFlags(cmd)
	return cmd
}

func insertIndexRoutine(ictx context.Context, idx int, escli *elasticsearch.Client) {
	ctx := logs.WithContextFields(ictx, "routine_idx", idx)
	logs.InfoContextf(ctx, "insert index graph worker started with routine index: %d", idx)
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
			newInsertIndexWorker(ctx, taskId, payload, escli).Run()
			time.Sleep(2 * time.Second)
		}
	}
}

type insertIndexWorker struct {
	ctx     context.Context
	taskID  uint
	payload *ragtask.TaskPayload
	rootDir string
	escli   *elasticsearch.Client
	// nbcli    *nbgraph.NebulaCli
	chunkIDs []string

	content    string
	taskStatus task.TaskStatus
	taskErrMsg []string
	taskResult interface{}
}

func newInsertIndexWorker(ctx context.Context, taskID uint, payload *ragtask.TaskPayload, escli *elasticsearch.Client) *insertIndexWorker {
	taskKey := fmt.Sprintf("%v-%s", taskID, encryptor.UUID())
	return &insertIndexWorker{
		ctx:     ctx,
		taskID:  taskID,
		payload: payload,
		rootDir: filepath.Join(os.TempDir(), "yg_insertIndex", taskKey),
		escli:   escli,

		taskStatus: task.TaskStatusFail,
	}
}

func (w *insertIndexWorker) Run() (err error) {
	defer func() {
		if err != nil {
			w.taskErrMsg = append(w.taskErrMsg, err.Error())
			w.taskStatus = task.TaskStatusFail
		}
		if err := w.CallBack(); err != nil {
			logs.ErrorContextf(w.ctx, "Callback failed: %v", err)
		}
	}()

	if err = w.PreRun(); err != nil {
		return
	}

	//request to insert graphdb and es
	_, err = DoFormRequest(w.ctx, workerServerURL+"/index", &url.Values{
		"uin":        []string{strconv.FormatUint(uint64(w.payload.Uin), 10)},
		"file_id":    []string{strconv.FormatUint(uint64(w.payload.FileID), 10)},
		"company_id": []string{strconv.FormatUint(uint64(w.payload.CompanyID), 10)},
		"forest_id":  []string{strconv.FormatUint(uint64(w.payload.ForestID), 10)},
		"es_index":   []string{w.payload.ESIndex},
		"llm_model":  []string{w.payload.LLM.ModelName},
		"llm_url":    []string{w.payload.LLM.BaseURL},
		"llm_key":    []string{w.payload.LLM.APIKEY},
	})
	if err != nil {
		return
	}

	w.taskStatus = task.TaskStatusSuccess
	return
}

func (w *insertIndexWorker) PreRun() error {
	resp, err := http.Get(w.payload.FileURL)
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to get file content: %v", err)
		return err
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to read file content: %v", err)
		return err
	}
	w.content = string(all)

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
	cli := NewNBCli(&appCfg.Nebula)
	if err = nbgraph.DeleteFilesEntAndRel(context.TODO(), w.payload.ForestID, []uint{w.payload.FileID}, w.payload.ESIndex, cli); err != nil {
		logs.ErrorContextf(w.ctx, "Failed to delete nebula entities/relation: %v", err)
		err = nil
	}
	defer cli.Release()
	// 删除es数据
	err = DeleteFileReferences(w.ctx, w.escli, w.payload.ESIndex, []uint{w.payload.FileID})
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to delete file references: %v", err)
		return err
	}

	return err
}

func (w *insertIndexWorker) StoragePath() string {
	return filepath.Join(w.rootDir, "storage")
}

func (w *insertIndexWorker) Init() error {
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

func (w *insertIndexWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), w.taskResult)
}

func (w *insertIndexWorker) Close() error {
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
