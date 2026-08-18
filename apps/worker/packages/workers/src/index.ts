export type {
  TaskContext,
  TaskHandlerResult,
  TaskHandlerFn,
  TaskHandlerDef,
  TaskResultMessage,
  AgentClientConfig,
} from "./types.js";
export { agentRequest, splitMerge } from "./agent-client.js";
export {
  extractMarkdownTitles,
  extractCodeBlock,
  processEmbeddedUuid,
} from "./markdown-utils.js";
export { daemonProcessPdf, daemonProcessVideo } from "./daemon-client.js";
export type { DaemonPdfOptions, DaemonVideoOptions } from "./daemon-client.js";
export { algoSplit, algoIndex } from "./algo-client.js";
export type { AlgoSplitOptions, AlgoIndexOptions } from "./algo-client.js";
export { handlerRegistry } from "./registry.js";
export { analysisHandler } from "./handlers/analysis.js";
export { copyHandler } from "./handlers/copy.js";
export { descHandler } from "./handlers/desc.js";
export { mindmapHandler } from "./handlers/mindmap.js";
export { pdfExtractHandler } from "./handlers/pdf-extract.js";
export { videoExtractHandler } from "./handlers/video-extract.js";
export { splitTextChunkHandler } from "./handlers/split-text-chunk.js";
export { insertIndexHandler } from "./handlers/insert-index.js";
