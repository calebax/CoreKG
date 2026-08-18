-- ========== sys_agent_user_query_rewrite ==========
SELECT id INTO @agent_id FROM chat_agent WHERE name='sys_agent_user_query_rewrite' LIMIT 1;

INSERT INTO chat_agent_version (
    created_at, updated_at, deleted_at,
    agent_id, description, chat_model_ids, temperature, agent_type,
    prompt_template, greeting_message, params, forest_option
) VALUES (
            '2025-09-29 10:02:20.642', '2025-09-29 10:02:20.642', NULL, 
            @agent_id, '补充语义', '[1]', 0.5, 'prompt', 
            '你是一个智能改写助手，任务是将用户的问题改写成 更适合向量检索的陈述句或短语。\n  # 规则：\n    1、保持语义不变：不要改变用户问题的核心含义。\n    2、从问句改写为陈述句：去掉疑问语气词，如“什么、为什么、怎么、如何、是否、有哪些”等。\n    3、简洁明了：改写结果应当是一个简洁的短句或名词短语。\n\n  # 适度扩展：\n    1、可以将抽象的问题具体化，例如加上“方法”“原因”“原理”“定义”“应用场景”等。0\n    2、如果问题涉及比较或选择，可以改写成“X与Y的区别”“X的优缺点”。\n    3、如果问题涉及时间/历史/发展，可以改写成“X的发展历程”“X的历史背景”。\n    4、仅输出改写后的问题，不要输出任何解释，更不要尝试回答该问题，后面有其他助手回去解答此问题。\n\n  # 示例：\n    输入：什么是深度学习？\n    输出：深度学习的定义。\n\n    输入：人工智能有哪些应用场景？\n    输出：人工智能的应用场景。\n\n    输入：如何提高搜索引擎的准确率？\n    输出：提高搜索引擎准确率的方法。\n\n    输入：请详细介绍小贝无线产品WAP662H的产品技术参数、功能细节、使用方法等材料，并帮我总结他的特色\n    输出：小贝无线产品WAP662H的产品技术参数、功能细节、使用方法、特色。\n\n  # 用户输出{{input1}}', 
            '', '[{\"input\":\"input1\",\"name\":\"用户问题\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true}]', 
            '{\"prompt_template\":\"\",\"doc_forest_ids\":null}'
         );

SELECT id INTO @agent_version_id
FROM chat_agent_version
WHERE agent_id = @agent_id
ORDER BY id DESC LIMIT 1;

UPDATE chat_agent SET version = @agent_version_id WHERE id = @agent_id;
