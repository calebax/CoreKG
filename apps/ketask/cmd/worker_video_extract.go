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
	rootCmd.AddCommand(videoExtraTextCmd())
}

func videoExtraTextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video_extract",
		Short: "Extract text from Video files",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InfoContextf(cmd.Context(), "Starting Video text extraction worker with worker ID: %s, task type: %s, base URL: %s", workerID, taskType, baseURL)
			ctx := lifecycle.Std().Context()
			loadConfig(ctx, configFile)
			for i := 0; i < routineSize; i++ {
				go videoExtraTextRoutine(ctx, i)
			}

			lifecycle.Std().WaitExit()
		},
	}
	return cmd
}

func videoExtraTextRoutine(ictx context.Context, idx int) {
	ctx := logs.WithContextFields(ictx, "routine_idx", idx)
	logs.InfoContextf(ctx, "Video text extraction worker started with routine index: %d", idx)
	for {
		select {
		case <-lifecycle.Std().C():
			return
		default:
			if err := checkURL(workerServerURL); err != nil {
				logs.WarnContextf(ctx, "Worker server URL(%s) check failed: %v", workerServerURL, err)
				time.Sleep(5 * time.Second) // Retry after a delay
				continue
			}
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
			newVideoExtraTextWorker(ctx, taskId, payload).Run()
		}
	}
}

type videoExtraTextWorker struct {
	ctx     context.Context
	s3Cli   *storage.S3Fs
	taskID  uint
	payload *ragtask.TaskPayload
	rootDir string

	taskStatus task.TaskStatus
	taskErrMsg []string
	taskResult interface{}
}

func newVideoExtraTextWorker(ctx context.Context, taskID uint, payload *ragtask.TaskPayload) *videoExtraTextWorker {
	taskKey := fmt.Sprintf("%v-%s", taskID, encryptor.UUID())
	return &videoExtraTextWorker{
		s3Cli:   s3m[payload.Bucket],
		ctx:     logs.WithContextFields(ctx, "task_id", taskID, "task_key", taskKey),
		taskID:  taskID,
		payload: payload,
		rootDir: filepath.Join(os.TempDir(), "yg_video_extratext", taskKey),

		taskStatus: task.TaskStatusFail,
	}
}

func (w *videoExtraTextWorker) Run() (err error) {
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
		logs.ErrorContextf(w.ctx, "Failed to prepare Video: %v", err)
		w.taskErrMsg = append(w.taskErrMsg, fmt.Sprintf("Failed to prepare Video: %v", err))
		return err
	}

	if err := w.ProcessVideo(); err != nil {
		logs.ErrorContextf(w.ctx, "Failed to process Video: %v", err)
		w.taskErrMsg = append(w.taskErrMsg, fmt.Sprintf("Failed to process Video: %v", err))
		return err
	}

	err = w.PostRun()
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to upload Video: %v", err)
		w.taskErrMsg = append(w.taskErrMsg, fmt.Sprintf("Failed to upload Video: %v", err))
		return err
	}
	w.taskStatus = task.TaskStatusSuccess

	return nil
}

func (w *videoExtraTextWorker) Init() error {
	if w.s3Cli == nil {
		logs.ErrorContextf(w.ctx, "S3 client (%s) is not initialized", w.payload.Bucket)
		return fmt.Errorf("S3 client is not initialized")
	}
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

func (w *videoExtraTextWorker) PreRun() error {
	return DownloadFile(w.ctx, w.payload.FileURL, w.OriginFilePath())
}

func (w *videoExtraTextWorker) PostRun() error {
	logs.InfoContextf(w.ctx, "Uploading files from storage path: %s", w.StoragePath())
	files, err := w.s3Cli.UploadDirectory(w.StoragePath(), w.payload.UploadPath)
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to upload directory: %v", err)
		w.taskErrMsg = append(w.taskErrMsg, fmt.Sprintf("Failed to upload directory: %v", err))
		return err
	}
	w.taskResult = files
	logs.InfoContextf(w.ctx, "Successfully uploaded files %v", len(files))
	return nil
}

func (w *videoExtraTextWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), w.taskResult)
}

func (w *videoExtraTextWorker) Close() error {
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

func (w *videoExtraTextWorker) OriginFilePath() (path string) {
	path = filepath.Join(w.rootDir, "origin.video")
	u, _ := url.Parse(w.payload.FileURL)
	if u == nil {
		return
	}
	path = filepath.Join(w.rootDir, "origin"+filepath.Ext(u.Path))
	return
}

func (w *videoExtraTextWorker) StoragePath() string {
	return filepath.Join(w.rootDir, "storage")
}

func (w *videoExtraTextWorker) ProcessVideo() error {
	logs.InfoContextf(w.ctx, "Processing video: %s", w.OriginFilePath())
	cli := http.Client{}
	reqBody := map[string]interface{}{
		"video_path":      w.OriginFilePath(),
		"output_base_dir": w.StoragePath(),
		"image_prefix":    w.s3Cli.GetPublicURL(w.payload.UploadPath, false),
		// "threshold":              w.payload.VideoExtractext.Threshold,
		// "frame_interval_seconds": w.payload.VideoExtractext.FrameIntervalSeconds,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to marshal request body: %v", err)
		return err
	}
	req, err := http.NewRequest("POST", workerServerURL, bytes.NewBuffer(reqData))
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		logs.ErrorContextf(w.ctx, "Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(w.ctx, "Received non-OK HTTP status: %s, body: %s", resp.Status, string(body))
		return fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	return nil
}
