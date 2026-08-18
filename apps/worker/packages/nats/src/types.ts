import { z } from "zod";

export const LLMModelConfigSchema = z.object({
  api_key: z.string(),
  base_url: z.string(),
  model_name: z.string(),
  provider: z.string().optional(),
});

export const SplitConfigSchema = z.object({
  split_mode: z.string().default("smart"),
  chunk_size: z.number().optional(),
  chunk_token_num: z.number().optional(),
  split_mark: z.union([z.string(), z.array(z.string())]).optional(),
  regex_pattern: z.string().optional(),
  split_overlap: z.number().optional(),
  overlap_ratio: z.number().optional(),
  min_chunk_tokens: z.number().default(10),
  split_level: z.number().default(2),
  enable_heading_in_content: z.boolean().default(false),
  preprocessing_rules: z.object({
    remove_email: z.boolean().default(true),
    remove_url: z.boolean().default(true),
    remove_empty_line: z.boolean().default(true),
  }).optional(),
  llm_enabled: z.boolean().optional(),
  llm_model: z.string().optional(),
  vllm_enabled: z.boolean().optional(),
  vllm_model: z.string().optional(),
  image_width: z.number().optional(),
  eb_max_concurrency: z.number().optional(),
  llm_max_concurrency: z.number().optional(),
});

export const TaskPayloadSchema = z.object({
  task_id: z.union([z.string(), z.number()]).optional(),
  task_type: z.string(),
  file_id: z.union([z.string(), z.number()]),
  file_url: z.string(),
  filename: z.string().optional(),
  file_name: z.string().optional(),
  file_ext: z.string().optional(),
  company_id: z.union([z.string(), z.number()]),
  forest_id: z.union([z.string(), z.number()]),
  uin: z.union([z.string(), z.number()]),
  es_index: z.string().optional(),
  bucket: z.string().optional(),
  upload_path: z.string().optional(),
  storage_path: z.string().optional(),
  llm: LLMModelConfigSchema.optional(),
  vllm: LLMModelConfigSchema.optional(),
  embedding: LLMModelConfigSchema.optional(),
  split_config: SplitConfigSchema.optional(),
  es: z.object({
    index_name: z.string(),
    addr: z.string(),
    username: z.string(),
    password: z.string(),
  }).optional(),
});

export type TaskPayload = z.infer<typeof TaskPayloadSchema>;


