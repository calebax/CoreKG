SET NAMES utf8mb4;

--
-- Table structure for table `account_api_key`
--
CREATE TABLE IF NOT EXISTS `account_api_key` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL COMMENT 'uin',
  `company_id` bigint DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `api_key` varchar(255) NOT NULL,
  `purpose` varchar(255) DEFAULT NULL,
  `resource_type` varchar(255) DEFAULT NULL,
  `resource_id` varchar(255) DEFAULT NULL,
  `status` varchar(11) NOT NULL DEFAULT 'normal',
  `expired_at` datetime DEFAULT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_api_key` (`api_key`),
  KEY `idx_account_api_key_deleted_at` (`deleted_at`),
  KEY `company_id` (`company_id`),
  KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_api_key_privilege`
--
CREATE TABLE IF NOT EXISTS `account_api_key_privilege` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `api_key_id` bigint NOT NULL,
  `api_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_account_api_key_privilege_deleted_at` (`deleted_at`),
  KEY `api_key_id` (`api_key_id`),
  KEY `api_id` (`api_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_employee`
--
CREATE TABLE IF NOT EXISTS `account_employee` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint NOT NULL COMMENT '公司ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `uin` bigint NOT NULL COMMENT 'uin',
  `sys_role` varchar(16) DEFAULT '' COMMENT '系统角色',
  PRIMARY KEY (`id`),
  KEY `idx_account_employee_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_position`
--
CREATE TABLE IF NOT EXISTS `account_position` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint NOT NULL COMMENT '公司ID',
  `name` varchar(64) NOT NULL COMMENT '''名称''',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '''描述''',
  PRIMARY KEY (`id`),
  KEY `idx_account_position_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_privilege`
--
CREATE TABLE IF NOT EXISTS `account_privilege` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint NOT NULL COMMENT '公司ID',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '''描述''',
  `api` varchar(64) NOT NULL COMMENT '''API''',
  `action` varchar(64) NOT NULL COMMENT '''操作''',
  `action_path` varchar(64) NOT NULL COMMENT '''操作路径''',
  `parent_id` bigint NOT NULL DEFAULT '0',
  `type` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_account_privilege_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_rel_employee_position`
--
CREATE TABLE IF NOT EXISTS `account_rel_employee_position` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `employee_id` bigint NOT NULL,
  `position_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniidx_employee_position` (`employee_id`,`position_id`),
  KEY `idx_account_rel_employee_position_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_rel_position_privilege`
--
CREATE TABLE IF NOT EXISTS `account_rel_position_privilege` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `position_id` bigint NOT NULL,
  `privilege_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniidx_position_privilege` (`position_id`,`privilege_id`),
  KEY `idx_account_rel_position_privilege_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `account_wechat_binding`
--
CREATE TABLE IF NOT EXISTS `account_wechat_binding` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `app_id` varchar(100) NOT NULL,
  `open_id` varchar(100) NOT NULL,
  `union_id` varchar(100) NOT NULL,
  `subscribe` tinyint(1) DEFAULT '0',
  `nickname` varchar(255) DEFAULT '',
  `sex` tinyint(1) DEFAULT '0',
  `city` varchar(255) DEFAULT '',
  `country` varchar(255) DEFAULT '',
  `province` varchar(255) DEFAULT '',
  `headimgurl` varchar(255) DEFAULT '',
  `subscribe_time` int DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_account_wechat_binding_open_id` (`open_id`),
  KEY `idx_account_wechat_binding_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `admin_login_setting`
--
CREATE TABLE IF NOT EXISTS `admin_login_setting` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `domain_name` varchar(255) NOT NULL,
  `path` varchar(128) DEFAULT NULL,
  `env` varchar(128) DEFAULT NULL,
  `is_enable_we_chat` tinyint NOT NULL,
  `is_enable_we_chat_com` tinyint NOT NULL,
  `is_enable_phone` tinyint NOT NULL,
  `is_enable_email` tinyint NOT NULL,
  `is_enable_password` tinyint NOT NULL DEFAULT '0',
  `title` varchar(255) NOT NULL,
  `image_url` varchar(255) NOT NULL,
  `app_id` varchar(255) NOT NULL,
  `app_id_com` varchar(255) NOT NULL,
  `agent_id` varchar(255) NOT NULL,
  `allow_register` tinyint NOT NULL,
  `issuer` varchar(128) NOT NULL,
  `auth_key` varchar(128) NOT NULL,
  `login_url` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_login_setting_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_agent`
--
CREATE TABLE IF NOT EXISTS `chat_agent` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '公司ID',
  `avatar_url` varchar(256) DEFAULT NULL COMMENT '机器人头像',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '机器人名称',
  `show_name` varchar(50) NOT NULL DEFAULT '' COMMENT '机器人显示名称',
  `agent_type` varchar(32) NOT NULL DEFAULT '' COMMENT '机器人类型',
  `public_scope` varchar(32) NOT NULL DEFAULT 'private' COMMENT '公开范围',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '版本',
  `path` varchar(256) NOT NULL DEFAULT '' COMMENT '应用路径',
  `created_type` varchar(32) NOT NULL DEFAULT 'user' COMMENT '创建类型',
  `publish_status` varchar(32) DEFAULT 'draft' COMMENT '发布状态',
  `manager_ids` varchar(256) DEFAULT NULL COMMENT '管理员ID',
  `external_status` varchar(63) DEFAULT 'disabled' COMMENT '''外部调用状态''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `idx_chat_agent_deleted_at` (`deleted_at`),
  KEY `uin` (`uin`),
  KEY `company_id` (`company_id`),
  KEY `show_name` (`show_name`),
  KEY `agent_type` (`agent_type`),
  KEY `version` (`version`),
  KEY `path` (`path`),
  KEY `publish_status` (`publish_status`)
) ENGINE=InnoDB AUTO_INCREMENT=100040 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_agent_collect`
--
CREATE TABLE IF NOT EXISTS `chat_agent_collect` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL COMMENT '用户唯一标识',
  `agent_app_id` bigint NOT NULL COMMENT '应用ID',
  PRIMARY KEY (`id`),
  KEY `idx_chat_agent_collect_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_agent_history`
--
CREATE TABLE IF NOT EXISTS `chat_agent_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `api_key` varchar(255) NOT NULL,
  `request_id` varchar(64) NOT NULL,
  `retry_times` bigint NOT NULL DEFAULT '1' COMMENT '''重试次数''',
  `input` mediumtext,
  `output` text,
  `reasoning_output` text,
  `agent_name` varchar(255) NOT NULL DEFAULT '' COMMENT '''机器人名称''',
  `model_name` varchar(255) NOT NULL,
  `model_provider` varchar(50) NOT NULL DEFAULT '' COMMENT '''模型供应商''',
  `code` bigint NOT NULL DEFAULT '0' COMMENT '''返回码''',
  `message` text COMMENT '''返回信息''',
  `status` varchar(50) NOT NULL DEFAULT '' COMMENT '''返回状态''',
  `start_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '''开始时间''',
  `end_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '''结束时间''',
  `cost` double NOT NULL DEFAULT '0' COMMENT '''花费时间''',
  `prompt_tokens` bigint NOT NULL DEFAULT '0' COMMENT '''提示token''',
  `out_token` bigint NOT NULL DEFAULT '0' COMMENT '''输出token''',
  `cache_hit_token` bigint NOT NULL DEFAULT '0' COMMENT '''缓存命中token''',
  `cache_miss_token` bigint NOT NULL DEFAULT '0' COMMENT '''缓存未命中token''',
  `query_reference_list` mediumtext COMMENT '''引用内容列表''',
  `is_charged` tinyint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_chat_agent_history_deleted_at` (`deleted_at`),
  KEY `idx_chat_agent_history_uin` (`uin`),
  KEY `idx_chat_agent_history_company_id` (`company_id`),
  KEY `api_key` (`api_key`),
  KEY `request_id` (`request_id`),
  KEY `model_name` (`model_name`),
  KEY `prompt_tokens` (`prompt_tokens`),
  KEY `out_token` (`out_token`),
  KEY `cache_hit_token` (`cache_hit_token`),
  KEY `cache_miss_token` (`cache_miss_token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_agent_public_scope`
--
CREATE TABLE IF NOT EXISTS `chat_agent_public_scope` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `agent_id` bigint NOT NULL DEFAULT '0' COMMENT '机器人ID',
  `scope_type` varchar(32) NOT NULL DEFAULT '' COMMENT '公开范围类型',
  `scope_id` bigint NOT NULL DEFAULT '0' COMMENT '公开范围ID',
  PRIMARY KEY (`id`),
  KEY `idx_chat_agent_public_scope_deleted_at` (`deleted_at`),
  KEY `agent_id` (`agent_id`),
  KEY `scope_type` (`scope_type`),
  KEY `scope_id` (`scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_agent_version`
--
CREATE TABLE IF NOT EXISTS `chat_agent_version` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `agent_id` bigint NOT NULL DEFAULT '0' COMMENT '机器人ID',
  `description` varchar(256) DEFAULT NULL COMMENT '描述',
  `chat_model_ids` varchar(256) DEFAULT NULL COMMENT '模型ID',
  `temperature` double DEFAULT '0.5' COMMENT '自定义温度',
  `agent_type` varchar(32) NOT NULL DEFAULT '' COMMENT '机器人类型',
  `prompt_template` text NOT NULL COMMENT '提示词模板',
  `greeting_message` varchar(256) DEFAULT NULL COMMENT '问候信息',
  `tag` varchar(256) DEFAULT NULL COMMENT '标签',
  `params` text COMMENT '参数',
  `forest_option` varchar(256) DEFAULT NULL COMMENT '知识库ID',
  PRIMARY KEY (`id`),
  KEY `idx_chat_agent_version_deleted_at` (`deleted_at`),
  KEY `agent_id` (`agent_id`),
  KEY `agent_type` (`agent_type`)
) ENGINE=InnoDB AUTO_INCREMENT=1000444 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_model`
--
CREATE TABLE IF NOT EXISTS `chat_model` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `show_name` varchar(64) NOT NULL COMMENT '''模型的显示名称''',
  `api_key` varchar(255) DEFAULT NULL COMMENT '''API密钥''',
  `model_name` varchar(64) NOT NULL COMMENT '''模型名称''',
  `model_url` varchar(255) DEFAULT NULL COMMENT '''模型URL''',
  `public_type` varchar(32) DEFAULT NULL COMMENT '''公开类型''',
  `model_provider` varchar(50) DEFAULT NULL COMMENT '''模型供应商''',
  `head_url` varchar(255) DEFAULT NULL COMMENT '''模型图片地址''',
  PRIMARY KEY (`id`),
  KEY `idx_chat_model_deleted_at` (`deleted_at`),
  KEY `idx_chat_model_uin` (`uin`),
  KEY `idx_chat_model_company_id` (`company_id`),
  KEY `idx_chat_model_show_name` (`show_name`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_questions`
--
CREATE TABLE IF NOT EXISTS `chat_questions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `chat_session_id` bigint NOT NULL DEFAULT '0' COMMENT '''群ID''',
  `api_key_id` bigint NOT NULL DEFAULT '0' COMMENT '''APIKeyID''',
  `stream_key` varchar(32) DEFAULT NULL COMMENT '''流密钥''',
  `real_ip` varchar(32) NOT NULL DEFAULT '' COMMENT '''用户IP''',
  `from` varchar(32) NOT NULL DEFAULT '' COMMENT '''来源''',
  `question` mediumtext COMMENT '''问题''',
  `reasoning` mediumtext COMMENT '''推理''',
  `reasoning_seconds` int NOT NULL DEFAULT '0' COMMENT '''推理耗时''',
  `answer` text COMMENT '''回答''',
  `cost_seconds` int NOT NULL DEFAULT '0' COMMENT '''耗时''',
  `status` varchar(8) NOT NULL DEFAULT 'pending' COMMENT '''状态''',
  `model_name` varchar(255) NOT NULL,
  `model_id` bigint NOT NULL DEFAULT '0' COMMENT '模型ID',
  `base_agent_id` bigint NOT NULL DEFAULT '0' COMMENT '机器人ID',
  `out_token` bigint NOT NULL DEFAULT '0' COMMENT '''输出token''',
  `cache_hit_token` bigint NOT NULL DEFAULT '0' COMMENT '''缓存命中token''',
  `cache_miss_token` bigint NOT NULL DEFAULT '0' COMMENT '''缓存未命中token''',
  `is_charged` tinyint NOT NULL,
  `answer_stars` bigint NOT NULL DEFAULT '0' COMMENT '''回答星星''',
  `image_url_list` text COMMENT '''图片列表''',
  `query_reference_list` mediumtext COMMENT '''引用内容列表''',
  `chat_reference_list` mediumtext COMMENT '''引用内容列表''',
  `external_id` varchar(127) DEFAULT NULL COMMENT '''外部调用标识''',
  PRIMARY KEY (`id`),
  KEY `idx_chat_questions_deleted_at` (`deleted_at`),
  KEY `idx_chat_questions_uin` (`uin`),
  KEY `idx_chat_questions_company_id` (`company_id`),
  KEY `idx_chat_questions_chat_session_id` (`chat_session_id`),
  KEY `idx_chat_questions_api_key_id` (`api_key_id`),
  KEY `idx_chat_questions_stream_key` (`stream_key`),
  KEY `model_name` (`model_name`),
  KEY `model_id` (`model_id`),
  KEY `base_agent_id` (`base_agent_id`),
  KEY `out_token` (`out_token`),
  KEY `cache_hit_token` (`cache_hit_token`),
  KEY `cache_miss_token` (`cache_miss_token`),
  KEY `answer_stars` (`answer_stars`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `chat_sessions`
--
CREATE TABLE IF NOT EXISTS `chat_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `name` varchar(255) NOT NULL,
  `model_name` varchar(255) NOT NULL,
  `model_id` bigint NOT NULL DEFAULT '0' COMMENT '模型ID',
  `base_agent_id` bigint NOT NULL DEFAULT '0' COMMENT '机器人ID',
  `resource_type` varchar(255) NOT NULL COMMENT '''资源类型，用于区分是 Model 还是 Agent''',
  `agent_version` bigint DEFAULT '0' COMMENT '机器人版本',
  `input` json DEFAULT NULL COMMENT '''输入参数''',
  `is_top` tinyint NOT NULL DEFAULT '-1' COMMENT '''是否置顶''',
  `file_id` bigint NOT NULL DEFAULT '0' COMMENT 'file_id',
  `file_id_list` text COMMENT '''文件ID列表''',
  `forest_id_list` text COMMENT '''森林ID列表''',
  `es_index` varchar(255) DEFAULT 'ke_0' COMMENT '''es索引''',
  `external_id` varchar(127) DEFAULT NULL COMMENT '''外部调用标识''',
  PRIMARY KEY (`id`),
  KEY `idx_chat_sessions_deleted_at` (`deleted_at`),
  KEY `idx_chat_sessions_uin` (`uin`),
  KEY `idx_chat_sessions_company_id` (`company_id`),
  KEY `name` (`name`),
  KEY `model_name` (`model_name`),
  KEY `model_id` (`model_id`),
  KEY `base_agent_id` (`base_agent_id`),
  KEY `resource_type` (`resource_type`),
  KEY `agent_version` (`agent_version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `company`
--
CREATE TABLE IF NOT EXISTS `company` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `alias` varchar(64) DEFAULT NULL,
  `description` varchar(256) DEFAULT NULL,
  `logo` varchar(256) DEFAULT NULL,
  `address` varchar(256) DEFAULT NULL,
  `tel` varchar(256) DEFAULT NULL,
  `email` varchar(256) DEFAULT NULL,
  `website` varchar(256) DEFAULT NULL,
  `company_status` varchar(32) DEFAULT NULL,
  `version` varchar(63) DEFAULT 'free_trail',
  `quota` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `udx_name` (`name`),
  KEY `idx_company_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `company_invitation`
--
CREATE TABLE IF NOT EXISTS `company_invitation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `issuer` varchar(30) NOT NULL,
  `count` bigint unsigned DEFAULT '1',
  `expired` datetime NOT NULL,
  `company_id` bigint unsigned NOT NULL,
  `key` varchar(255) NOT NULL,
  `already_bind` tinyint(1) DEFAULT '0',
  `role` varchar(50) NOT NULL,
  `perm_set` text,
  `uin` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_company_invitation_key` (`key`),
  KEY `idx_company_invitation_deleted_at` (`deleted_at`),
  KEY `idx_company_invitation_issuer` (`issuer`),
  KEY `idx_company_invitation_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `company_upgrade_apply`
--
CREATE TABLE IF NOT EXISTS `company_upgrade_apply` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(127) DEFAULT NULL,
  `phone` varchar(127) DEFAULT NULL,
  `company_name` varchar(255) DEFAULT NULL,
  `scale` varchar(511) DEFAULT NULL,
  `industry` varchar(1023) DEFAULT NULL,
  `note` varchar(1023) DEFAULT NULL,
  `claim` varchar(1023) DEFAULT NULL,
  `type` varchar(63) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_company_upgrade_apply_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `core_settings`
--
CREATE TABLE IF NOT EXISTS `core_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `group` varchar(16) DEFAULT NULL,
  `key` varchar(64) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `describe` varchar(128) DEFAULT NULL,
  `value_type` varchar(16) DEFAULT 'text',
  `value` text,
  `default` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_setting` (`group`,`key`),
  KEY `idx_core_settings_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=10004 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `core_task`
--
CREATE TABLE IF NOT EXISTS `core_task` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `subject_id` bigint NOT NULL COMMENT 'subject id',
  `step` bigint NOT NULL COMMENT 'task step',
  `task_type` varchar(32) NOT NULL COMMENT 'task type',
  `priority` tinyint NOT NULL COMMENT 'task priority',
  `task_status` varchar(12) NOT NULL COMMENT 'task status',
  `redo` bigint NOT NULL COMMENT 'redo times for parsing',
  `err_msg` text COMMENT 'error message',
  `comment` varchar(255) DEFAULT NULL COMMENT 'comment for parse task',
  `worker_id` varchar(255) DEFAULT NULL COMMENT 'worker id',
  `start_at` datetime DEFAULT NULL COMMENT 'task start time',
  `end_at` datetime DEFAULT NULL COMMENT 'task end time',
  `cost` bigint DEFAULT NULL COMMENT 'task cost',
  `payload` text COMMENT 'task payload',
  `result` text COMMENT 'task result',
  `app_group` varchar(32) DEFAULT NULL COMMENT 'task app group',
  `task_config_redo` bigint NOT NULL DEFAULT '0' COMMENT '''任务配置重试次数''',
  `task_config_timeout` bigint NOT NULL DEFAULT '0' COMMENT '''任务配置超时时间''',
  PRIMARY KEY (`id`),
  KEY `idx_core_task_deleted_at` (`deleted_at`),
  KEY `idx_core_task_uin` (`uin`),
  KEY `idx_core_task_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `core_upload_files`
--
CREATE TABLE IF NOT EXISTS `core_upload_files` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint unsigned DEFAULT NULL,
  `uin` bigint unsigned DEFAULT NULL,
  `purpose` varchar(32) DEFAULT NULL,
  `filename` varchar(128) DEFAULT NULL,
  `file_ext` varchar(8) DEFAULT NULL,
  `mime_type` longtext,
  `size` bigint DEFAULT NULL,
  `hash` varchar(256) DEFAULT NULL,
  `chunk_hash` varchar(64) DEFAULT NULL,
  `path` varchar(128) DEFAULT NULL,
  `public_url` varchar(256) DEFAULT NULL,
  `copy_number` bigint DEFAULT '1',
  `status` varchar(15) DEFAULT 'normal',
  `err_msg` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_core_upload_files_deleted_at` (`deleted_at`),
  KEY `idx_core_upload_files_hash` (`hash`),
  KEY `idx_core_upload_files_public_url` (`public_url`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `core_upload_files_tmp`
--
CREATE TABLE IF NOT EXISTS `core_upload_files_tmp` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint unsigned DEFAULT NULL,
  `uin` bigint unsigned DEFAULT NULL,
  `employee_id` bigint unsigned DEFAULT NULL,
  `purpose` varchar(32) DEFAULT NULL,
  `filename` varchar(128) DEFAULT NULL,
  `file_ext` varchar(8) DEFAULT NULL,
  `mime_type` longtext,
  `size` bigint DEFAULT NULL,
  `chunk_hash` varchar(64) DEFAULT NULL,
  `path` varchar(128) DEFAULT NULL,
  `public_url` varchar(256) DEFAULT NULL,
  `err_msg` varchar(256) DEFAULT NULL,
  `annotations` varchar(255) DEFAULT NULL,
  `third_upload_id` varchar(128) DEFAULT NULL,
  `expired_at` datetime(3) DEFAULT NULL,
  `part_size` bigint DEFAULT NULL,
  `part_count` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_core_upload_files_tmp_deleted_at` (`deleted_at`),
  KEY `idx_core_upload_files_tmp_company_id` (`company_id`),
  KEY `idx_core_upload_files_tmp_uin` (`uin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `individual`
--
CREATE TABLE IF NOT EXISTS `individual` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `real_name` varchar(32) DEFAULT NULL COMMENT '真实姓名',
  `id_card` varchar(18) DEFAULT NULL COMMENT '身份证号',
  `real_name_status` varchar(32) DEFAULT 'pending' COMMENT '实名状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_individual_user_id` (`user_id`),
  KEY `idx_individual_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_company_db`
--
CREATE TABLE IF NOT EXISTS `ke_company_db` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint unsigned NOT NULL COMMENT '公司ID',
  `db_instance_id` bigint unsigned NOT NULL COMMENT '数据库实例ID',
  `db_name` varchar(255) NOT NULL COMMENT '数据库名',
  PRIMARY KEY (`id`),
  KEY `idx_ke_company_db_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_file_parse`
--
CREATE TABLE IF NOT EXISTS `ke_file_parse` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `forest_name` varchar(64) NOT NULL COMMENT 'forest name',
  `path` varchar(255) NOT NULL COMMENT 'source file relative path',
  `md5` varchar(255) NOT NULL COMMENT 'check whether source file was touched',
  `content` longtext COMMENT 'parsed content',
  `mind_map` longtext COMMENT 'mind_map content',
  `analysis` longtext COMMENT 'analysis content',
  `status` varchar(30) NOT NULL COMMENT 'parse status of forest',
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_forest_path_md5_index` (`uin`,`forest_name`,`path`,`md5`),
  KEY `idx_ke_file_parse_deleted_at` (`deleted_at`),
  KEY `idx_ke_file_parse_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_file_qa`
--
CREATE TABLE IF NOT EXISTS `ke_file_qa` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `forest_id` bigint NOT NULL COMMENT 'forest_id',
  `file_id` bigint NOT NULL COMMENT 'forest_id',
  `question` text COMMENT '''问题''',
  `answer` text COMMENT '''回答''',
  `status` varchar(8) NOT NULL DEFAULT 'pending' COMMENT '''状态''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_file_qa_deleted_at` (`deleted_at`),
  KEY `idx_ke_file_qa_uin` (`uin`),
  KEY `idx_ke_file_qa_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest`
--
CREATE TABLE IF NOT EXISTS `ke_forest` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `name` varchar(255) NOT NULL COMMENT '知识森林名称',
  `knowledge_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT '知识库状态',
  `avatar_url` varchar(255) NOT NULL DEFAULT '' COMMENT '知识森林头像',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '知识森林描述',
  `forest_type` varchar(64) NOT NULL DEFAULT 'file' COMMENT '知识森林类型',
  `data_source_type` varchar(64) NOT NULL DEFAULT '' COMMENT '数据源类型',
  `data_source_subtype` varchar(64) NOT NULL DEFAULT '' COMMENT '数据源子类型',
  `config_id` bigint NOT NULL DEFAULT '0' COMMENT '配置ID',
  `public_scope` varchar(32) NOT NULL DEFAULT 'private' COMMENT '公开范围',
  `manager_ids` varchar(256) DEFAULT NULL COMMENT '管理员ID',
  `count` bigint DEFAULT '0' COMMENT '知识库资源计数',
  `disk_storage` varchar(255) DEFAULT NULL COMMENT '知识库磁盘占用',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_uin` (`uin`),
  KEY `idx_ke_forest_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_db`
--
CREATE TABLE IF NOT EXISTS `ke_forest_db` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `db_instance_id` bigint unsigned NOT NULL COMMENT '数据库实例ID',
  `db_meta` text NOT NULL COMMENT '数据库元数据，字符集等',
  `db_name` varchar(255) NOT NULL COMMENT '数据库名',
  `forest_id` bigint unsigned NOT NULL COMMENT '知识库ID',
  `size` bigint unsigned NOT NULL COMMENT '数据库大小（Bytes）',
  `synced_at` datetime NOT NULL COMMENT '同步时间',
  `uin` bigint unsigned NOT NULL COMMENT '用户uin',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_db_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_db_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_db_instance`
--
CREATE TABLE IF NOT EXISTS `ke_forest_db_instance` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `connect_mode` varchar(32) NOT NULL COMMENT '连接模式，standard：标准连接，ssh：ssh隧道模式',
  `connect_name` varchar(128) NOT NULL COMMENT '数据库连接名称',
  `forest_id` bigint unsigned NOT NULL COMMENT '知识库ID',
  `host` varchar(255) NOT NULL COMMENT '数据库地址',
  `instance_type` varchar(32) NOT NULL COMMENT '数据库类型，oracle：oracle，mysql：mysql',
  `ownership_type` varchar(32) NOT NULL COMMENT '实例归属类型：system-系统自有，external-外部实例',
  `username` varchar(32) NOT NULL COMMENT '连接用户名',
  `password` varchar(128) NOT NULL COMMENT '连接密码（加密存储）',
  `database` varchar(128) NOT NULL COMMENT '数据库名称',
  `port` bigint NOT NULL COMMENT '端口号',
  `uin` bigint unsigned NOT NULL COMMENT '用户uin',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_db_instance_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_db_instance_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_excel_sheet`
--
CREATE TABLE IF NOT EXISTS `ke_forest_excel_sheet` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `data_end_row_num` tinyint unsigned NOT NULL COMMENT '数据结束行号',
  `data_start_row_num` tinyint unsigned NOT NULL COMMENT '数据开始行号',
  `forest_file_id` bigint unsigned NOT NULL COMMENT '文件ID',
  `forest_id` bigint unsigned NOT NULL COMMENT '知识库ID',
  `forest_table_id` bigint unsigned NOT NULL COMMENT '关联数据表ID',
  `header_mode` varchar(16) NOT NULL COMMENT '表头模式，row_title：行表头，column_title：列表头',
  `header_row_num` tinyint unsigned NOT NULL COMMENT '表头行号（从1开始）',
  `sheet_meta` text COMMENT 'Sheet元数据（如字段结构的JSON描述）',
  `sheet_name` varchar(255) NOT NULL COMMENT 'Sheet名称',
  `total_column` bigint unsigned NOT NULL COMMENT '总列数（Excel数据列数）',
  `total_row` bigint unsigned NOT NULL COMMENT '总行数（Excel数据行数）',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_excel_sheet_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_file`
--
CREATE TABLE IF NOT EXISTS `ke_forest_file` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `forest_id` bigint NOT NULL COMMENT '知识森林id',
  `file_id` bigint NOT NULL COMMENT '文件id',
  `priview_file_id` bigint NOT NULL COMMENT '预览文件id',
  `priview_ext` varchar(64) NOT NULL DEFAULT '' COMMENT '预览文件后缀',
  `is_dir` tinyint(1) NOT NULL DEFAULT '-1',
  `parent_id` bigint NOT NULL COMMENT '知识森林id',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `size` bigint NOT NULL COMMENT '文件大小',
  `ext` varchar(64) NOT NULL COMMENT '文件后缀',
  `parent_ids` varchar(255) NOT NULL COMMENT 'parent_ids',
  `depth` tinyint NOT NULL COMMENT '深度',
  `parse_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT 'md解析状态',
  `mindmap_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT '思维导图状态',
  `analysis_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT '智能分析状态',
  `knowledge_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT '知识库状态',
  `graph_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT '知识图谱状态',
  `desc_status` varchar(64) NOT NULL DEFAULT 'pending' COMMENT '文件描述状态',
  `preview_able` varchar(64) DEFAULT 'accept' COMMENT '文件是否可预览',
  `status` varchar(15) DEFAULT 'normal',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_file_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_file_uin` (`uin`),
  KEY `idx_ke_forest_file_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_public_scope`
--
CREATE TABLE IF NOT EXISTS `ke_forest_public_scope` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `forest_id` bigint NOT NULL DEFAULT '0' COMMENT '知识森林ID',
  `scope_type` varchar(32) NOT NULL DEFAULT '' COMMENT '公开范围类型',
  `scope_id` bigint NOT NULL DEFAULT '0' COMMENT '公开范围ID',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_public_scope_deleted_at` (`deleted_at`),
  KEY `forest_id` (`forest_id`),
  KEY `scope_type` (`scope_type`),
  KEY `scope_id` (`scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_qa`
--
CREATE TABLE IF NOT EXISTS `ke_forest_qa` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `session_id` bigint NOT NULL COMMENT 'session_id',
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `question` text COMMENT '''问题''',
  `answer` text COMMENT '''回答''',
  `reasoning` text COMMENT '''推理''',
  `mind_graph` text COMMENT '''思维导图''',
  `content` text COMMENT '''image用到的content''',
  `image_url_list` text COMMENT '''图片列表''',
  `status` varchar(8) NOT NULL DEFAULT 'pending' COMMENT '''状态''',
  `query_reference_list` mediumtext COMMENT '''引用内容列表''',
  `chat_reference_list` mediumtext COMMENT '''引用内容列表''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_qa_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_qa_uin` (`uin`),
  KEY `idx_ke_forest_qa_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_forest_table`
--
CREATE TABLE IF NOT EXISTS `ke_forest_table` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `column_count` tinyint unsigned NOT NULL COMMENT '列数',
  `db_id` bigint unsigned NOT NULL COMMENT '数据库ID',
  `db_instance_id` bigint unsigned NOT NULL COMMENT '数据库实例ID',
  `forest_id` bigint unsigned NOT NULL COMMENT '知识库ID',
  `row_count` bigint unsigned NOT NULL COMMENT '行数',
  `size` bigint unsigned NOT NULL COMMENT '数据表大小（Bytes）',
  `synced_at` datetime NOT NULL COMMENT '同步时间',
  `table_meta` text COMMENT '表元数据（如字段结构的JSON描述）',
  `table_name` varchar(255) NOT NULL COMMENT '表名',
  `uin` bigint unsigned NOT NULL COMMENT '用户uin',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_table_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_parse_history`
--
CREATE TABLE IF NOT EXISTS `ke_parse_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `forest_name` varchar(64) NOT NULL COMMENT 'forest name',
  `path` varchar(255) NOT NULL COMMENT 'source file relative path',
  `md5` varchar(255) NOT NULL COMMENT 'check whether source file was touched',
  `content` longtext COMMENT 'parsed content',
  `mind_map` longtext COMMENT 'mind_map content',
  `analysis` longtext COMMENT 'analysis content',
  PRIMARY KEY (`id`),
  KEY `idx_ke_parse_history_deleted_at` (`deleted_at`),
  KEY `idx_ke_parse_history_uin` (`uin`),
  KEY `idx_ke_parse_history_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_qa_session`
--
CREATE TABLE IF NOT EXISTS `ke_qa_session` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `type` varchar(255) NOT NULL COMMENT '''类型''',
  `base_type` varchar(32) DEFAULT 'standard' COMMENT '''基础类型，standard：标准, data_excel：Excel, data_mysql：MySQL''',
  `file_id` bigint NOT NULL DEFAULT '0' COMMENT 'file_id',
  `file_id_list` text COMMENT '''文件ID列表''',
  `forest_id_list` text COMMENT '''森林ID列表''',
  `excel_id_list` text COMMENT '''excelIDList''',
  `excel_sheet_id_list` text COMMENT '''excelSheetIDList''',
  `name` varchar(255) DEFAULT 'New Chat' COMMENT '''名称''',
  `es_index` varchar(255) DEFAULT 'ke_0' COMMENT '''es索引''',
  `llm_model_id` bigint DEFAULT '0' COMMENT '''llm模型id''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_qa_session_deleted_at` (`deleted_at`),
  KEY `idx_ke_qa_session_uin` (`uin`),
  KEY `idx_ke_qa_session_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_resource_scope`
--
CREATE TABLE IF NOT EXISTS `ke_resource_scope` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `resource_type` varchar(24) NOT NULL DEFAULT '' COMMENT '资源类型',
  `resource_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '资源ID',
  `scope_type` varchar(24) NOT NULL DEFAULT '' COMMENT '作用域类型',
  `scope_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '作用域对象ID',
  `action` varchar(24) NOT NULL DEFAULT '' COMMENT '操作类型',
  PRIMARY KEY (`id`),
  KEY `idx_ke_resource_scope_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `ke_task`
--
CREATE TABLE IF NOT EXISTS `ke_task` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `forest_id` bigint NOT NULL COMMENT 'forest id',
  `file_id` bigint NOT NULL COMMENT 'file id',
  `task_type` varchar(12) NOT NULL COMMENT 'task type',
  `priority` tinyint NOT NULL COMMENT 'task priority',
  `path` varchar(255) NOT NULL COMMENT 'source file relative path',
  `task_status` varchar(12) NOT NULL COMMENT 'task status',
  `machine_id` varchar(255) DEFAULT NULL COMMENT 'machine id',
  `redo` bigint NOT NULL COMMENT 'redo times for parsing',
  `err_msg` varchar(255) DEFAULT NULL COMMENT 'error message',
  `comment` varchar(255) DEFAULT NULL COMMENT 'comment for parse task',
  `start_at` datetime DEFAULT NULL COMMENT 'task start time',
  `end_at` datetime DEFAULT NULL COMMENT 'task end time',
  `cost` bigint DEFAULT NULL COMMENT 'task cost',
  PRIMARY KEY (`id`),
  KEY `idx_ke_task_deleted_at` (`deleted_at`),
  KEY `idx_ke_task_uin` (`uin`),
  KEY `idx_ke_task_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `user`
--
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `identify` varchar(256) NOT NULL COMMENT '客户标识',
  `name` varchar(64) NOT NULL COMMENT '客户名称',
  `bio` varchar(256) DEFAULT NULL COMMENT '客户简介',
  `avatar_url` varchar(256) DEFAULT NULL COMMENT '客户头像',
  `email` varchar(64) DEFAULT NULL COMMENT '客户邮箱',
  `phone` varchar(16) DEFAULT NULL COMMENT '客户手机',
  `password` varchar(128) DEFAULT NULL,
  `github_id` bigint DEFAULT NULL COMMENT 'Github ID',
  `work_wechat_user_id` varchar(64) DEFAULT NULL COMMENT '企业微信用户ID',
  `wechat_union_id` varchar(64) DEFAULT NULL COMMENT '微信用户UnionID',
  `wechat_web_open_id` varchar(64) DEFAULT NULL COMMENT '微信用户WebOpenID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_identify` (`identify`),
  UNIQUE KEY `idx_user_github_id` (`github_id`),
  UNIQUE KEY `idx_user_work_wechat_user_id` (`work_wechat_user_id`),
  UNIQUE KEY `idx_user_wechat_union_id` (`wechat_union_id`),
  UNIQUE KEY `idx_user_wechat_web_open_id` (`wechat_web_open_id`),
  KEY `idx_user_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `user_identification`
--
CREATE TABLE IF NOT EXISTS `user_identification` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `subject_type` varchar(16) NOT NULL COMMENT '主体类型',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `uin_status` varchar(16) NOT NULL DEFAULT 'normal' COMMENT '状态',
  `issuer` varchar(128) NOT NULL COMMENT '颁发者',
  `name` varchar(128) DEFAULT NULL COMMENT '用户uin关联名',
  PRIMARY KEY (`id`),
  KEY `idx_user_identification_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1000 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `wecom_app`
--
CREATE TABLE IF NOT EXISTS `wecom_app` (
  `company_id` varchar(18) DEFAULT NULL,
  `app_id` varchar(191) DEFAULT NULL,
  `name` varchar(18) DEFAULT NULL,
  `secret` varchar(64) DEFAULT NULL,
  `token` varchar(32) DEFAULT NULL,
  `encoding_aes_key` varchar(48) DEFAULT NULL,
  UNIQUE KEY `idx_app` (`company_id`,`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `wecom_company`
--
CREATE TABLE IF NOT EXISTS `wecom_company` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(32) DEFAULT NULL,
  `namespace` varchar(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_wecom_company_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `wecom_dept`
--
CREATE TABLE IF NOT EXISTS `wecom_dept` (
  `company_id` varchar(32) DEFAULT NULL,
  `dept_id` bigint DEFAULT NULL,
  `name` varchar(32) DEFAULT NULL,
  `parent_id` bigint DEFAULT NULL,
  `order` int unsigned DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Table structure for table `wecom_rel_dept_user`
--
CREATE TABLE IF NOT EXISTS `wecom_rel_dept_user` (
  `company_id` varchar(32) DEFAULT NULL,
  `dept_id` bigint DEFAULT NULL,
  `user_id` varchar(32) DEFAULT NULL,
  `order` int unsigned DEFAULT NULL,
  `is_leader` tinyint DEFAULT NULL,
  UNIQUE KEY `idx_rel_dept_user` (`company_id`,`dept_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

