/**
 * NATS Subject 常量
 *
 * 任务分发 (dispatch): Go → Worker
 *   Stream: CORE_TASK_DISPATCH (workqueue, 每个消息只投递给一个 consumer)
 *   Subject 格式: core.task.dispatch.<short_name>
 *
 * 任务结果 (result): Worker → 业务方
 *   Stream: CORE_TASK_RESULT (limits, 保留24h, 可被多个业务方订阅)
 *   Subject 格式: core.task.result.<short_name>
 *
 * short_name 与 Go 侧 pkgs/task/nats_bridge.go 中的 taskTypeToShort 映射一致
 */

// 任务分发 subject: Go 通过 JetStream publish 到这些 subject, Worker PullSubscribe 消费
export const DISPATCH_SUBJECTS = {
  /** 文件拷贝 / doc转pdf (ke.copy_task, ke.doc_to_pdf_task) */
  copy: "core.task.dispatch.copy",
  /** PDF 解析为 markdown (ke.prase_pdf_task) */
  pdfExtract: "core.task.dispatch.pdf_extract",
  /** 视频解析 (ke.prase_video_task) */
  videoExtract: "core.task.dispatch.video_extract",
  /** 思维导图生成 (ke.mind_map_task) */
  mindmap: "core.task.dispatch.mindmap",
  /** 智能分析 (ke.analysis_task) */
  analysis: "core.task.dispatch.analysis",
  /** 文件描述生成 (ke.description_task) */
  desc: "core.task.dispatch.desc",
  /** 文本分块 + 嵌入 + 索引 (ke.knowledge_task) */
  splitTextChunk: "core.task.dispatch.split_text_chunk",
  /** 插入 ES 索引 (ke.insert_index) */
  insertIndex: "core.task.dispatch.insert_index",
} as const;

// 任务结果 subject: Worker 完成后 publish 到这些 subject, 业务方订阅消费
export const RESULT_SUBJECTS = {
  /** 文件拷贝结果 */
  copy: "core.task.result.copy",
  /** PDF 解析结果 */
  pdfExtract: "core.task.result.pdf_extract",
  /** 视频解析结果 */
  videoExtract: "core.task.result.video_extract",
  /** 思维导图结果 */
  mindmap: "core.task.result.mindmap",
  /** 智能分析结果 */
  analysis: "core.task.result.analysis",
  /** 文件描述结果 */
  desc: "core.task.result.desc",
  /** 分块索引结果 */
  splitTextChunk: "core.task.result.split_text_chunk",
  /** 插入索引结果 */
  insertIndex: "core.task.result.insert_index",
} as const;

// Stream 名称 (与 Go 侧 DispatchStreamName / ResultStreamName 一致)
export const STREAM_NAMES = {
  /** 任务分发流, workqueue 策略 */
  dispatch: "CORE_TASK_DISPATCH",
  /** 任务结果流, limits 策略, 24h 保留 */
  result: "CORE_TASK_RESULT",
} as const;

// Subject 通配符
export const SUBJECT_WILDCARDS = {
  /** 匹配所有分发消息 */
  dispatchAll: "core.task.dispatch.*",
  /** 匹配所有结果消息 */
  resultAll: "core.task.result.*",
  /** 订阅所有结果消息 (用于 JetStream subscribe) */
  resultSubscribeAll: "core.task.result.>",
} as const;

export type DispatchSubjectKey = keyof typeof DISPATCH_SUBJECTS;
export type ResultSubjectKey = keyof typeof RESULT_SUBJECTS;
