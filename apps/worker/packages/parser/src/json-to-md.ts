interface ContentItem {
  type?: string;
  text?: string;
  text_level?: number;
  page_idx?: number;
  bbox?: number[];
  img_path?: string;
  image_caption?: unknown[];
  image_footnote?: unknown[];
  table_caption?: unknown[];
  table_body?: string;
  table_footnote?: unknown[];
  chart_caption?: unknown[];
  content?: string;
  chart_footnote?: unknown[];
  code_caption?: unknown[];
  code_body?: string;
  code_footnote?: unknown[];
  sub_type?: string;
  list_items?: unknown[];
  [key: string]: unknown;
}

export function contentListToMarkdown(contentList: ContentItem[], imageDir = "images"): string {
  const lines: string[] = [];

  for (const item of contentList) {
    if (!item) continue;
    try {
      const md = convertItem(item, imageDir);
      if (md && md.trim()) lines.push(md);
    } catch {
      continue;
    }
  }

  return lines.join("\n\n");
}

function getFirstListItem(lst: unknown, fallback = ""): string {
  if (Array.isArray(lst) && lst.length > 0 && lst[0]) return String(lst[0]).trim();
  return fallback;
}

function getListAsStrings(lst: unknown): string[] {
  if (!lst) return [];
  if (!Array.isArray(lst)) return [String(lst).trim()];
  return lst.filter((i) => i && String(i).trim()).map((i) => String(i).trim());
}

function posMarker(item: ContentItem): string {
  const pageIdx = item.page_idx ?? 0;
  const bbox = item.bbox ?? [0, 0, 0, 0];
  return ` \n<!--yg_pos${pageIdx + 1},${bbox[0]},${bbox[1]},${bbox[2]},${bbox[3]}yg_pos-->`;
}

function buildImageUrl(imgPath: string, imageDir: string): string {
  if (!imgPath) return "";
  const p = String(imgPath).trim();
  if (p.includes("/") || p.includes("\\") || p.startsWith(".")) return p;
  return `${imageDir}/${p}`;
}

function convertItem(item: ContentItem, imageDir: string): string {
  const t = item.type;
  if (t === "text") return convertText(item);
  if (t === "image") return convertImage(item, imageDir);
  if (t === "table") return convertTable(item, imageDir);
  if (t === "chart") return convertChart(item, imageDir);
  if (t === "equation") return convertEquation(item, imageDir);
  if (t === "code") return convertCode(item);
  if (t === "list") return convertList(item);
  if (t === "seal") return convertSeal(item, imageDir);
  return "";
}

function convertText(item: ContentItem): string {
  const text = (item.text ?? "").trim();
  if (!text) return "";
  const level = item.text_level ?? 0;
  const pm = posMarker(item);
  if (level > 0) return `${"#".repeat(level)} ${text}${pm}`;
  return `${text}${pm}`;
}

function convertImage(item: ContentItem, imageDir: string): string {
  const parts: string[] = [];
  const pm = posMarker(item);
  const imgPath = item.img_path ?? "";
  if (imgPath) {
    const caption = getFirstListItem(item.image_caption, "");
    parts.push(`![${caption}](${buildImageUrl(imgPath, imageDir)})${pm}`);
  }
  for (const cap of getListAsStrings(item.image_caption)) {
    if (cap) parts.push(cap);
  }
  for (const note of getListAsStrings(item.image_footnote)) {
    if (note) parts.push(`> ${note}`);
  }
  return parts.join("\n");
}

function convertTable(item: ContentItem, imageDir: string): string {
  const parts: string[] = [];
  const pm = posMarker(item);
  for (const cap of getListAsStrings(item.table_caption)) {
    if (cap) parts.push(`**${cap}**`);
  }
  let tableHtml = item.table_body ?? "";
  if (!tableHtml) {
    for (const key of Object.keys(item)) {
      if (key.toLowerCase().includes("body") && typeof item[key] === "string") {
        tableHtml = item[key] as string;
        break;
      }
    }
  }
  if (tableHtml) {
    parts.push(`${tableHtml}${pm}`);
  } else {
    const imgPath = item.img_path ?? "";
    if (imgPath) parts.push(`![Table](${buildImageUrl(imgPath, imageDir)})${pm}`);
  }
  for (const note of getListAsStrings(item.table_footnote)) {
    if (note) parts.push(`> ${note}`);
  }
  return parts.join("\n");
}

function convertChart(item: ContentItem, imageDir: string): string {
  const parts: string[] = [];
  const pm = posMarker(item);
  for (const cap of getListAsStrings(item.chart_caption)) {
    if (cap) parts.push(`**${cap}**`);
  }
  const imgPath = item.img_path ?? "";
  if (imgPath) parts.push(`![Chart](${buildImageUrl(imgPath, imageDir)})${pm}`);
  const content = item.content;
  if (content && String(content).trim()) parts.push(String(content).trim());
  for (const note of getListAsStrings(item.chart_footnote)) {
    if (note) parts.push(`> ${note}`);
  }
  return parts.join("\n");
}

function convertEquation(item: ContentItem, imageDir: string): string {
  const parts: string[] = [];
  const pm = posMarker(item);
  const text = (item.text ?? "").trim();
  if (text) {
    if (text.startsWith("$$")) {
      parts.push(`${text}${pm}`);
    } else {
      parts.push(`$$\n${text}\n$$${pm}`);
    }
  } else {
    const imgPath = item.img_path ?? "";
    if (imgPath) parts.push(`![Equation](${buildImageUrl(imgPath, imageDir)})${pm}`);
  }
  return parts.join("\n");
}

function convertCode(item: ContentItem): string {
  const parts: string[] = [];
  const pm = posMarker(item);
  const subType = item.sub_type ?? "code";
  for (const cap of getListAsStrings(item.code_caption)) {
    if (cap) {
      parts.push(subType === "algorithm" ? `**算法：${cap}**` : `**${cap}**`);
    }
  }
  const codeBody = (item.code_body ?? "").trim();
  if (codeBody) {
    const lang = subType === "algorithm" ? "text" : "python";
    parts.push(`\`\`\`${lang}\n${codeBody}\n\`\`\`${pm}`);
  }
  for (const note of getListAsStrings(item.code_footnote)) {
    if (note) parts.push(`> ${note}`);
  }
  return parts.join("\n");
}

function convertList(item: ContentItem): string {
  const listItems = item.list_items ?? [];
  if (!listItems.length) return "";
  const pm = posMarker(item);
  const parts: string[] = [];
  for (const li of listItems) {
    const s = li ? String(li).trim() : "";
    if (s) parts.push(`- ${s}`);
  }
  let result = parts.join("\n");
  if (result) result += pm;
  return result;
}

function convertSeal(item: ContentItem, imageDir: string): string {
  const parts: string[] = [];
  const pm = posMarker(item);
  const imgPath = item.img_path ?? "";
  if (imgPath) parts.push(`![Seal](${buildImageUrl(imgPath, imageDir)})${pm}`);
  const text = (item.text ?? "").trim();
  if (text) parts.push(`*${text}*`);
  return parts.join("\n");
}
