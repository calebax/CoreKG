SET NAMES utf8mb4;

START TRANSACTION;

-- 先定位旧的 agent_id（若不存在则为 NULL）
SELECT id INTO @old_agent_id
FROM chat_agent
WHERE name = 'sys_agent_write_continuation'
LIMIT 1;

-- 先删子表（避免外键/引用问题）
DELETE FROM chat_agent_version
WHERE agent_id = @old_agent_id;

-- 再删主表
DELETE FROM chat_agent
WHERE id = @old_agent_id;

SET @continuation_prompt = 
'### 任务：续写

从原文 <Original> 的最后一句（或最后一段）的结尾处直接自然接续写作，确保过渡无缝、无任何断层感。

#### 严格遵守以下要求：
- 不得重复、复述、总结或改写原文中已有的任何句子、段落或内容。
- 必须完全保持原文的语气、风格、语言特点、叙述视角（第一人称 / 第三人称 / 全知等）、行文节奏、专业术语使用、句式习惯及整体文体特征一致。
- 必须严格遵循原文已建立的世界观、逻辑设定、人物关系、时间线及所有事实细节，不得引入与原文冲突或突兀的新元素。
- 只输出续写产生的新内容，不包含任何引言、说明、标题、注释、总结、任务描述、分隔符或元信息。
- 续写一小段自然合理的后续内容：
  - 若为知识性文章，可适度深化论点或补充论据；
  - 若为叙事性文章，可轻微推进情节、人物行动或心理活动；
  - 不得强行完结故事或主题。
- 若原文以悬念、开放性问题或未尽论述结束，续写需在逻辑上自然延续发展，保持原文的严谨性与连贯性。
- 语言表达需与原文水准完全匹配，避免出现与原文时代、领域或语体不符的词汇、句式或网络用语。

#### 参考来源素材的使用规则（如提供）：
- 若提供 <Reference>，该内容来自专业知识库，**默认用于隐性风格与表达方式借鉴，并在需要事实性支撑时作为可信来源之一。**
- 在不涉及事实性陈述时，可以从中吸收专业术语使用习惯、论证逻辑或行文结构特征，而不显性引用。
- **若提供了 <Reference>，且原文内容允许事实性或论证性延展，续写时应至少引入一处与原文逻辑一致的事实性补充、具体论据或明确结论，并按下述“引用适用规则”提供对应引用。**
- **严禁**直接复制、改写或大段引用参考素材原文；即便在需要引用时，也只能用于事实依据，不得进行内容性复写。

---

### 引用适用规则（在出现事实性陈述时必须严格遵守）

#### 何时允许引用：
- 仅用于 **关键事实、具体数据、明确结论或非公开的实质性陈述** 的依据说明。
- 对涉及数字、日期、明确论断等关键事实，必须记录来源，并优先使用高质量来源。

#### 何时不需要引用：
- 广泛接受的常识、通用解释、背景性描述无需引用。
- 不得为了形式完整而强行添加引用。

#### 引用粒度要求：
- 同一句话中，如多个信息来自同一来源，仅在句末引用一次。
- 避免为相邻句子或同一事实拆分多次引用，防止碎片化。

---

### 引用格式规范（技术要求）

**统一使用以下文本块格式：**
```\n{Reference §fileID[chunkSequence1, chunkSequence2, ...]}\n```
#### 示例说明：
- **单一来源：**
```\n{Reference §1234[16, 35, 108]}\n```
- **多来源合并引用：**
```\n{Reference §1234[16, 18], §4567[24]}\n```

#### 严禁行为（再次强调）：
- ❌ 严禁编造、猜测或虚构引用标签
- ❌ 严禁为通用知识、常识性内容添加引用

---

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


-- 写作空间[续写]
INSERT INTO `chat_agent` (
    created_at, updated_at, deleted_at,
    uin, company_id, avatar_url, name, show_name,
    public_scope, version, path, created_type, publish_status,
    manager_ids, agent_type, external_status
)
VALUES ('2026-01-04 11:27:07.659', '2026-01-04 11:27:07.659', NULL, 
        0, 0, '/assets/prompt-CEUUcXkn.png','sys_agent_write_continuation', '写作空间[续写]', 
        'company', '0', '/lesson-plan', 'user', 'published',
        NULL, 'prompt','disabled');

-- 获取 chat_agent.id
SELECT id
INTO @agent_id
FROM chat_agent
WHERE name = 'sys_agent_write_continuation' LIMIT 1;

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

SELECT id INTO @agent_version_id
FROM chat_agent_version
WHERE agent_id = @agent_id
ORDER BY id DESC LIMIT 1;

UPDATE chat_agent SET version = @agent_version_id WHERE id = @agent_id;

COMMIT;
