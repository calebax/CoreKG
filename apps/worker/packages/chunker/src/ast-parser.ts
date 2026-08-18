import MarkdownIt from "markdown-it";

const md = new MarkdownIt();

export function parseMarkdownTokens(content: string) {
  if (!content) return [];
  return md.parse(content, {});
}
