package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
)

func init() {
	rootCmd.AddCommand(descCmd())
}

func descCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "desc",
		Short: "get file's description by agent",
		Run: func(cmd *cobra.Command, args []string) {
			//logs.Infof("Starting Description worker with worker ID: %s, task type: %s, webserver URL: %s", workerID, taskType, workerServerURL)
			ctx := lifecycle.Std().Context()
			loadConfig(ctx, configFile)

			for i := 0; i < routineSize; i++ {
				go descRoutine(ctx, i)
			}

			lifecycle.Std().WaitExit()
		},
	}
	withAgentFlags(cmd)
	return cmd
}

func descRoutine(ictx context.Context, idx int) {
	ctx := logs.WithContextFields(ictx, "routine_idx", idx)
	logs.InfoContextf(ctx, "Description worker started with routine index: %d", idx)
	for {
		select {
		case <-lifecycle.Std().C():
			return
		default:
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
			newDescWorker(ctx, taskId, payload).Run()
		}
	}
}

type descWorker struct {
	ctx     context.Context
	taskID  uint
	payload *ragtask.TaskPayload
	rootDir string
	content string
	record  *ragtypes.FileDescription
	wg      *sync.WaitGroup
	sigChan chan string
	calls   []SubCall

	taskStatus task.TaskStatus
	taskErrMsg []string
	taskResult interface{}
}

func newDescWorker(ctx context.Context, taskID uint, payload *ragtask.TaskPayload) *descWorker {
	taskKey := fmt.Sprintf("%v-%s", taskID, encryptor.UUID())
	now := time.Now()
	return &descWorker{
		ctx:     ctx,
		taskID:  taskID,
		payload: payload,
		rootDir: filepath.Join(os.TempDir(), "yg_desc", taskKey),
		wg:      &sync.WaitGroup{},
		record: &ragtypes.FileDescription{
			//* id pair
			Common: ragtypes.Common{
				ForestID:  payload.ForestID,
				Uin:       payload.Uin,
				CompanyID: payload.CompanyID,
				//* meta-type
				Type:      ragtypes.ChunkTypeFileDescription,
				CreatedAt: now,
				UpdatedAt: now,
				Version:   "",
			},

			FileID: payload.FileID,
		},
		sigChan: make(chan string, 1),

		taskStatus: task.TaskStatusFail,
	}
}

func (w *descWorker) Run() (err error) {
	defer func() {
		w.Close()
		if err != nil {
			w.taskErrMsg = append(w.taskErrMsg, err.Error())
			w.taskStatus = task.TaskStatusFail
		}
		if err := w.CallBack(); err != nil {
			logs.ErrorContextf(w.ctx, "Callback failed: %v", err)
		}
	}()

	if err := w.Init(); err != nil {
		logs.ErrorContextf(w.ctx, "Initialization failed: %v", err)
		w.taskErrMsg = append(w.taskErrMsg, fmt.Sprintf("Initialization failed: %v", err))
		return err
	}

	if err := w.PreRun(); err != nil {
		logs.ErrorContextf(w.ctx, "PreRun content failed: %v", err)
		return err
	}

	//*Do description sub jobs, for all sub workers start a subroutine
	//*mindmap
	//*abstract: (origin name is analysis)
	//*short description: (get short description and calculate embedding)
	w.calls = []SubCall{w.MindMap, w.Abstract, w.Description}
	for _, c := range w.calls {
		w.wg.Add(1)
		ct := c
		go func() {
			if err := func() error {
				defer func() {
					w.Show()
					w.wg.Done()
				}()
				return ct()
			}(); err == nil {
				return
			}
		}()
	}
	//wait until all subroutine done
	w.wg.Wait()
	//insert description es record
	if err := w.PostRun(); err != nil {
		logs.ErrorContextf(w.ctx, "PostRun content failed: %v", err)
		return err
	}

	w.taskStatus = task.TaskStatusSuccess
	w.taskResult = fmt.Sprintf("%v\n%v\n%v", w.record.Description, w.record.MindMap, w.record.Abstract)

	return nil
}

func (w *descWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), w.taskResult)
}

func (w *descWorker) PreRun() error {
	//delete deprecated es data
	if err := essearch.NewPureWrapper(w.ctx, w.payload.ESIndex,
		[]uint{w.payload.ForestID}, []uint{w.payload.FileID},
		esClient).DeleteType(ragtypes.ChunkTypeFileDescription); err != nil {
		return err
	}

	//get content about file's md
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
	w.Show()
	return err
}

func (w *descWorker) PostRun() error {
	//*store record to es
	if err := essearch.NewPureWrapper(w.ctx, w.payload.ESIndex,
		[]uint{w.payload.ForestID}, []uint{w.payload.FileID},
		esClient).Insert(w.record); err != nil {
		return err
	}
	return nil
}

func (w *descWorker) Init() error {
	return nil
}

func (w *descWorker) Close() error {
	if debug {
		return nil
	}
	return nil
}

func (w *descWorker) MindMap() error {
	//extract title for markdown
	//do agent request
	resp, err := doAgentRequest(w.ctx, map[string]string{
		"input1": strings.Join(ExtractMarkdownTitles(w.content), "\n"),
	}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDMindmap], appCfg.Agent.Pool[MDMindChunk], appCfg.Agent.Pool[MDMergeMindmap])
	if err != nil {
		logs.ErrorContextf(w.ctx, "doAgentRequest content failed: %v", err)
		return err
	}

	//extract json code block
	jsonCode := ExtractCode("json", resp)

	//process resp with embedded uuid
	embeddedUuid, err := ProcessEmbeddedUuid(jsonCode)
	if err != nil {
		return err
	}
	w.record.MindMap = embeddedUuid
	return nil
}

func (w *descWorker) Abstract() error {
	resp, err := doAgentRequest(w.ctx, map[string]string{
		"input1": w.content,
	}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDAbstract], appCfg.Agent.Pool[MDAbsChunk], appCfg.Agent.Pool[MDMergeAbstract])
	if err != nil {
		logs.ErrorContextf(w.ctx, "doAgentRequest failed: %v", err)
		return err
	}

	w.record.Abstract = resp
	//notify
	w.sigChan <- resp
	return nil
}

func (w *descWorker) Description() (err error) {
	for {
		select {
		case res := <-w.sigChan:
			var (
				resp      string
				embedding ragtypes.Embedding
			)

			resp, err = doAgentRequest(w.ctx, map[string]string{
				"input1": res,
			}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDShortDesc])
			if err != nil {
				logs.ErrorContextf(w.ctx, "doAgentRequest failed: %v", err)
				return
			}
			embedding, err = essearch.GetEmbedding(resp)
			if err != nil {
				logs.ErrorContextf(w.ctx, "GetEmbedding failed: %v", err)
				return
			}

			w.record.Description = resp
			w.record.Embedding = embedding
			return
		case <-time.Tick(time.Minute * 5):
			logs.ErrorContextf(w.ctx, "desc waiting sigChan timeout")
		}
	}
}

// Show func will display obj as a pretty show
func (w *descWorker) Show() {
	if w == nil {
		logs.InfoContextf(w.ctx, "[DescWorker]: nil")
		return
	}

	logs.InfoContextf(w.ctx, `DescWorker{
  taskID: %d
  rootDir: %s
  content: %s
  taskStatus: %s
  taskErrMsg: %v
  taskResult: %v
  payload: %s
  record: %s
}`,
		w.taskID,
		w.rootDir,
		w.truncate(w.content, 100),
		w.taskStatus,
		w.taskErrMsg,
		w.taskResult,
		w.showPayload(),
		w.showRecord(),
	)
}
func (w *descWorker) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
func (w *descWorker) showPayload() string {
	if w.payload == nil {
		return "nil"
	}
	return fmt.Sprintf("TaskPayload{ForestIDs:%v, FileID:%v, FileURL:%s, Uin:%d, CompanyID:%d, UploadPath:%s}",
		w.payload.ForestID, w.payload.FileID, w.truncate(w.payload.FileURL, 50),
		w.payload.Uin, w.payload.CompanyID, w.payload.UploadPath)
}
func (w *descWorker) showRecord() string {
	if w.record == nil {
		return "nil"
	}
	return fmt.Sprintf("Chunk{Type:%v, ForestIDs:%v, FileID:%v, Abstract:%s, Description:%s, MindMap:%s, EmbeddingLen:%d}",
		w.record.Type, w.record.ForestID, w.record.FileID,
		w.truncate(w.record.Abstract, 50), w.truncate(w.record.Description, 50),
		w.truncate(w.record.MindMap, 50), len(w.record.Embedding))
}
