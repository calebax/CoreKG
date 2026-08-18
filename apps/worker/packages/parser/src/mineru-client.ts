import { readFile } from "node:fs/promises";
import { basename, resolve } from "node:path";
import mime from "mime-types";
import AdmZip from "adm-zip";

export interface MinerUResult {
  contentList: unknown[];
  outputDir: string;
}

export interface MinerUClientOptions {
  apiUrl: string;
}

export async function processPdfWithMinerU(
  localFilePath: string,
  outputDir: string,
  options: MinerUClientOptions,
): Promise<MinerUResult> {
  const fileData = await readFile(localFilePath);
  const filename = basename(localFilePath);
  const mimeType = mime.lookup(localFilePath) || "application/octet-stream";

  const form = new FormData();
  form.append("files", new Blob([fileData], { type: mimeType }), filename);
  form.append("return_md", "true");
  form.append("response_format_zip", "true");
  form.append("return_original_file", "true");
  form.append("return_images", "true");
  form.append("return_content_list", "true");
  form.append("return_middle_json", "true");

  const resp = await fetch(options.apiUrl, {
    method: "POST",
    body: form,
  });

  if (!resp.ok) throw new Error(`MinerU API error: ${resp.status} ${resp.statusText}`);

  const zipBuf = Buffer.from(await resp.arrayBuffer());
  const zip = new AdmZip(zipBuf);
  zip.extractAllTo(outputDir, true);

  const contentListPath = findContentListJson(outputDir, zip);
  const rawData = await readFile(contentListPath, "utf-8");
  const contentList: unknown[] = JSON.parse(rawData);

  return { contentList, outputDir };
}

function findContentListJson(outputDir: string, zip: AdmZip): string {
  const entries = zip.getEntries();
  const clEntry = entries.find((e) => e.entryName.endsWith("content_list.json"));
  if (clEntry) return resolve(outputDir, clEntry.entryName);
  const jsonEntry = entries.find((e) => e.entryName.endsWith(".json"));
  if (jsonEntry) return resolve(outputDir, jsonEntry.entryName);
  throw new Error("No content_list.json found in MinerU response");
}
