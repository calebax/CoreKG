import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { mindmapHandler } from "@corekg/workers";
import { buildLocalContext } from "./build-context.js";

const logger = createLogger("cmd-mindmap");

export function createMindmapCommand(): Command {
  return new Command("mindmap")
    .description("Run mindmap task locally")
    .requiredOption("--file-url <url>", "File URL to download")
    .requiredOption("--upload-path <path>", "S3 upload path")
    .requiredOption("--bucket <bucket>", "S3 bucket name")
    .action(async (opts) => {
      try {
        const ctx = buildLocalContext();
        const result = await mindmapHandler(ctx, {
          task_type: "mindmap",
          file_id: "0",
          file_url: opts.fileUrl,
          company_id: "0",
          forest_id: "0",
          uin: "0",
          upload_path: opts.uploadPath,
          bucket: opts.bucket,
        });
        if (result.status === "success") {
          console.log(JSON.stringify(result.result, null, 2));
        } else {
          console.error(result.error);
          process.exit(1);
        }
      } catch (err) {
        logger.error(err, "fatal");
        process.exit(1);
      }
    });
}
