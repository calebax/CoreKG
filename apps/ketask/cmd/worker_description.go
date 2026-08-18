package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
)

func init() {
	rootCmd.AddCommand(descCmd())
}

func descCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "description",
		Short: "get file's description by agent",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InfoContextf(cmd.Context(), "Starting Description worker with worker ID: %s, task type: %s, webserver URL: %s", workerID, taskType, workerServerURL)
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
		sigChan: make(chan string, 2),

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
	g, gCtx := errgroup.WithContext(w.ctx)

	w.calls = []SubCall{w.MindMap, w.Abstract, w.Description, w.RecommendQuestion}
	for _, c := range w.calls {
		ct := c
		g.Go(func() error {
			defer w.Show()
			return ct(gCtx)
		})
	}

	// 等待所有协程完成，并获取第一个错误
	if err := g.Wait(); err != nil {
		logs.ErrorContextf(w.ctx, "Sub-workers failed: %v", err)
		// errgroup.Wait() 返回的 err 就是第一个失败协程的错误
		return err
	}

	//insert description es record
	if err := w.PostRun(); err != nil {
		logs.ErrorContextf(w.ctx, "PostRun content failed: %v", err)
		return err
	}

	w.taskStatus = task.TaskStatusSuccess
	w.taskResult = fmt.Sprintf("%v\n%v\n%v",
		w.truncate(w.record.Description, 50),
		w.truncate(w.record.MindMap, 50),
		w.truncate(w.record.Abstract, 50))

	return nil
}

func (w *descWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), map[string]string{
		"description": w.truncate(w.record.Description, 50),
		"mind_map":    w.truncate(w.record.MindMap, 50),
		"abstract":    w.truncate(w.record.Abstract, 50),
	})
}

func (w *descWorker) PreRun() error {
	//delete deprecated es data
	if err := essearch.NewPureWrapper(w.ctx, w.payload.ESIndex,
		[]uint{w.payload.ForestID}, []uint{w.payload.FileID},
		esClient).DeleteType(ragtypes.ChunkTypeFileDescription); err != nil {
		return err
	}

	switch w.payload.FileExt {
	//if orig-file's ext equals to .jpg,search its chunk
	case ".jpg", ".png", "jpeg", ".svg":
		logs.DebugContextf(w.ctx, "image would be found as chunk(type:image)")
		chunks, err := essearch.NewPureWrapper(w.ctx, w.payload.ESIndex,
			[]uint{w.payload.ForestID}, []uint{w.payload.FileID},
			esClient).GetFileChunk(ragtypes.ChunkTypeImage)
		if err != nil {
			logs.ErrorContextf(w.ctx, "GetFileChunk failed: %v", err)
			return err
		}

		if len(chunks) == 0 {
			logs.DebugContextf(w.ctx, "image chunks len:%v, try to find chunk type and aggregate description", len(chunks))
			cks, err := essearch.NewPureWrapper(w.ctx, w.payload.ESIndex,
				[]uint{w.payload.ForestID}, []uint{w.payload.FileID},
				esClient).GetFileChunk(ragtypes.ChunkTypeChunk, ragtypes.ChunkTypeTable)
			if err != nil {
				logs.ErrorContextf(w.ctx, "GetFileChunk failed: %v", err)
				return err
			}
			if len(cks) == 0 {
				logs.DebugContextf(w.ctx, "chunk type and aggregate description failed")
				return fmt.Errorf("chunk type and aggregate description failed no chunks found")
			}

			w.content = "整体描述如下:\n"
			for i, v := range cks {
				logs.DebugContextf(w.ctx, "chunk index:%v description:%v", i, v.Source.Description)
				switch v.Source.Type {
				case ragtypes.ChunkTypeChunk:
					w.content += fmt.Sprintf("文本内容[%v]:\n%v\n", i, v.Source.Description)
				case ragtypes.ChunkTypeTable:
					w.content += fmt.Sprintf("表格内容[%v]:\n%v\n", i, v.Source.Description)
				}
			}
		} else {
			logs.DebugContextf(w.ctx, "image chunks len:%v, use first chunk description", len(chunks))
			w.content = chunks[0].Source.Description
		}
		w.Show()
		return nil
	//if equals to .mp4 then get chunks and sort by seq and construct it
	case ".mp4":
		logs.DebugContextf(w.ctx, "video would be found as chunk(type:image,video)")
		chunks, err := essearch.NewPureWrapper(w.ctx, w.payload.ESIndex,
			[]uint{w.payload.ForestID}, []uint{w.payload.FileID},
			esClient).GetFileChunk(ragtypes.ChunkTypeVideo, ragtypes.ChunkTypeImage)
		if err != nil {
			logs.ErrorContextf(w.ctx, "GetFileChunk failed: %v", err)
			return err
		}

		logs.DebugContextf(w.ctx, "chunk list len:%v", len(chunks))

		if len(chunks) == 0 {
			logs.DebugContextf(w.ctx, "video chunks len:%v", len(chunks))
			w.content = "该视频没有描述信息"
		} else {
			slices.SortFunc(chunks, func(a, b essearch.Hits) int {
				return a.Source.Sequence - b.Source.Sequence
			})
			w.content = "整体描述如下:\n"
			for i, v := range chunks {
				logs.DebugContextf(w.ctx, "chunk index:%v description:%v", i, v.Source.Description)
				if len(v.Source.Description) > 0 {
					w.content += v.Source.Description + "\n"
				}
			}
		}

		w.Show()
		return nil
	default:
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

func (w *descWorker) MindMap(ctx context.Context) error {
	//extract title for markdown
	//do agent request
	resp, err := doAgentRequest(ctx, map[string]string{
		"input1": w.content,
	}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDMindmap], appCfg.Agent.Pool[MDMindChunk], appCfg.Agent.Pool[MDMergeMindmap])
	if err != nil {
		logs.ErrorContextf(ctx, "[MindMap]doAgentRequest content failed: %v", err)
		return err
	}

	//extract json code block
	jsonCode := ExtractCode("json", resp)

	if len(jsonCode) == 0 {
		n := &Node{
			ID:   "BlankNode",
			UUID: uuid.NewString(),
		}
		bt, err := json.Marshal(n)
		if err != nil {
			return err
		}

		w.record.MindMap = string(bt)
		return nil
	}

	//process resp with embedded uuid
	embeddedUuid, err := ProcessEmbeddedUuid(jsonCode)
	if err != nil {
		return err
	}
	w.record.MindMap = embeddedUuid
	return nil
}

func (w *descWorker) Abstract(ctx context.Context) error {
	resp, err := doAgentRequest(ctx, map[string]string{
		"input1": w.content,
	}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDAbstract], appCfg.Agent.Pool[MDAbsChunk], appCfg.Agent.Pool[MDMergeAbstract])
	if err != nil {
		logs.ErrorContextf(ctx, "[Abstract]doAgentRequest failed: %v", err)
		return err
	}

	w.record.Abstract = resp
	//notify
	w.sigChan <- resp
	w.sigChan <- resp
	return nil
}

func (w *descWorker) Description(ctx context.Context) (err error) {
	for {
		select {
		case res := <-w.sigChan:
			var (
				resp      string
				embedding ragtypes.Embedding
			)

			resp, err = doAgentRequest(ctx, map[string]string{
				"input1": res,
			}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDShortDesc])
			if err != nil {
				logs.ErrorContextf(ctx, "[Description]doAgentRequest failed: %v", err)
				return
			}
			embedding, err = GetEmbedding(resp)
			if err != nil {
				logs.ErrorContextf(ctx, "[Description]GetEmbedding failed: %v", err)
				return
			}

			w.record.Description = resp
			w.record.Embedding = embedding
			return
		case <-time.Tick(time.Minute * 30):
			logs.ErrorContextf(ctx, "desc waiting sigChan timeout")
			return fmt.Errorf("desc waiting sigChan timeout")
		case <-ctx.Done():
			logs.ErrorContextf(ctx, "desc's ctx done")
			return fmt.Errorf("desc's ctx done")
		}
	}
}

func (w *descWorker) RecommendQuestion(ctx context.Context) (err error) {
	for {
		select {
		case res := <-w.sigChan:
			var (
				resp string
			)

			resp, err = doAgentRequest(ctx, map[string]string{
				"input1": res,
			}, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, appCfg.Agent.Pool[MDQuestions])
			if err != nil {
				logs.ErrorContextf(ctx, "[RecommendQuestion]doAgentRequest failed: %v", err)
				return
			}

			w.record.Questions = strings.Split(strings.TrimSpace(resp), ",")
			return
		case <-time.Tick(time.Minute * 30):
			logs.ErrorContextf(ctx, "desc waiting sigChan timeout")
			return fmt.Errorf("desc waiting sigChan timeout")
		case <-ctx.Done():
			logs.ErrorContextf(ctx, "RecommendQuestion's ctx done")
			return fmt.Errorf("RecommendQuestion's ctx done")
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
