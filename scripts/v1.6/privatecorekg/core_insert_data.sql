INSERT INTO core_settings 
(`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES (NULL,
        NULL, 
        NULL, 
        'knowledge', 
        'agent_excel_question_to_sql', 
        'excel问答问题转 SQL', 
        NULL, 
        'yaml', 
        'api_key: @yg_AGENT_MODEL_APIKEY\nbase_url: http://${BASE_URL}/v3/chat.Agent/chat/completions\nmodel_name: \"4RyUuX7\"', 
        NULL);

INSERT INTO core_settings 
(`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES (NULL,
        NULL, 
        NULL, 
        'knowledge', 
        'agent_excel_sql_result_analysis', 
        'excel问答sql结果分析', 
        NULL, 
        'yaml', 
        'api_key: @yg_AGENT_MODEL_APIKEY\nbase_url: http://${BASE_URL}/v3/chat.Agent/chat/completions\nmodel_name: \"CJKGBXd\"', 
        NULL);


INSERT INTO core_settings
(`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES (now(),
        now(), 
        NULL, 
        'knowledge', 
        'mysql_excel_instance',
        'excel问答 mysql 实例信息', 
        NULL, 
        'yaml', 
        'host: ${MYSQL_HOST}\nport: ${MYSQL_PORT}\r\nusername: {MYSQL_USER} \r\npassword: ${MYSQL_PASSWORD} \r\charset: utf8mb4\r\n', 
        NULL);
        