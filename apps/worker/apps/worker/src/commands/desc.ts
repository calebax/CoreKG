import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { descHandler } from "@corekg/workers";
import { buildLocalContext } from "./build-context.js";

const logger = createLogger("cmd-desc");

export function createDescCommand(): Command {
  return new Command("desc")
    .description("Run desc task locally")
    .requiredOption("--file-url <url>", "File URL to download")
    .option("--file-id <id>", "File ID", "0")
    .option("--forest-id <id>", "Forest ID", "0")
    .option("--uin <uin>", "User ID", "0")
    .option("--company-id <id>", "Company ID", "0")
    .option("--es-index <index>", "Elasticsearch index name")
    .action(async (opts) => {
      try {
        const ctx = buildLocalContext();
        const result = await descHandler(ctx, {
          task_type: "desc",
          file_id: opts.fileId,
          file_url: opts.fileUrl,
          company_id: opts.companyId,
          forest_id: opts.forestId,
          uin: opts.uin,
          es_index: opts.esIndex,
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
