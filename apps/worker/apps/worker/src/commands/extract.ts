import fs from "node:fs/promises";
import { resolve, extname } from "node:path";
import { Command } from "commander";
import { createLogger, withContext } from "@corekg/logger";
import {
  processPdfWithMinerU,
  contentListToMarkdown,
  handleTextFile,
  handleCSVFile,
  handleJSONFile,
  handleVideoFile,
} from "@corekg/parser";

const logger = createLogger("extract");

export function createExtractCommand(): Command {
  const cmd = new Command("extract")
    .description("Local CLI tool for document parsing (all formats)")
    .requiredOption("--file <path>", "local file path to parse")
    .option("--output <path>", "output directory or file path")
    .option("--mineru-url <url>", "MinerU API endpoint URL")
    .action(async (opts) => {
      try {
        await runExtract(opts);
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        logger.error({ err: msg }, "extract command failed");
        process.exit(1);
      }
    });
  return cmd;
}

interface ExtractOptions {
  file: string;
  output?: string;
  mineruUrl?: string;
}

const PDF_EXTS = new Set([".pdf"]);
const TEXT_EXTS = new Set([".txt", ".md", ".markdown"]);
const CSV_EXTS = new Set([".csv"]);
const JSON_EXTS = new Set([".json"]);
const VIDEO_EXTS = new Set([".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv"]);
const OFFICE_EXTS = new Set([".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx"]);

async function runExtract(opts: ExtractOptions) {
  const filePath = resolve(opts.file);
  const stat = await fs.stat(filePath).catch(() => null);
  if (!stat || !stat.isFile()) {
    throw new Error(`file not found: ${filePath}`);
  }

  const ext = extname(filePath).toLowerCase();
  const log = withContext(logger, { file: filePath, ext });
  let markdown: string;

  if (PDF_EXTS.has(ext) || OFFICE_EXTS.has(ext)) {
    const mineruUrl = opts.mineruUrl || process.env.MINERU_API_URL;
    if (!mineruUrl) {
      throw new Error("MinerU API URL required for PDF/Office files. Use --mineru-url or set MINERU_API_URL env var.");
    }
    const outputDir = opts.output ? resolve(opts.output) : `/tmp/extract-${Date.now()}`;
    await fs.mkdir(outputDir, { recursive: true });

    log.info({ mineruUrl }, "parsing with MinerU");
    const result = await processPdfWithMinerU(filePath, outputDir, { apiUrl: mineruUrl });
    markdown = contentListToMarkdown(result.contentList as Parameters<typeof contentListToMarkdown>[0]);

    log.info({ outputDir, items: result.contentList.length }, "MinerU parsing complete");
  } else if (TEXT_EXTS.has(ext)) {
    markdown = await handleTextFile(filePath);
  } else if (CSV_EXTS.has(ext)) {
    markdown = await handleCSVFile(filePath);
  } else if (JSON_EXTS.has(ext)) {
    markdown = await handleJSONFile(filePath);
  } else if (VIDEO_EXTS.has(ext)) {
    const outputDir = opts.output ? resolve(opts.output) : `/tmp/extract-${Date.now()}`;
    await fs.mkdir(outputDir, { recursive: true });

    log.info("extracting video frames");
    handleVideoFile(filePath, outputDir);
    markdown = `Video frames extracted to ${outputDir}`;
  } else {
    throw new Error(`unsupported file extension: ${ext}`);
  }

  if (opts.output && !isDir(opts.output)) {
    const outPath = resolve(opts.output);
    await fs.writeFile(outPath, markdown, "utf-8");
    log.info({ path: outPath }, "extract result written");
  } else if (!opts.output) {
    process.stdout.write(markdown + "\n");
  }
}

function isDir(p: string): boolean {
  return !extname(p) || p.endsWith("/");
}
