INSERT INTO `core_settings` (`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES
	(NULL, NULL, NULL, 'knowledge', 'agent_excel_question_to_sql', 'excel问答问题转 SQL', NULL, 'yaml', 'api_key: CHANGE_ME_API_KEY\nbase_url: https://api.example.com/v3/chat.Agent/chat/completions\nmodel_name: \"4RyUuX7\"', NULL);

INSERT INTO `core_settings` (`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES
	(NULL, NULL, NULL, 'knowledge', 'agent_excel_sql_result_analysis', 'excel问答sql结果分析', NULL, 'yaml', 'api_key: CHANGE_ME_API_KEY\nbase_url: https://api.example.com/v3/chat.Agent/chat/completions\nmodel_name: \"CJKGBXd\"', NULL);


INSERT INTO `core_settings` (`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES
	(now(), now(), NULL, 'knowledge', 'mysql_excel_instance', 'excel问答 mysql 实例信息', NULL, 'yaml', 'host: CHANGE_ME_HOST\nport: 63807\r\nusername: yygu_instance_test_rw\r\npassword: gTxG6Kgq4YmDCji2\r\charset: utf8mb4\r\n', NULL);