package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/storage"
)

func init() {
	rootCmd.AddCommand(CopyCmd())
}

func CopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "copy file by agent",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InfoContextf(cmd.Context(), "Starting copy worker with worker ID: %s, task type: %s, webserver URL: %s", workerID, taskType, workerServerURL)
			if len(model)*len(apiKey) == 0 {
				logs.ErrorContextf(cmd.Context(), "Empty agent model[%v] or apikey[%v]", model, apiKey)
			}
			ctx := lifecycle.Std().Context()
			loadConfig(ctx, configFile)
			for i := 0; i < routineSize; i++ {
				go copyRoutine(ctx, i)
			}

			lifecycle.Std().WaitExit()
		},
	}
	withAgentFlags(cmd)
	return cmd
}

func copyRoutine(ictx context.Context, idx int) {
	ctx := logs.WithContextFields(ictx, "routine_idx", idx)
	logs.InfoContextf(ctx, "copy worker started with routine index: %d", idx)
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
			newcopyWorker(ctx, taskId, payload).Run()
		}
	}
}

type copyWorker struct {
	ctx     context.Context
	taskID  uint
	payload *ragtask.TaskPayload
	rootDir string
	content string
	s3Cli   *storage.S3Fs
	r       io.Reader

	taskStatus task.TaskStatus
	taskErrMsg []string
	taskResult interface{}
}

func newcopyWorker(ctx context.Context, taskID uint, payload *ragtask.TaskPayload) *copyWorker {
	taskKey := fmt.Sprintf("%v-%s", taskID, encryptor.UUID())
	return &copyWorker{
		ctx:     ctx,
		taskID:  taskID,
		payload: payload,
		rootDir: filepath.Join(os.TempDir(), "yg_copy", taskKey),
		s3Cli:   s3m[payload.Bucket],

		taskStatus: task.TaskStatusFail,
	}
}

func (w *copyWorker) Run() (err error) {
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

	if err = w.Init(); err != nil {
		logs.ErrorContextf(w.ctx, "Initialization failed: %v", err)
		w.taskErrMsg = append(w.taskErrMsg, fmt.Sprintf("Initialization failed: %v", err))
		return err
	}

	if err = w.PreRun(); err != nil {
		logs.ErrorContextf(w.ctx, "PreRun content failed: %v", err)
		return err
	}

	//logs.Debugf("copy agent chat response: %v", resp)
	//
	//w.r = strings.NewReader(resp)

	err = w.PostRun()

	w.taskStatus = task.TaskStatusSuccess
	w.taskResult = w.payload.UploadPath

	return nil
}

func (w *copyWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), w.taskResult)
}

func (w *copyWorker) PreRun() error {
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
	return err
}

func (w *copyWorker) PostRun() error {
	if err := w.s3Cli.Save(
		w.ctx,
		&storage.FileInfo{
			StoragePath: w.payload.UploadPath + "content.md",
			FileExt:     path.Ext(w.payload.UploadPath)},
		strings.NewReader(w.content)); err != nil {
		return err
	}
	return nil
}

func (w *copyWorker) Init() error {
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

func (w *copyWorker) Close() error {
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

func (w *copyWorker) StoragePath() string {
	return filepath.Join(w.rootDir, "storage")
}
