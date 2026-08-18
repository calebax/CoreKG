SET NAMES utf8mb4;

ALTER TABLE `chat_model`
ADD COLUMN `coze_model_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '对应coze model_instance id' AFTER `support_function_call`,
ADD INDEX `idx_chat_model_coze_model_id` (`coze_model_id`);
