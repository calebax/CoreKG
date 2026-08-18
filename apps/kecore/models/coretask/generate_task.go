package coretask

import (
	"context"
	"strings"
	"time"

	chatmodel "github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/systemconfig"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
	"gorm.io/gorm"
)

const (
	CopyTask        = "ke.copy_task"        // 文件格式转换上传任务ss
	Doc2PDFTask     = "ke.doc_to_pdf_task"  // 文档转PDF任务
	PraseTask       = "ke.prase_pdf_task"   // 解析PDF任务
	PraseVideoTask  = "ke.prase_video_task" // 解析视频任务
	MindMapTask     = "ke.mind_map_task"    // 思维导图任务
	AnalysisTask    = "ke.analysis_task"    // 分析任务
	DescriptionTask = "ke.description_task"
	KnowledgeTask   = "ke.knowledge_task"    // 知识库任务
	GraphTask       = "ke.graph_task"        // 图谱任务
	TaskTimeout     = 150 * time.Minute      // 任务超时时间 30分钟
	TaskRedo        = 3                      // 任务重试次数
	TaskPriority    = 10                     // 任务优先级
	AppGroup        = "forest"

	GraphFileTask = "ke.graph_file_task" // 图谱任务
	GraphAppGroup = "graph"
)

// -------------------- 创建任务入口 --------------------

// CreateForestTask 创建任务
func CreateForestTask(ctx context.Context, file *foresttype.KnownowForestFile) error {

	var cfg config.StorageConfig
	if err := settings.GetYaml("core", "cos-ke", &cfg); err != nil {
		logs.ErrorContextf(ctx, "get storage config error: %v", err)
		return err
	}

	forestInfo, err := forest.GetForestByID(ctx, file.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "get forest_info error: %v", err)
		return err
	}

	doc2pdfTask, err := GenerateDoc2pdfTask(ctx, file, cfg.S3.Bucket)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateForestTask] GenerateDoc2pdfTask failed, forest file id: %d, error %v", file.ID, err)
		return err
	}

	ptsk := GenerateFileTask(ctx, forestInfo, file, cfg.S3.Bucket)

	var firstTask *task.Task
	if doc2pdfTask != nil {
		firstTask = doc2pdfTask
	} else {
		firstTask = ptsk
	}
	if firstTask == nil {
		logs.ErrorContextf(ctx, "[CreateForestTask] no first task, forest file id: %d", file.ID)
		return nil
	}
	if err := dbutil.Core().Create(firstTask).Error; err != nil {
		logs.ErrorContextf(ctx, "CreateForestTask error %v", err)
		return err
	}
	if err := task.PushTaskQueue(ctx, firstTask.TaskType); err != nil {
		logs.ErrorContextf(ctx, "PushTaskQueue error %v", err)
		return err
	}
	return nil
}

func GenerateDoc2pdfTask(ctx context.Context, file *foresttype.KnownowForestFile, bucket string) (*task.Task, error) {
	filePath, err := file.GetForestFilePath()
	if err != nil {
		logs.ErrorContextf(ctx, "[GenerateDoc2pdfTask] GetForestFilePath failed, forest file id: %d, error %v", file.ID, err)
		return nil, err
	}
	previewFileEntity, err := storage.GetFileByID(dbutil.Core().WithContext(ctx), file.PriviewFileID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GenerateDoc2pdfTask] GetFileByID failed, preview file id: %d, error %v", file.PriviewFileID, err)
		return nil, err
	}
	if previewFileEntity.Status != storage.FileStatusUploading {
		logs.InfoContextf(ctx, "[GenerateDoc2pdfTask] preview file is not uploading, file id: %d, status: %s", file.PriviewFileID, previewFileEntity.Status)
		return nil, nil
	}
	originalFileEntity, err := storage.GetFileByID(dbutil.Core().WithContext(ctx), file.CoreFileID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GenerateDoc2pdfTask] GetFileByID failed, original file id: %d, error %v", file.CoreFileID, err)
		return nil, err
	}
	payload := buildTaskPayload(
		file,
		&task.CommonPayload{
			TaskType: Doc2PDFTask,
			Timeout:  int64(TaskTimeout),
		},
		bucket)
	payload.FileURL = fs.Forest.GetPublicURL(*filePath, false)
	payload.FileExt = file.Ext
	payload.StoragePath = originalFileEntity.StoragePath
	payload.UploadPath = previewFileEntity.StoragePath
	payload.PreviewFileID = previewFileEntity.ID
	return buildTask(file, Doc2PDFTask, payload, "doc转pdf任务"), nil
}

// GenerateFileTask 生成解析任务
func GenerateFileTask(ctx context.Context, forest_info *foresttype.KnownowForest, file *foresttype.KnownowForestFile, bucket string) *task.Task {
	filePath, err := file.GetForestPriviewFilePath()
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestPriviewFilePath error %v", err)
		return nil
	}
	payload := buildTaskPayload(
		file,
		&task.CommonPayload{
			TaskType: PraseTask,
			Timeout:  int64(TaskTimeout),
		},
		bucket,
	)
	payload.FileURL = fs.Forest.GetPublicURL(*filePath, false)
	payload.UploadPath = fs.FileFileAlgoPath(file.GetAlgoFilePath(), file.ID)
	payload.ForestType = string(forest_info.ForestType)
	return buildTask(file, PraseTask, payload, "pdf解析md任务")
}

// GenerateKnowledgeTask 单文档建库chunk
func GenerateKnowledgeTask(ctx context.Context, file *foresttype.KnownowForestFile, forest_info *foresttype.KnownowForest, model chattype.ChatModel, bucket string) *task.Task {
	// 制作payload
	payload := &ragtask.TaskPayload{
		CommonPayload: task.CommonPayload{
			TaskType: KnowledgeTask,
			Timeout:  int64(TaskTimeout)},
		FileID:     file.ID,
		FileURL:    fs.Forest.GetPublicURL(fs.FileContentPath(file.GetAlgoFilePath(), file.ID), false),
		SubjectID:  file.ID,
		UploadPath: fs.FileKnowledgeDirPath(file.GetAlgoFilePath(), file.ID),
		CompanyID:  file.CompanyID,
		Bucket:     bucket,
		ForestID:   file.ForestID,
		Uin:        file.Uin,
		ESIndex:    forest_info.EsIndex(),
		FileExt:    file.PriviewExt,
		LLM: &config.LLMModelConfig{
			ModelName: model.ModelName,
			BaseURL:   strings.TrimSuffix(model.ModelUrl, "/chat/completions"),
			APIKEY:    model.APIKey,
		},
		SplitConfig: file.FileConfig.SplitConfig,
		FileName:    file.Name,
	}
	return buildTask(file, KnowledgeTask, payload, "单文档建库chunk")
}

// CreateReChunkTask 创建任务
func CreateReChunkTask(ctx context.Context, tx *gorm.DB, file *foresttype.KnownowForestFile, forest_info *foresttype.KnownowForest) error {
	var cfg config.StorageConfig
	if err := settings.GetYaml("core", "cos-ke", &cfg); err != nil {
		logs.ErrorContextf(ctx, "get storage config error11: %v", err)
		return err
	}
	model := loadChatModel(ctx)
	ktsk := GenerateKnowledgeTask(ctx, file, forest_info, model, cfg.S3.Bucket)
	// gpsk := GenerateGraphTask(file, forest_info, model, cfg.S3.Bucket)
	tsks := []*task.Task{ktsk}
	if err := tx.WithContext(ctx).CreateInBatches(tsks, 1).Error; err != nil {
		logs.ErrorContextf(ctx, "CreateReChunkTask error %v", err)
		return err
	}

	// 插入完后向redis中插入chunk任务
	if err := task.PushTaskQueue(ctx, ktsk.TaskType); err != nil {
		logs.ErrorContextf(ctx, "PushTaskQueue error %v", err)
		return err
	}
	return nil
}

// -------------------- 公共生成函数 --------------------

func buildTaskPayload(file *foresttype.KnownowForestFile, commonPayload *task.CommonPayload, bucket string) *ragtask.TaskPayload {
	return &ragtask.TaskPayload{
		CommonPayload: *commonPayload,
		FileID:        file.ID,
		SubjectID:     file.ID,
		CompanyID:     file.CompanyID,
		Bucket:        bucket,
		ForestID:      file.ForestID,
		Uin:           file.Uin,
		FileExt:       file.PriviewExt,
		FileName:      file.Name,
		Filename:      file.Name,
	}
}

func buildTask(file *foresttype.KnownowForestFile, taskType string, payload *ragtask.TaskPayload, comment string) *task.Task {
	return &task.Task{
		Uin:               file.Uin,
		CompanyID:         file.CompanyID,
		TaskType:          taskType,
		TaskStatus:        task.TaskStatusPending,
		SubjectID:         file.ID,
		Comment:           comment,
		Payload:           payload.String(),
		TaskConfigTimeout: TaskTimeout,
		AppGroup:          AppGroup,
		TaskConfigRedo:    TaskRedo,
		Priority:          TaskPriority,
	}
}

// -------------------- 辅助函数 --------------------

func loadChatModel(ctx context.Context) chattype.ChatModel {
	scfg, _ := systemconfig.GetSystemConfig()
	model := chattype.ChatModel{}
	if scfg != nil {
		if m, err := chatmodel.GetModelByID(ctx, scfg.AlgoModelID); err == nil {
			model = *m
		}
	}
	return model
}
