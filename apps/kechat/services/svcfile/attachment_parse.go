package svcfile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/decoupler"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	taskpkg "github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/httptools"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
)

const (
	// parseAppGroup identifies standalone attachment parse tasks in core_task.
	parseAppGroup = "pdf_md_convert"
	// parsePriority gives interactive attachment parsing higher priority than background forest parsing.
	parsePriority = 100
	// standaloneForestID marks files that do not belong to a knowledge forest.
	standaloneForestID = 0
	// markdownFilename is the output name written by the PDF parse worker.
	markdownFilename = "content.md"
	// taskPollInterval controls how frequently the service checks core_task state.
	taskPollInterval = 2 * time.Second
	// taskWaitTimeout limits how long the upload request waits for parsing to finish.
	taskWaitTimeout = 10 * time.Minute
	// outputCheckInterval controls retries while waiting for parsed output storage to become readable.
	outputCheckInterval = time.Second
	// outputCheckTimeout limits how long the service waits for parsed output storage to become readable.
	outputCheckTimeout = 30 * time.Second
	// outputRequestTimeout limits a single parsed output HTTP request.
	outputRequestTimeout = 10 * time.Second
)

// ParseRequest describes an uploaded file that may need conversion or Markdown parsing.
type ParseRequest struct {
	// SourceID identifies the original file in core_upload_files.
	SourceID uint
	// SourceURL is the public URL of the uploaded file.
	SourceURL string
	// Name is the original uploaded file name.
	Name string
	// Purpose selects the storage namespace for converted files.
	Purpose string
	// File provides the uploaded stream when format conversion is required.
	File multipart.File
}

// ParseResult describes the Markdown output and optional asynchronous task metadata.
type ParseResult struct {
	// TaskID identifies the core_task used for parsing.
	TaskID uint
	// Status is the final core_task status and remains empty for synchronous analyser calls.
	Status taskpkg.TaskStatus
	// URL is the public URL of the generated Markdown output.
	URL string
}

// ParseToMarkdown converts supported uploads when needed and returns their Markdown representation.
func ParseToMarkdown(ctx *gin.Context, req *ParseRequest) (*ParseResult, error) {
	if req == nil {
		err := errors.New("parse request is nil")
		logs.ErrorContextf(ctx, "[ParseToMarkdown] invalid request, err: %v", err)
		return nil, err
	}
	if req.SourceURL == "" {
		err := errors.New("source url is empty")
		logs.ErrorContextf(ctx, "[ParseToMarkdown] invalid request, sourceID: %d, err: %v", req.SourceID, err)
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(req.Name))
	sourceURL := req.SourceURL
	switch ext {
	case ".doc", ".docx", ".ofd", ".ppt", ".pptx":
		if req.File == nil {
			err := errors.New("uploaded file stream is nil")
			logs.ErrorContextf(ctx, "[ParseToMarkdown] missing file stream, sourceID: %d, err: %v", req.SourceID, err)
			return nil, err
		}
		if _, err := req.File.Seek(0, 0); err != nil {
			logs.ErrorContextf(ctx, "[ParseToMarkdown] seek uploaded file failed, sourceID: %d, err: %v", req.SourceID, err)
			return nil, fmt.Errorf("seek uploaded file: %w", err)
		}
		convertedURL, err := convertToPDF(ctx, req.File, req.Name, req.Purpose)
		if err != nil {
			logs.ErrorContextf(ctx, "[ParseToMarkdown] convert file to PDF failed, sourceID: %d, err: %v", req.SourceID, err)
			return nil, err
		}
		sourceURL = convertedURL
	case ".pdf":
	case ".txt", ".md", ".log", ".json":
		return &ParseResult{}, nil
	default:
		markdownURL, err := analyse(ctx, sourceURL)
		if err != nil {
			logs.ErrorContextf(ctx, "[ParseToMarkdown] analyse file failed, sourceID: %d, err: %v", req.SourceID, err)
			return nil, err
		}
		return &ParseResult{URL: markdownURL}, nil
	}

	taskID, err := createParseTask(ctx, req.SourceID, sourceURL, req.Name)
	if err != nil {
		logs.ErrorContextf(ctx, "[ParseToMarkdown] create parse task failed, sourceID: %d, err: %v", req.SourceID, err)
		return nil, err
	}
	markdownURL, taskStatus, err := waitParseTask(ctx, taskID)
	if err != nil {
		logs.ErrorContextf(ctx, "[ParseToMarkdown] wait parse task failed, taskID: %d, err: %v", taskID, err)
		return nil, err
	}

	return &ParseResult{
		TaskID: taskID,
		Status: taskStatus,
		URL:    markdownURL,
	}, nil
}

func convertToPDF(ctx *gin.Context, file multipart.File, filename, purpose string) (string, error) {
	pdfReader, err := decoupler.FileToPDF(ctx, file, filename)
	if err != nil {
		return "", fmt.Errorf("convert to pdf: %w", err)
	}
	defer pdfReader.Close()

	pdfInfo := &storage.FileInfo{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		Filename:  filename + ".pdf",
		Purpose:   purpose,
		FileExt:   ".pdf",
	}
	pdfInfo.StoragePath = storage.GenerateFileStoragePath(purpose, pdfInfo.Uin, pdfInfo.FileExt)

	storager, err := storage.LoadStorager(purpose)
	if err != nil {
		return "", fmt.Errorf("load storager: %w", err)
	}
	if err := storager.Save(ctx, pdfInfo, pdfReader); err != nil {
		return "", fmt.Errorf("save pdf: %w", err)
	}
	if err := dbutil.Core().WithContext(ctx).Create(pdfInfo).Error; err != nil {
		return "", fmt.Errorf("create pdf info: %w", err)
	}
	return storager.GetPublicURL(pdfInfo.StoragePath, false), nil
}

func analyse(ctx context.Context, sourceURL string) (string, error) {
	var cfg struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	}
	if err := settings.GetYaml("corekg", "yg_api_analysis_file", &cfg); err != nil {
		return "", fmt.Errorf("load analyser config: %w", err)
	}
	if cfg.URL == "" {
		return "", errors.New("analyser url is empty")
	}
	if cfg.Token == "" {
		return "", errors.New("analyser token is empty")
	}

	payload := map[string]any{
		"request": map[string]any{
			"output_formats": "md",
			"file_url":       sourceURL,
		},
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal analyser request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("build analyser request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call analyser: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read analyser response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("analyser returned %s: %s", resp.Status, string(responseBody))
	}

	var result struct {
		Code     int    `json:"code"`
		Message  string `json:"message"`
		Response struct {
			URL string `json:"md_url"`
		} `json:"response"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal analyser response: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("analyser returned code %d: %s", result.Code, result.Message)
	}
	if result.Response.URL == "" {
		return "", errors.New("analyser returned empty md_url")
	}
	return result.Response.URL, nil
}

func createParseTask(ctx *gin.Context, sourceID uint, sourceURL, filename string) (uint, error) {
	uin := runtime.Uin(ctx)
	companyID := runtime.CompanyID(ctx)
	if sourceID == 0 {
		err := errors.New("source file id is empty")
		logs.ErrorContextf(ctx, "[createParseTask] invalid source file id, err: %v", err)
		return 0, err
	}
	if sourceURL == "" {
		err := errors.New("file url is empty")
		logs.ErrorContextf(ctx, "[createParseTask] invalid file url, sourceID: %d, err: %v", sourceID, err)
		return 0, err
	}

	var cfg config.StorageConfig
	if err := settings.GetYaml("core", storage.SettingPrefix+fs.PurposeKeFile, &cfg); err != nil {
		logs.ErrorContextf(ctx, "[createParseTask] load storage config failed, sourceID: %d, err: %v", sourceID, err)
		return 0, fmt.Errorf("load storage config error: %w", err)
	}
	if cfg.S3 == nil || cfg.S3.Bucket == "" {
		err := errors.New("storage s3 bucket is empty")
		logs.ErrorContextf(ctx, "[createParseTask] invalid storage config, sourceID: %d, err: %v", sourceID, err)
		return 0, err
	}

	outputPath := buildOutputPath(uin, filename)
	payload := &ragtask.TaskPayload{
		CommonPayload: taskpkg.CommonPayload{
			TaskType: coretask.PraseTask,
			Timeout:  int64(coretask.TaskTimeout),
		},
		FileID:     sourceID,
		FileURL:    sourceURL,
		Filename:   filename,
		SubjectID:  sourceID,
		UploadPath: outputPath,
		CompanyID:  companyID,
		ForestID:   standaloneForestID,
		Uin:        uin,
		Bucket:     cfg.S3.Bucket,
		FileExt:    ".pdf",
		FileName:   filename,
	}
	tsk := &taskpkg.Task{
		Uin:               uin,
		CompanyID:         companyID,
		SubjectID:         sourceID,
		TaskType:          coretask.PraseTask,
		Priority:          parsePriority,
		TaskStatus:        taskpkg.TaskStatusPending,
		Comment:           "chat attachment pdf to markdown task",
		Payload:           payload.String(),
		AppGroup:          parseAppGroup,
		TaskConfigRedo:    coretask.TaskRedo,
		TaskConfigTimeout: coretask.TaskTimeout,
	}
	if err := taskpkg.CreateTask(tsk); err != nil {
		logs.ErrorContextf(ctx, "[createParseTask] create task failed, sourceID: %d, err: %v", sourceID, err)
		return 0, fmt.Errorf("create task error: %w", err)
	}
	logs.InfoContextf(ctx, "[createParseTask] task created, sourceID: %d, taskID: %d, outputPath: %s", sourceID, tsk.ID, outputPath)
	return tsk.ID, nil
}

func buildOutputPath(uin uint, filename string) string {
	baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	baseName = strings.TrimSpace(baseName)
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-")
	baseName = replacer.Replace(baseName)
	baseName = strings.Trim(baseName, ".- ")
	if baseName == "" {
		baseName = "file"
	}
	return fmt.Sprintf("%s/%d/%d/%s-%d-%s/", fs.PurposeForestAlgo, uin, standaloneForestID, baseName, time.Now().UnixMilli(), strings.ToLower(random.Alphanum(6)))
}

func waitParseTask(ctx context.Context, taskID uint) (string, taskpkg.TaskStatus, error) {
	pollCtx, cancel := context.WithTimeout(ctx, taskWaitTimeout)
	defer cancel()
	ticker := time.NewTicker(taskPollInterval)
	defer ticker.Stop()

	for {
		tsk, err := taskpkg.GetTaskByID(taskID)
		if err != nil {
			logs.ErrorContextf(ctx, "[waitParseTask] get task by id failed, taskID: %d, err: %v", taskID, err)
			return "", "", fmt.Errorf("get task by id error: %w", err)
		}

		switch tsk.TaskStatus {
		case taskpkg.TaskStatusPending, taskpkg.TaskStatusRunning:
		case taskpkg.TaskStatusSuccess:
			outputURL, err := getOutputURL(ctx, tsk)
			if err != nil {
				logs.ErrorContextf(ctx, "[waitParseTask] get output url failed, taskID: %d, err: %v", taskID, err)
				return "", tsk.TaskStatus, err
			}
			if err := waitOutputReady(pollCtx, outputURL); err != nil {
				logs.ErrorContextf(ctx, "[waitParseTask] output is not readable, taskID: %d, outputURL: %s, err: %v", taskID, outputURL, err)
				return "", tsk.TaskStatus, err
			}
			logs.InfoContextf(ctx, "[waitParseTask] task success, taskID: %d, outputURL: %s", taskID, outputURL)
			return outputURL, tsk.TaskStatus, nil
		case taskpkg.TaskStatusFail, taskpkg.TaskStatusCancel, taskpkg.TaskStatusTimeout:
			err := fmt.Errorf("task status is %s, err_msg: %s", tsk.TaskStatus, tsk.ErrMsg)
			logs.ErrorContextf(ctx, "[waitParseTask] task finished unsuccessfully, taskID: %d, err: %v", taskID, err)
			return "", tsk.TaskStatus, err
		default:
			err := fmt.Errorf("unsupported task status: %s", tsk.TaskStatus)
			logs.ErrorContextf(ctx, "[waitParseTask] unsupported task status, taskID: %d, err: %v", taskID, err)
			return "", tsk.TaskStatus, err
		}

		select {
		case <-pollCtx.Done():
			err := fmt.Errorf("wait parse task timeout: %w", pollCtx.Err())
			logs.ErrorContextf(ctx, "[waitParseTask] wait parse task timeout, taskID: %d, err: %v", taskID, err)
			return "", "", err
		case <-ticker.C:
		}
	}
}

func getOutputURL(ctx context.Context, tsk *taskpkg.Task) (string, error) {
	payload := &ragtask.TaskPayload{}
	if err := json.Unmarshal([]byte(tsk.Payload), payload); err != nil {
		logs.ErrorContextf(ctx, "[getOutputURL] unmarshal task payload failed, taskID: %d, err: %v", tsk.ID, err)
		return "", fmt.Errorf("unmarshal task payload error: %w", err)
	}
	if payload.UploadPath == "" {
		err := errors.New("task payload upload_path is empty")
		logs.ErrorContextf(ctx, "[getOutputURL] invalid task payload, taskID: %d, err: %v", tsk.ID, err)
		return "", err
	}

	outputPath := path.Join(payload.UploadPath, markdownFilename)
	st, err := storage.LoadStorager(fs.PurposeKeFile)
	if err != nil {
		logs.ErrorContextf(ctx, "[getOutputURL] load storager failed, taskID: %d, err: %v", tsk.ID, err)
		return "", fmt.Errorf("load storager error: %w", err)
	}
	outputURL := st.GetPublicURL(outputPath, false)
	if outputURL == "" {
		err := errors.New("output url is empty")
		logs.ErrorContextf(ctx, "[getOutputURL] empty output url, taskID: %d, storagePath: %s, err: %v", tsk.ID, outputPath, err)
		return "", err
	}
	return outputURL, nil
}

func waitOutputReady(ctx context.Context, outputURL string) error {
	checkCtx, cancel := context.WithTimeout(ctx, outputCheckTimeout)
	defer cancel()
	ticker := time.NewTicker(outputCheckInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := checkOutput(checkCtx, outputURL); err == nil {
			return nil
		} else {
			lastErr = err
			logs.WarnContextf(ctx, "[waitOutputReady] output is not readable yet, outputURL: %s, err: %v", outputURL, err)
		}

		select {
		case <-checkCtx.Done():
			err := fmt.Errorf("output is not readable: %w, last error: %v", checkCtx.Err(), lastErr)
			logs.ErrorContextf(ctx, "[waitOutputReady] output check timeout, outputURL: %s, err: %v", outputURL, err)
			return err
		case <-ticker.C:
		}
	}
}

func checkOutput(ctx context.Context, outputURL string) error {
	reqCtx, cancel := context.WithTimeout(ctx, outputRequestTimeout)
	defer cancel()

	resp, err := httptools.Get(reqCtx, outputURL)
	if err != nil {
		logs.ErrorContextf(ctx, "[checkOutput] open output failed, outputURL: %s, err: %v", outputURL, err)
		return fmt.Errorf("open output: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.CopyN(io.Discard, resp.Body, 1); err != nil && !errors.Is(err, io.EOF) {
		logs.ErrorContextf(ctx, "[checkOutput] read output failed, outputURL: %s, err: %v", outputURL, err)
		return fmt.Errorf("read output: %w", err)
	}
	return nil
}
