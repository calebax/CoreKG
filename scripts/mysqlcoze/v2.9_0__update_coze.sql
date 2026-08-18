SET NAMES utf8mb4;
START TRANSACTION;

-- start 20251015103940_update.sql

-- Modify "agent_tool_draft" table
ALTER TABLE `agent_tool_draft` ADD COLUMN `source` tinyint NOT NULL DEFAULT 0 COMMENT "tool source 1 coze saas 0 default";
-- Modify "agent_tool_version" table
ALTER TABLE `agent_tool_version` ADD COLUMN `source` tinyint NOT NULL DEFAULT 0 COMMENT "tool source 1 coze saas 0 default";

-- Modify "shortcut_command" table
ALTER TABLE `shortcut_command` ADD COLUMN `source` tinyint NULL DEFAULT 0 COMMENT "plugin source 1 coze saas 0 default";


-- Modify "app_conversation_template_draft" table
ALTER TABLE `app_conversation_template_draft` COMMENT "app_conversation_template_draft";
-- Modify "app_conversation_template_online" table
ALTER TABLE `app_conversation_template_online` COMMENT "app_conversation_template_online";
-- Modify "app_dynamic_conversation_draft" table
ALTER TABLE `app_dynamic_conversation_draft` COMMENT "app_dynamic_conversation_draft";
-- Modify "app_dynamic_conversation_online" table
ALTER TABLE `app_dynamic_conversation_online` COMMENT "app_dynamic_conversation_online";
-- Modify "chat_flow_role_config" table
ALTER TABLE `chat_flow_role_config` COMMENT "chat_flow_role_config", MODIFY COLUMN `description` mediumtext NULL COMMENT "role description", MODIFY COLUMN `version` varchar(256) NULL COMMENT "version", MODIFY COLUMN `background_image_info` mediumtext NULL COMMENT "background image information, object structure", MODIFY COLUMN `onboarding_info` mediumtext NULL COMMENT "intro information, object structure", MODIFY COLUMN `suggest_reply_info` mediumtext NULL COMMENT "user suggestions, object structure", MODIFY COLUMN `audio_config` mediumtext NULL COMMENT "agent audio config, object structure";
-- Modify "conversation" table
ALTER TABLE `conversation` MODIFY COLUMN `name` varchar(255) NULL DEFAULT "" COMMENT "conversation name";
-- Modify "node_execution" table
ALTER TABLE `node_execution` COLLATE utf8mb4_unicode_ci;
-- Modify "prompt_resource" table
ALTER TABLE `prompt_resource` COLLATE utf8mb4_unicode_ci;
-- Modify "variable_instance" table
ALTER TABLE `variable_instance` COLLATE utf8mb4_unicode_ci;
-- Modify "variables_meta" table
ALTER TABLE `variables_meta` COLLATE utf8mb4_unicode_ci;
-- Create "kv_entries" table
CREATE TABLE `kv_entries` (`id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT "id", `namespace` varchar(255) NOT NULL COMMENT "namespace", `key_data` varchar(255) NOT NULL COMMENT "key_data", `value_data` longblob NULL COMMENT "value_data", PRIMARY KEY (`id`), UNIQUE INDEX `uniq_namespace_key` (`namespace`, `key_data`)) CHARSET utf8mb4 COLLATE utf8mb4_general_ci COMMENT "kv data";
-- Create "model_instance" table
CREATE TABLE `model_instance` (`id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT "id", `type` tinyint NOT NULL COMMENT "Model Type 0-LLM 1-TextEmbedding 2-Rerank ", `provider` json NOT NULL COMMENT "Provider Information", `display_info` json NOT NULL COMMENT "Display Information", `connection` json NOT NULL COMMENT "Connection Information", `capability` json NOT NULL COMMENT "Model Capability", `parameters` json NOT NULL COMMENT "Model Parameters", `extra` json NULL COMMENT "Extra Information", `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "Create Time in Milliseconds", `updated_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT "Update Time in Milliseconds", `deleted_at` datetime(3) NULL COMMENT "Delete Time", PRIMARY KEY (`id`)) CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT "Model Instance Management Table";


-- Modify "app_static_conversation_draft" table
ALTER TABLE `app_static_conversation_draft` COMMENT "app_static_conversation_draft";
-- Modify "app_static_conversation_online" table
ALTER TABLE `app_static_conversation_online` COMMENT "app_static_conversation_online";


-- Modify "conversation" table
ALTER TABLE `conversation` ADD COLUMN `user_id` varchar(255) NOT NULL DEFAULT "" COMMENT "user id with runtime";


-- end 20251028085526_update.sql

COMMIT;
