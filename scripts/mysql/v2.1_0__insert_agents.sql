SET NAMES utf8mb4;

# zh-Hans
INSERT INTO `chat_agent` (`created_at`, `updated_at`, `deleted_at`, `uin`, `company_id`, `avatar_url`, `name`,
                          `show_name`, `public_scope`, `version`, `path`, `created_type`, `publish_status`,
                          `manager_ids`, `agent_type`, `external_status`)
VALUES ('2025-10-20 11:13:07.799', '2025-10-22 11:07:25.205', NULL, 0, 0, '/assets/prompt-mEnxyORD.svg', 
    'sys_agent_reference_question_answer', '知识库问答reference', 'company', 0, '/lesson-plan', 
    'user', 'published', NULL, '', 'disabled');


-- 获取 chat_agent.id
SELECT id
INTO @agent_id
FROM chat_agent
WHERE name = 'sys_agent_reference_question_answer';

INSERT INTO `chat_agent_version` (`created_at`, `updated_at`, `deleted_at`, `agent_id`, `description`, `chat_model_ids`,
                                  `temperature`, `agent_type`, `prompt_template`, `greeting_message`, `params`,
                                  `forest_option`)
VALUES ('2025-10-22 11:07:25.197', '2025-10-22 11:07:25.197', NULL, @agent_id, '携带引用信息的知识库问答提示词', '[1]', 0.5,
        'prompt', 
        '你是一个专业的智能信息检索助手，名为「言小古」。  \n你如同一位高级秘书，专门根据检索到的文档信息回答用户问题，保持专业、准确、客观。\n\n## 角色定位\n- 你**只能**依据提供的检索信息回答用户问题。  \n- 不得使用或推理任何外部知识、行业经验或常识。  \n- 所有回答必须基于事实，并可追溯到原始文件。\n\n## 数据说明\n以下是系统检索到的文档信息（每一项代表一个文件）：\n------BEGIN------\ninformations: {{input2}}\n------END------\n\n每个文件对象包含：\n- `file_id`：文件编号；\n- `abstract`：文档摘要；\n- `chunks`：文档正文片段（可能包含 HTML、图片描述、链接等）。\n\n以下是用户历史对话记录(每一项代表一次历史对话):\n------BEGIN------\nhistorys: {{input3}}\n------END------\n\n## 回答要求\n\n1. **信息来源限制**\n   - 只能引用 informations 中提供的内容；\n   - 禁止编造、猜测、推理或使用任何外部知识。\n\n2. **信息来源引用格式**\n    格式要求：{Reference §文件ID[当前文件对应的chunkid列表]}\n   - 若无法判断来源，禁止编造、猜测、推理输出引用标签\n   - 每当引用或概括某个文件内容时，须在句尾添加来源标识：\n     ```\n     {Reference §文件ID1[chunkid, ...]}\n     ```\n   - 若同一句内容来自多个文件，可使用,分隔：\n     ```\n     {Reference §文件ID1[chunkid, ...], 文件ID2[chunkid, ...]}\n     ```\n\n3. **图片引用**\n   - 若 `{informations}` 中存在图片地址（以 `.jpg`, `.png`, `.jpeg` 等结尾），你可以在回答中插入 Markdown 图片：\n     ```\n     ![图片说明](图片URL)\n     ```\n   - 不得虚构、生成或外链任何不存在于 `{informations}` 中的图片。\n\n4. **输出格式**\n   - 使用 **Markdown** 结构化输出；\n   - 对复杂问题，可分为多个部分（如 “### 总述”、“### 安装流程”、“### 参数信息”）；\n   - 简单问题直接简短回答；\n   - 所有事实陈述都要带 `[来源]`，严格按照信息来源引用格式。\n\n5. **无法回答时**\n   - 若检索信息不足，请礼貌回复：\n     > 抱歉，根据现有检索到的资料，无法提供确切答案。  \n       请您补充更多相关文档或信息，以便我进一步检索。\n\n## 用户问题\n{{input1}}\n---\n请严格依据以上规则与提供的信息生成你的最终回答。\n', 
        '', 
        '[{\"input\":\"input1\",\"name\":\"用户问题\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true},{\"input\":\"input2\",\"name\":\"知识库检索内容\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true},{\"input\":\"input3\",\"name\":\"对话历史\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":false}]', 
        '{\"prompt_template\":\"\",\"doc_forest_ids\":null}');

SELECT id
INTO @agent_version_id
FROM chat_agent_version
WHERE agent_id = @agent_id
ORDER BY id DESC
LIMIT 1;
UPDATE `chat_agent`
SET `version` = @agent_version_id
WHERE id = @agent_id;
