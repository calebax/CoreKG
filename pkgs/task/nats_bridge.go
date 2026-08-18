package task

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

// Stream 名称
const (
	// DispatchStreamName 任务分发流，workqueue 策略保证每个任务只被一个 worker 消费
	DispatchStreamName = "CORE_TASK_DISPATCH"
	// ResultStreamName 任务结果流，limits 策略允许多个业务方订阅消费
	ResultStreamName = "CORE_TASK_RESULT"
)

// Subject 前缀与通配符
const (
	// DispatchSubjectPrefix 任务分发 subject 前缀，格式: core.task.dispatch.<short_name>
	DispatchSubjectPrefix = "core.task.dispatch."
	// DispatchSubjectWildcard 任务分发流通配符，匹配所有分发消息
	DispatchSubjectWildcard = "core.task.dispatch.*"
	// ResultSubjectPrefix 任务结果 subject 前缀，格式: core.task.result.<short_name>
	ResultSubjectPrefix = "core.task.result."
	// ResultSubjectWildcard 任务结果流通配符，匹配所有结果消息
	ResultSubjectWildcard = "core.task.result.*"
	// ResultSubscribeAll 业务方订阅所有结果消息的通配符
	ResultSubscribeAll = "core.task.result.>"
)

// 各任务类型的 dispatch subject（常量化，避免拼接出错）
const (
	DispatchCopy          = "core.task.dispatch.copy"           // 文件拷贝 / doc转pdf
	DispatchPDFExtract    = "core.task.dispatch.pdf_extract"    // PDF 解析为 markdown
	DispatchVideoExtract  = "core.task.dispatch.video_extract"  // 视频解析
	DispatchMindmap       = "core.task.dispatch.mindmap"        // 思维导图生成
	DispatchAnalysis      = "core.task.dispatch.analysis"       // 智能分析
	DispatchDesc          = "core.task.dispatch.desc"           // 文件描述生成
	DispatchSplitChunk    = "core.task.dispatch.split_text_chunk" // 文本分块+嵌入+索引
	DispatchInsertIndex   = "core.task.dispatch.insert_index"   // 插入 ES 索引
)

// 各任务类型的 result subject（常量化）
const (
	ResultCopy          = "core.task.result.copy"
	ResultPDFExtract    = "core.task.result.pdf_extract"
	ResultVideoExtract  = "core.task.result.video_extract"
	ResultMindmap       = "core.task.result.mindmap"
	ResultAnalysis      = "core.task.result.analysis"
	ResultDesc          = "core.task.result.desc"
	ResultSplitChunk    = "core.task.result.split_text_chunk"
	ResultInsertIndex   = "core.task.result.insert_index"
)

// taskTypeToShort 内部 task type 到 NATS subject short name 的映射
// dispatch 和 result subject 均通过此映射动态构建
var taskTypeToShort = map[string]string{
	"ke.copy_task":        "copy",
	"ke.doc_to_pdf_task":  "copy",
	"ke.prase_pdf_task":   "pdf_extract",
	"ke.prase_video_task": "video_extract",
	"ke.mind_map_task":    "mindmap",
	"ke.analysis_task":    "analysis",
	"ke.description_task": "desc",
	"ke.knowledge_task":   "split_text_chunk",
	"ke.insert_index":     "insert_index",
}

// NATSBridge 封装 NATS 连接和 JetStream 上下文
type NATSBridge struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// NewNATSBridge 创建 NATS Bridge 实例
func NewNATSBridge(nc *nats.Conn) *NATSBridge {
	if nc == nil {
		return nil
	}
	js, err := nc.JetStream()
	if err != nil {
		logs.Errorf("create JetStream context failed: %v", err)
		return nil
	}
	return &NATSBridge{nc: nc, js: js}
}

// DispatchSubject 根据 task type 返回对应的 dispatch subject
func DispatchSubject(taskType string) (string, error) {
	short, ok := taskTypeToShort[taskType]
	if !ok {
		return "", fmt.Errorf("no dispatch mapping for task type: %s", taskType)
	}
	return DispatchSubjectPrefix + short, nil
}

// ResultSubject 根据 task type 返回对应的 result subject
func ResultSubject(taskType string) (string, error) {
	short, ok := taskTypeToShort[taskType]
	if !ok {
		return "", fmt.Errorf("no result mapping for task type: %s", taskType)
	}
	return ResultSubjectPrefix + short, nil
}

func (b *NATSBridge) PublishTaskRPC(taskType string, payload []byte) error {
	if b == nil || b.js == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	subject, err := DispatchSubject(taskType)
	if err != nil {
		return err
	}
	_, err = b.js.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("nats dispatch publish %s: %w", subject, err)
	}
	return nil
}

func EnsureDispatchStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo(DispatchStreamName)
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("check dispatch stream: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      DispatchStreamName,
		Subjects:  []string{DispatchSubjectWildcard},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
		MaxMsgs:   100000,
		MaxBytes:  256 * 1024 * 1024,
	})
	if err != nil {
		return fmt.Errorf("create dispatch stream: %w", err)
	}
	logs.Infof("created NATS JetStream stream: %s", DispatchStreamName)
	return nil
}

func EnsureResultStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo(ResultStreamName)
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return fmt.Errorf("check result stream: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      ResultStreamName,
		Subjects:  []string{ResultSubjectWildcard},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    24 * time.Hour,
		MaxMsgs:   1000000,
		MaxBytes:  1024 * 1024 * 1024,
	})
	if err != nil {
		return fmt.Errorf("create result stream: %w", err)
	}
	logs.Infof("created NATS JetStream stream: %s", ResultStreamName)
	return nil
}

func (b *NATSBridge) EnsureStreams() error {
	if b == nil || b.js == nil {
		return fmt.Errorf("nats bridge not initialized")
	}
	if err := EnsureDispatchStream(b.js); err != nil {
		return err
	}
	return EnsureResultStream(b.js)
}
