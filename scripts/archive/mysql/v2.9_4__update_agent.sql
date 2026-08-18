SET NAMES utf8mb4;
START TRANSACTION;

SET @continuation_prompt = 
'### 任务：续写

从原文 <Original> 的最后一句（或最后一段）的结尾处直接自然接续写作，确保过渡无缝、无任何断层感。

#### 严格遵守以下要求：
- 不得重复、复述、总结或改写原文中已有的任何句子、段落或内容。
- **续写内容的输出语言必须与原文完全一致（包括使用的自然语言、语言变体与书面/口语层级），不得自行切换或混用其他语言。**
- 必须完全保持原文的语气、风格、语言特点、叙述视角（第一人称 / 第三人称 / 全知等）、行文节奏、专业术语使用、句式习惯及整体文体特征一致。
- 必须严格遵循原文已建立的世界观、逻辑设定、人物关系、时间线及所有事实细节，不得引入与原文冲突或突兀的新元素。
- 只输出续写产生的新内容，不包含任何引言、说明、标题、注释、总结、任务描述、分隔符或元信息。
- 续写一小段自然合理的后续内容：
  - 若为知识性文章，可适度深化论点或补充论据；
  - 若为叙事性文章，可轻微推进情节、人物行动或心理活动；
  - 不得强行完结故事或主题。
- 若原文以悬念、开放性问题或未尽论述结束，续写需在逻辑上自然延续发展，保持原文的严谨性与连贯性。
- 语言表达需与原文水准完全匹配，避免出现与原文时代、领域或语体不符的词汇、句式或网络用语。

{{input3}}

### 输入内容

原文：
<Original>
{{input1}}
</Original>

参考来源素材（可选，仅用于隐性风格借鉴）：
<Reference>
{{input2}}
</Reference>

---

直接开始续写，正文必须紧接原文结尾，无空行、无分隔符、无任何多余符号。
';

-- 获取 chat_agent.id
SELECT id INTO @agent_id FROM chat_agent WHERE name = 'sys_agent_write_continuation';

INSERT INTO `chat_agent_version` (
    created_at, updated_at, deleted_at,
    agent_id, description, chat_model_ids, temperature, agent_type,
    prompt_template, 
    greeting_message, 
    params, 
    forest_option
)
VALUES ('2026-01-04 11:27:07.659', '2026-01-04 11:27:07.659', NULL, 
        @agent_id,'写作空间续写命令', '[1]', '0.5','prompt',
        @continuation_prompt,
        '',
        '[{\"input\":\"input1\",\"name\":\"原始文本\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true}]',
        '{\"prompt_template\":\"\",\"doc_forest_ids\":null}');

SELECT id INTO @agent_version_id FROM chat_agent_version WHERE agent_id = @agent_id ORDER BY id DESC LIMIT 1;
UPDATE `chat_agent` SET `version` = @agent_version_id WHERE id = @agent_id;

COMMIT;
