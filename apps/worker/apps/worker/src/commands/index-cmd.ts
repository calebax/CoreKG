import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { insertIndexHandler } from "@corekg/workers";
import { buildLocalContext } from "./build-context.js";

const logger = createLogger("cmd-index-cmd");

export function createIndexCmdCommand(): Command {
  return new Command("index-cmd")
    .description("Run insert index task locally")
    .requiredOption("--file-url <url>", "File URL")
    .option("--uin <uin>", "User ID", "0")
    .option("--company-id <id>", "Company ID", "0")
    .option("--forest-id <id>", "Forest ID", "0")
    .option("--file-id <id>", "File ID", "0")
    .option("--es-index <index>", "Elasticsearch index name")
    .option("--algo-url <url>", "Algo service URL")
    .action(async (opts) => {
      try {
        const ctx = buildLocalContext({ algoUrl: opts.algoUrl });
        const result = await insertIndexHandler(ctx, {
          task_type: "insert_index",
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
