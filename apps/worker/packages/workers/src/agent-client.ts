import type { AgentClientConfig } from "./types.js";

export async function agentRequest(
  inputs: Record<string, string>,
  config: AgentClientConfig,
  model: string,
  ...backups: string[]
): Promise<string> {
  for (const value of Object.values(inputs)) {
    if (value.length > config.maxTokenSize && backups.length === 2) {
      return splitMerge(value, config, backups[0], backups[1]);
    }
  }

  const body = {
    model,
    chat_options: {
      input: Object.entries(inputs).map(([name, value]) => ({ name, value })),
    },
    stream: false,
  };

  for (let i = 0; i < 20; i++) {
    try {
      const resp = await fetch(config.apiUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${config.apiKey}`,
        },
        body: JSON.stringify(body),
      });
      if (!resp.ok) throw new Error(`Agent API error: ${resp.status}`);
      const data = (await resp.json()) as { choices?: Array<{ message?: { content?: string } }> };
      const content = data?.choices?.[0]?.message?.content;
      if (!content) throw new Error("No content in agent response");
      return content;
    } catch (err) {
      if (i === 19) throw err;
      await new Promise((r) => setTimeout(r, 1000));
    }
  }
  throw new Error("Agent request failed after 20 retries");
}

export async function splitMerge(
  content: string,
  config: AgentClientConfig,
  chunkModel: string,
  mergeModel: string,
): Promise<string> {
  const chunks = splitContent(content, config.chunkSize);
  const summaries = await mapPhase(chunks, config, chunkModel);
  return reducePhase(summaries, config, chunkModel, mergeModel);
}

function splitContent(content: string, chunkSize: number): string[] {
  if (content.length <= chunkSize) return [content];
  const paragraphs = content.split("\n");
  const chunks: string[] = [];
  let current = "";
  for (const para of paragraphs) {
    const newLen = current ? current.length + 1 + para.length : para.length;
    if (newLen <= chunkSize) {
      current = current ? current + "\n" + para : para;
    } else {
      if (current) chunks.push(current);
      current = para;
    }
  }
  if (current) chunks.push(current);
  return chunks;
}

async function mapPhase(
  chunks: string[],
  config: AgentClientConfig,
  chunkModel: string,
): Promise<string[]> {
  const results: string[] = new Array(chunks.length);
  let active = 0;

  const processChunk = async (idx: number): Promise<void> => {
    while (active >= config.maxWorkers) {
      await new Promise((r) => setTimeout(r, 10));
    }
    active++;
    try {
      const inputs = { input1: chunks[idx], input2: `第${idx + 1}块，共${chunks.length}块` };
      results[idx] = await agentRequest(inputs, config, chunkModel);
    } finally {
      active--;
    }
  };

  await Promise.all(chunks.map((_, idx) => processChunk(idx)));
  return results;
}

async function reducePhase(
  summaries: string[],
  config: AgentClientConfig,
  chunkModel: string,
  mergeModel: string,
): Promise<string> {
  const combined = summaries.join("\n---\n");
  const inputs = {
    input1: combined,
    input2: `共${summaries.length}个分块的总结`,
  };
  return agentRequest(inputs, config, mergeModel, chunkModel, mergeModel);
}
