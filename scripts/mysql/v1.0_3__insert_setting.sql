INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'core',
    'redis',
    'Redis',
    'Redis',
    'yaml',
    'addr: redis:6379\r\npassword:  \"\"\r\ndb: 1',
    '{}',
    NOW(),
    NOW()
  );


INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'core',
    'jwt-yygu',
    'jwt',
    'jwt',
    'yaml',
    'secret: 9RStaPSm0BtX9AiyLEYj4p2LyBVyG\r\nexpire: 480h\r\n',
    '{}',
    NOW(),
    NOW()
  );

-- INSERT INTO core_settings
--   (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
-- VALUES
--   (
--     'account',
--     'wechat_web_oauth',
--     '微信网页登录',
--     '微信网页登录',
--     'yaml',
--     'name: wechat_web_oauth\r\nappid: xxxxxxxxxxxxxxx\r\nappsecret: xxxxxxxxxxxxxxxxxxxxxxxxxxxx\n',
--     '{}',
--     NOW(),
--     NOW()
--   );

-- INSERT INTO core_settings
--   (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
-- VALUES
--   (
--     'account',
--     'sms_send_verify_code',
--     '发送短信验证码',
--     '发送短信验证码',
--     'yaml',
--     'aliyun:\r\n  access_key_id: \"LTAI5t6dRDZjh7EspsNE3qmC\"\r\n  access_key_secret: \"31IfGrNhAigTekRaBX2U6NbEOdyDAD\"\r\nsign_name: \"言古科技\"\r\ntemplate_code: \"SMS_462005456\"',
--     '{}',
--     NOW(),
--     NOW()
--   );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'forest_prompt',
    'forest_prompt',
    'forest_prompt',
    'yaml',
    'prompt_template: \"- 严格根据提供的知识库内容回答用户问题。\\n- 避免添加任何额外信息或解释。\\n- 只对用户提问进行思考分析，不要分析其他内容。\\n\\n## 输入\\n- 用户提问：{{input1}}\\n- 知识库详情：{{input2}}\\n- 对话记录：{{input3}}\"\r\n',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'chat',
    'llm',
    '模型地址',
    '模型地址',
    'yaml',
    'llmurl: https://yygu.cn/v3/llm.chat/chat/completions\r\napikey: CHANGE_ME_API_KEY',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knownow',
    'llm_image_parse',
    'LLM多模态解析',
    'LLM多模态解析',
    'yaml',
    'api_key: CHANGE_ME_API_KEY\r\nbase_url: https://api.example.com/v3/llm.chat/chat/completions\r\nmodel_name: "qwen2.5-vl-72b-instruct"',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'redis',
    'Redis',
    'Redis',
    'yaml',
    'addr: redis:6379\r\npassword: ""\r\ndb: 4\r\n',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'core',
    'cos-ke',
    'knowlegde存储',
    'knowlegde存储',
    'yaml',
    'purpose: forest-file\r\npresigned_timeout: \"24h\"\r\ns3:\r\n  end_point: http://CHANGE_ME_HOST:30000\r\n  access_key_id: admin\r\n  secret_access_key: CHANGE_ME_S3_SECRET_KEY\r\n  region: minio\r\n  bucket: corekg-bucket\r\n  use_path_style: true\r\n\r\n',
    '{}',
    NOW(),
    NOW()
  );


INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'agent_subquestion_chat',
    'agent_subquestion_chat',
    '',
    'yaml',
    'api_key: \"CHANGE_ME_API_KEY\"\r\nbase_url: \"http://CHANGE_ME_HOST:30000/v3/chat.Agent/chat/completions\"\r\nmodel_name: \"X9IIdTO\"',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'es',
    'es配置',
    'es配置',
    'yaml',
    'addresses:\r\n  - http://elasticsearch-0.elasticsearch:9200\r\nusername: elastic\r\npassword: CHANGE_ME_PASSWORD\r\nmax_retries: 3\r\nslow_threshold: 3s',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'agent_intention_recognition',
    '意图识别',
    '意图识别',
    'yaml',
    'api_key: CHANGE_ME_API_KEY\r\nbase_url: http://CHANGE_ME_HOST:30000/v3/chat.Agent/chat/completions\r\nmodel_name: \"oSMfiUt\"',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'agent_es_chat',
    'es问答提示词',
    'es问答提示词',
    'yaml',
    'api_key: \"CHANGE_ME_API_KEY\"\r\nbase_url: \"http://CHANGE_ME_HOST:30000/v3/chat.Agent/chat/completions\"\r\n#model_name: \"RO1rf6I\"\r\nmodel_name: \"WQLVajy\"',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'llm_image_parse',
    'llm_image_parse',
    'llm_image_parse',
    'yaml',
    'base_url: \"https://api.example.com/v3/llm.chat/chat/completions\"\r\napi_key: \"CHANGE_ME_API_KEY\"\r\nmodel_name: \"qwen2.5-vl-72b-instruct\"',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'nebula',
    '图数据库配置',
    '图数据库配置',
    'yaml',
    'address: \"nebula-graphd-0.nebula-graphd\"\r\nport: 9669\r\nusername: \"root\"\r\npassword: \"nebula\"\r\nprefix: \"know\"',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'nebulacount',
    '知识图谱数量配置',
    '知识图谱数量配置',
    'yaml',
    'graph_count: 3\r\nwordcloud_count: 50',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'preset_forest',
    '预置知识库',
    '预置知识库',
    'yaml',
    'forest_ids:\r\n  - 30',
    '{}',
    NOW(),
    NOW()
  );

INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'embedding',
    'embedding模型配置',
    'embedding模型配置',
    'yaml',
    'url: https://example.com:53081/v1/embeddings\r\nkey: \"123\"',
    '{}',
    NOW(),
    NOW()
  );


INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'convert_pdf',
    '文件转PDF',
    '文件转PDF',
    'yaml',
    'url: \"http://CHANGE_ME_HOST:30000/forms/libreoffice/convert\"',
    '{}',
    NOW(),
    NOW()
  );


INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'system_config',
    '全局配置',
    '全局配置',
    'yaml',
    'algo_model_id: 1',
    '{}',
    NOW(),
    NOW()
  );


INSERT INTO core_settings
  (`group`, `key`, `name`, `describe`, `value_type`, `value`, `default`, `created_at`, `updated_at`)
VALUES
  (
    'knowledge',
    'loc_redis',
    'Redis',
    'Redis',
    'yaml',
    'addr: redis:6379\r\npassword: ""\r\ndb: 4\r\n',
    '{}',
    NOW(),
    NOW()
  );