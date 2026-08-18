export interface PreprocessOptions {
  removeEmail: boolean;
  removeUrl: boolean;
  removeEmptyLine: boolean;
}

export function preprocessText(text: string, options: PreprocessOptions): string {
  let result = text;

  if (options.removeEmail) {
    result = result.replace(/[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g, "");
  }

  if (options.removeUrl) {
    result = result.replace(/(?<!!\[.*?\]\()https?:\/\/[^\s)]+/g, "");
  }

  if (options.removeEmptyLine) {
    result = result
      .split("\n")
      .map((line) => line.trimEnd())
      .join("\n")
      .replace(/\n{3,}/g, "\n\n");
  }

  return result;
}
