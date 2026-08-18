import type { TaskHandlerDef } from "./types.js";
import { analysisHandler } from "./handlers/analysis.js";
import { copyHandler } from "./handlers/copy.js";
import { descHandler } from "./handlers/desc.js";
import { mindmapHandler } from "./handlers/mindmap.js";
import { pdfExtractHandler } from "./handlers/pdf-extract.js";
import { videoExtractHandler } from "./handlers/video-extract.js";
import { splitTextChunkHandler } from "./handlers/split-text-chunk.js";
import { insertIndexHandler } from "./handlers/insert-index.js";
import { DISPATCH_SUBJECTS, RESULT_SUBJECTS } from "@corekg/rpc";

export const handlerRegistry: TaskHandlerDef[] = [
  { name: "analysis", dispatchSubject: DISPATCH_SUBJECTS.analysis, resultSubject: RESULT_SUBJECTS.analysis, handler: analysisHandler },
  { name: "copy", dispatchSubject: DISPATCH_SUBJECTS.copy, resultSubject: RESULT_SUBJECTS.copy, handler: copyHandler },
  { name: "desc", dispatchSubject: DISPATCH_SUBJECTS.desc, resultSubject: RESULT_SUBJECTS.desc, handler: descHandler },
  { name: "mindmap", dispatchSubject: DISPATCH_SUBJECTS.mindmap, resultSubject: RESULT_SUBJECTS.mindmap, handler: mindmapHandler },
  { name: "pdf_extract", dispatchSubject: DISPATCH_SUBJECTS.pdfExtract, resultSubject: RESULT_SUBJECTS.pdfExtract, handler: pdfExtractHandler },
  { name: "video_extract", dispatchSubject: DISPATCH_SUBJECTS.videoExtract, resultSubject: RESULT_SUBJECTS.videoExtract, handler: videoExtractHandler },
  { name: "split_text_chunk", dispatchSubject: DISPATCH_SUBJECTS.splitTextChunk, resultSubject: RESULT_SUBJECTS.splitTextChunk, handler: splitTextChunkHandler },
  { name: "insert_index", dispatchSubject: DISPATCH_SUBJECTS.insertIndex, resultSubject: RESULT_SUBJECTS.insertIndex, handler: insertIndexHandler },
];
