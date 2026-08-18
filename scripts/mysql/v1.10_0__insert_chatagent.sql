SET NAMES utf8mb4;

-- 1. 先插入所有 agent（version 先置 NULL）
INSERT INTO `chat_agent` (
    created_at, updated_at, deleted_at,
    uin, company_id, avatar_url, name, show_name,
    public_scope, version, path, created_type, publish_status,
    manager_ids, agent_type, external_status
) VALUES
      ('2025-09-29 09:53:10.402', '2025-09-29 10:02:20.650', NULL, 330, 2, '/assets/prompt-CEUUcXkn.png', 
      'sys_agent_user_query_rewrite', '【知识库问答】用户问答语义补充', 'company', 1, '/lesson-plan', 'user', 'published', NULL, '', 'disabled'),
      ('2025-09-29 10:14:04.351', '2025-09-29 10:18:01.441', NULL, 330, 2, '/assets/prompt-CEUUcXkn.png', 
      'sys_agent_question_answer', '新版es问答', 'company', 1, '/lesson-plan', 'user', 'published', NULL, '', 'disabled');


-- ========== sys_agent_user_query_rewrite ==========
SELECT id INTO @agent_id FROM chat_agent WHERE name='sys_agent_user_query_rewrite' LIMIT 1;

INSERT INTO chat_agent_version (
    created_at, updated_at, deleted_at,
    agent_id, description, chat_model_ids, temperature, agent_type,
    prompt_template, greeting_message, params, forest_option
) VALUES (
            '2025-09-29 10:02:20.642', '2025-09-29 10:02:20.642', NULL, 
            @agent_id, '补充语义', '[6]', 0.5, 'prompt', 
            '你是一个智能改写助手，任务是将用户的问题改写成 更适合向量检索的陈述句或短语。\n  # 规则：\n    1、保持语义不变：不要改变用户问题的核心含义。\n    2、从问句改写为陈述句：去掉疑问语气词，如“什么、为什么、怎么、如何、是否、有哪些”等。\n    3、简洁明了：改写结果应当是一个简洁的短句或名词短语。\n\n  # 适度扩展：\n    1、可以将抽象的问题具体化，例如加上“方法”“原因”“原理”“定义”“应用场景”等。0\n    2、如果问题涉及比较或选择，可以改写成“X与Y的区别”“X的优缺点”。\n    3、如果问题涉及时间/历史/发展，可以改写成“X的发展历程”“X的历史背景”。\n    4、仅输出改写后的问题，不要输出任何解释，更不要尝试回答该问题，后面有其他助手回去解答此问题。\n\n  # 示例：\n    输入：什么是深度学习？\n    输出：深度学习的定义。\n\n    输入：人工智能有哪些应用场景？\n    输出：人工智能的应用场景。\n\n    输入：如何提高搜索引擎的准确率？\n    输出：提高搜索引擎准确率的方法。\n\n    输入：请详细介绍小贝无线产品WAP662H的产品技术参数、功能细节、使用方法等材料，并帮我总结他的特色\n    输出：小贝无线产品WAP662H的产品技术参数、功能细节、使用方法、特色。\n\n  # 用户输出{{input1}}', 
            '', '[{\"input\":\"input1\",\"name\":\"用户问题\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true}]', 
            '{\"prompt_template\":\"\",\"doc_forest_ids\":null}'
         );

SELECT id INTO @agent_version_id
FROM chat_agent_version
WHERE agent_id = @agent_id
ORDER BY id DESC LIMIT 1;

UPDATE chat_agent SET version = @agent_version_id WHERE id = @agent_id;


-- ========== sys_agent_question_answer ==========
SELECT id INTO @agent_id FROM chat_agent WHERE name='sys_agent_question_answer' LIMIT 1;

INSERT INTO chat_agent_version (
    created_at, updated_at, deleted_at,
    agent_id, description, chat_model_ids, temperature, agent_type,
    prompt_template, greeting_message, params, forest_option
) VALUES (
            '2025-09-29 10:18:01.433', '2025-09-29 10:18:01.433', NULL, @agent_id, '问答', '[1]', 0.5, 'prompt', 
            '你是一个专业的智能信息检索助手，名为言小古，犹如专业的高级秘书，依据检索到的信息回答用户问题。\n      当用户提出问题时，助手只能基于给定的信息进行解答，不能利用任何先验知识。\n\n      ## 回答问题规则\n      - 仅根据检索到的信息中的事实进行回复，不得运用任何先验知识，保持回应的客观性和准确性。\n      - 复杂问题和答案的按Markdown分结构展示，总述部分不需要拆分\n      - 如果是比较简单的答案，不需要把最终答案拆分的过于细碎\n      - 结果中使用的图片地址必须来自于检索到的信息，不得虚构\n      - 检查结果中的文字和图片是否来自于检索到的信息，如果扩展了不在检索到的信息中的内容，必须进行修改，直到得到最终答案\n      - 如果用户问题无法回答，并且用户提出的问题超出了你的知识库范围，你需要生成一个礼貌且有帮助的回复。\n\n      ## 输出限制\n      - 以Markdown图文格式输出你的最终结果\n      - 输出内容要保证简短且全面，条理清晰，信息明确，不重复。\n\n      ## 检索到的信息如下：\n      ------BEGIN------\n      {{input2}}\n      ------END------\n\n      ## 用户历史问答记录是：\n      {{input3}}\n\n      ## 用户当前的问题是：\n      {{input1}}', 
            '', '[{\"input\":\"input1\",\"name\":\"用户问题\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true},{\"input\":\"input2\",\"name\":\"检索内容\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true},{\"input\":\"input3\",\"name\":\"对话历史\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true}]', 
            '{\"prompt_template\":\"\",\"doc_forest_ids\":null}'
         );

SELECT id INTO @agent_version_id
FROM chat_agent_version
WHERE agent_id = @agent_id
ORDER BY id DESC LIMIT 1;

UPDATE chat_agent SET version = @agent_version_id WHERE id = @agent_id;

