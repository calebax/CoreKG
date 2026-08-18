import { Command } from "commander";
import { createLogger } from "@corekg/logger";
import { createChunkCommand } from "./commands/chunk.js";
import { createExtractCommand } from "./commands/extract.js";
import { createAnalysisCommand } from "./commands/analysis.js";
import { createCopyCommand } from "./commands/copy.js";
import { createDescCommand } from "./commands/desc.js";
import { createMindmapCommand } from "./commands/mindmap.js";
import { createPdfExtractCommand } from "./commands/pdf-extract.js";
import { createVideoExtractCommand } from "./commands/video-extract.js";
import { createSplitCommand } from "./commands/split.js";
import { createIndexCmdCommand } from "./commands/index-cmd.js";

const logger = createLogger("kealgo");

export { main } from "./worker-main.js";

const program = new Command("kealgo")
  .description("kealgo worker and local CLI tools for document processing")
  .action(async () => {
    const mod = await import("./worker-main.js");
    await mod.main();
  });

program.addCommand(createChunkCommand());
program.addCommand(createExtractCommand());
program.addCommand(createAnalysisCommand());
program.addCommand(createCopyCommand());
program.addCommand(createDescCommand());
program.addCommand(createMindmapCommand());
program.addCommand(createPdfExtractCommand());
program.addCommand(createVideoExtractCommand());
program.addCommand(createSplitCommand());
program.addCommand(createIndexCmdCommand());

program.parseAsync(process.argv).catch((err) => {
  logger.error(err, "fatal");
  process.exit(1);
});
