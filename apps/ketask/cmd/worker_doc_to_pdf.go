package main

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/decoupler"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/lifecycle"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/storage"
)

func init() {
	rootCmd.AddCommand(doc2PDFCmd())
}

func doc2PDFCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc2pdf",
		Short: "transfer document to pdf",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InfoContextf(cmd.Context(), "Starting doc2pdf worker with worker id :%s, task type: %s, web server URL: %s", workerID, taskType, workerServerURL)
			ctx := lifecycle.Std().Context()
			loadConfig(ctx, configFile)
			// 初始化文件存储
			if err := fs.InitForestStorage(); err != nil {
				logs.FatalContextf(cmd.Context(), "[doc2PDFCmd] InitForestStorage failed, %s", err)
				return
			}
			// 初始化文件存储
			for i := 0; i < routineSize; i++ {
				go doc2PDFRoutine(ctx, i)
			}
			lifecycle.Std().WaitExit()
		},
	}
	return cmd
}

func doc2PDFRoutine(ictx context.Context, idx int) {
	ctx := logs.WithContextFields(ictx, "routine_idx", idx)
	logs.InfoContextf(ctx, "doc2PDFRoutine worker started with routine index: %d", idx)
	for {
		select {
		case <-lifecycle.Std().C():
			return
		default:
			payload := &ragtask.TaskPayload{}
			taskId, err := GetPendingTask(ctx, payload)
			if err != nil {
				logs.ErrorContextf(ctx, "doc2PDFRoutine worker Failed to get pending task: %v", err)
				time.Sleep(5 * time.Second) // Retry after a delay
				continue
			}
			logs.InfoContextf(ctx, "doc2PDFRoutine worker first received task ID: %d, %v", taskId, payload)
			if taskId == 0 {
				time.Sleep(2 * time.Second)
				continue
			}
			logs.InfoContextf(ctx, "doc2PDFRoutine worker second received task ID: %d, %v", taskId, payload)
			newDoc2PDFWorker(ctx, taskId, payload).Run()
		}
	}
}

type doc2PDFWorker struct {
	ctx        context.Context
	s3Cli      *storage.S3Fs
	taskID     uint
	payload    *ragtask.TaskPayload
	taskStatus task.TaskStatus
	taskErrMsg []string
	taskResult interface{}
}

func newDoc2PDFWorker(ctx context.Context, taskID uint, payload *ragtask.TaskPayload) *doc2PDFWorker {
	return &doc2PDFWorker{
		ctx:        ctx,
		s3Cli:      s3m[payload.Bucket],
		taskID:     taskID,
		payload:    payload,
		taskStatus: task.TaskStatusFail,
	}
}

func (w *doc2PDFWorker) Init() error {
	return nil
}

func (w *doc2PDFWorker) Close() error {
	if debug {
		return nil
	}
	return nil
}

func (w *doc2PDFWorker) Run() (err error) {
	defer func() {
		if err != nil {
			w.taskErrMsg = append(w.taskErrMsg, err.Error())
			w.taskStatus = task.TaskStatusFail
		}
		if err := w.CallBack(); err != nil {
			logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] Callback failed: %v", err)
		}
	}()
	logs.InfoContextf(w.ctx, "[doc2PDFWorker.Run] start processing file, core file id: %d, task id: %d", w.payload.FileID, w.taskID)
	f, err := fs.Forest.ReadFile(w.payload.StoragePath)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] ReadFile failed, core file id: %d, task id: %d, err: %v", w.payload.FileID, w.taskID, err)
		return err
	}

	var reader io.Reader
	var closeFunc func()
	switch w.payload.FileExt {
	case global.FileExtPPT, global.FileExtPPTX, global.FileExtDOC, global.FileExtDOCX, global.FileExtOFD:
		pdf, err := decoupler.FileToPDF(w.ctx, f, w.payload.Filename)
		if err != nil {
			logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] FileToPDF failed, core file id: %d, task id: %d, err: %v", w.payload.FileID, w.taskID, err)
			return err
		}
		reader = pdf
		closeFunc = func() {
			pdf.Close()
		}
	case global.FileExtCSV:
		excel, err := decoupler.CSVToExcel(w.ctx, f)
		if err != nil {
			logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] CSVToExcel failed, core file id: %d, task id: %d, err: %v", w.payload.FileID, w.taskID, err)
			return err
		}
		reader = excel
	default:
		return nil
	}

	if reader == nil {
		return nil
	}
	if closeFunc != nil {
		defer closeFunc()
	}

	previewFileEntity, err := storage.GetFileByID(dbutil.Core(), uint(w.payload.PreviewFileID))
	if err != nil {
		logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] GetFileByID failed, core file id: %d, task id: %d, err: %v", w.payload.PreviewFileID, w.taskID, err)
		return err
	}
	err = w.s3Cli.Save(w.ctx, previewFileEntity, reader)
	if err != nil {
		logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] Save failed, core file id: %d, task id: %d, err: %v", w.payload.PreviewFileID, w.taskID, err)
		return err
	}
	previewFileEntity.Status = storage.FileStatusNormal
	err = dbutil.Core().Save(previewFileEntity).Error
	if err != nil {
		logs.ErrorContextf(w.ctx, "[doc2PDFWorker.Run] update preview file failed, core file id: %d, task id: %d, err: %v", w.payload.PreviewFileID, w.taskID, err)
		return err
	}
	w.taskStatus = task.TaskStatusSuccess
	return nil
}

func (w *doc2PDFWorker) CallBack() error {
	return CallBackTask(w.ctx, w.taskID, w.taskStatus, strings.Join(w.taskErrMsg, ";"), w.taskResult)
}

func (w *doc2PDFWorker) PreRun() error {
	return nil
}

func (w *doc2PDFWorker) PostRun() error {
	return nil
}
