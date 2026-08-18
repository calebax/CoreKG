SET NAMES utf8mb4;

START TRANSACTION;

INSERT INTO `chat_agent` (`created_at`, `updated_at`, `deleted_at`, `uin`, `company_id`, `avatar_url`, `name`, `show_name`, `public_scope`, `version`, `path`, `created_type`, `publish_status`, `manager_ids`, `agent_type`, `external_status`)
VALUES
	('2025-10-13 16:33:25.623', '2025-10-13 16:43:16.962', NULL, 0, 0, '/assets/prompt-CEUUcXkn.png', 'sys_agent_excel_header_row_number_list_analysis', '表格知识库问答@获取 excel 标题行号集合', 'company', '1841', '/lesson-plan', 'user', 'published', NULL, '', 'disabled');

-- 获取 chat_agent.id
SELECT id INTO @agent_id FROM chat_agent WHERE name = 'sys_agent_excel_header_row_number_list_analysis';

INSERT INTO `chat_agent_version` (`created_at`, `updated_at`, `deleted_at`, `agent_id`, `description`, `chat_model_ids`, `temperature`, `agent_type`, `prompt_template`, `greeting_message`, `params`, `forest_option`)
VALUES
	('2025-10-13 16:43:16.955', '2025-10-13 16:43:16.955', NULL, @agent_id, '', '[1]', '0.5', 'prompt', '你是一个 excel 专家，现在提供你一个表格解析后的数据，判断下标题包含了那几行数据，返回对应的行号列表，如[1,2]。\n表格数据：\n{{input1}}\n\n注：只返回对应的行号，不要说其他的！！！', '', '[{\"input\":\"input1\",\"name\":\"表格数据\",\"description\":\"\",\"is_title\":false,\"input_type\":\"text\",\"input_array\":[],\"is_required\":true}]', '{\"prompt_template\":\"\",\"doc_forest_ids\":null}');

SELECT id INTO @agent_version_id FROM chat_agent_version WHERE agent_id = @agent_id ORDER BY id DESC LIMIT 1;
UPDATE `chat_agent` SET `version` = @agent_version_id WHERE id = @agent_id;

COMMIT;