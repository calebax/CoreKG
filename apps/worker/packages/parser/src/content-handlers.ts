import { readFile } from "node:fs/promises";
import { execSync } from "node:child_process";
import { basename, join } from "node:path";

export async function handleTextFile(filePath: string): Promise<string> {
  return readFile(filePath, "utf-8");
}

export async function handleCSVFile(filePath: string): Promise<string> {
  const raw = await readFile(filePath, "utf-8");
  const rows = parseCSV(raw);
  if (!rows.length) return "";
  const headers = rows[0];
  const lines: string[] = [];
  lines.push(`| ${headers.join(" | ")} |`);
  lines.push(`| ${headers.map(() => "---").join(" | ")} |`);
  for (let i = 1; i < rows.length; i++) {
    lines.push(`| ${rows[i].join(" | ")} |`);
  }
  return lines.join("\n");
}

export async function handleJSONFile(filePath: string): Promise<string> {
  const raw = await readFile(filePath, "utf-8");
  const data = JSON.parse(raw);
  return jsonToMarkdown(data);
}

export function handleVideoFile(filePath: string, outputDir: string): string {
  const imagesDir = join(outputDir, "images");
  execSync(
    `mkdir -p "${imagesDir}" && ffmpeg -y -i "${filePath}" -vf "select=gt(scene,0.3)" -vsync vfr -q:v 2 "${imagesDir}/frame_%04d.jpg"`,
    { stdio: "pipe" },
  );
  return imagesDir;
}

function parseCSV(raw: string): string[][] {
  const lines = raw.split(/\r?\n/).filter((l) => l.trim());
  return lines.map((line) => {
    const result: string[] = [];
    let current = "";
    let inQuotes = false;
    for (let i = 0; i < line.length; i++) {
      const ch = line[i];
      if (inQuotes) {
        if (ch === '"') {
          if (i + 1 < line.length && line[i + 1] === '"') {
            current += '"';
            i++;
          } else {
            inQuotes = false;
          }
        } else {
          current += ch;
        }
      } else {
        if (ch === '"') {
          inQuotes = true;
        } else if (ch === ",") {
          result.push(current);
          current = "";
        } else {
          current += ch;
        }
      }
    }
    result.push(current);
    return result;
  });
}

function jsonToMarkdown(data: unknown): string {
  if (data !== null && typeof data === "object" && !Array.isArray(data)) {
    let md = "## JSON Data\n\n";
    for (const [key, value] of Object.entries(data)) {
      md += `### ${key}\n\n`;
      md += jsonToMarkdown(value) + "\n";
    }
    return md;
  }
  if (Array.isArray(data)) {
    return data.map((item) => jsonToMarkdown(item)).join("\n");
  }
  return `${data}\n`;
}
