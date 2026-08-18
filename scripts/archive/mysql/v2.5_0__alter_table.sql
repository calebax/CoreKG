SET NAMES utf8mb4;

ALTER TABLE `chat_model`
ADD COLUMN `model_group` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '模型分组，如 deepseek、openai' AFTER `model_provider`,
ADD COLUMN `support_function_call` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'unsupported' COMMENT '是否支持 function call：supported/unsupported' AFTER `model_group`,
ADD INDEX `idx_chat_model_group` (`model_group`);



CREATE TABLE `chat_recent_used_model` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `uin` bigint NOT NULL COMMENT '用户ID',
  `company_id` bigint NOT NULL DEFAULT 0 COMMENT '公司ID',
  `model_id` bigint NOT NULL COMMENT '模型ID，对应 chat_model.id',
  `last_used_at` datetime(3) NOT NULL COMMENT '最近使用时间',
  `usage_count` int NOT NULL DEFAULT 1 COMMENT '使用次数',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_company_model` (`uin`, `company_id`, `model_id`),
  KEY `idx_last_used` (`last_used_at`),
  KEY `idx_uin` (`uin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户最近使用的模型记录';


