SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `account_external_binding` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uin` bigint unsigned NOT NULL COMMENT '系统用户唯一标识',
  `platform` varchar(50) NOT NULL COMMENT '第三方平台标识，如 github/google/slack',
  `provider` varchar(50) DEFAULT NULL COMMENT '平台下的服务/提供者，如 gmail/drive',
  `external_id` varchar(100) NOT NULL COMMENT '第三方平台用户ID',
  `access_token` text NOT NULL COMMENT '加密存储的 access_token',
  `refresh_token` text DEFAULT NULL COMMENT '加密存储的 refresh_token',
  `expires_at` datetime(3) DEFAULT NULL COMMENT 'access_token 过期时间',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '绑定状态: 1=绑定 0=解绑/失效',
  `email` varchar(255) DEFAULT NULL COMMENT '第三方平台邮箱',
  `avatar` varchar(500) DEFAULT NULL COMMENT '第三方平台头像 URL',
  `created_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `company_id` bigint unsigned DEFAULT NULL COMMENT '公司ID/组织ID',
  PRIMARY KEY (`id`),
  KEY `idx_uin_provider` (`uin`,`provider`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户第三方平台授权绑定记录表';

CREATE TABLE `core_jobs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID，自增',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '软删除时间',
  `job_uuid` varchar(36) COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务唯一标识符',
  `company_id` bigint NOT NULL COMMENT '公司ID',
  `uin` bigint NOT NULL COMMENT '用户标识符',
  `purpose` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务用途/目的',
  `job_status` varchar(20) COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务状态（如：pending/running/completed/failed）',
  `cost_seconds` bigint NOT NULL COMMENT '任务执行耗时（秒）',
  `output` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '任务输出结果',
  `error_msg` varchar(1024) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '错误信息',
  `extra` text COLLATE utf8mb4_general_ci COMMENT '额外信息（JSON格式）',
  PRIMARY KEY (`id`),
  KEY `idx_core_jobs_purpose` (`purpose`) COMMENT '任务用途索引',
  KEY `idx_core_jobs_job_status` (`job_status`) COMMENT '任务状态索引',
  KEY `idx_core_jobs_deleted_at` (`deleted_at`) COMMENT '软删除索引',
  KEY `idx_core_jobs_job_uuid` (`job_uuid`) COMMENT '任务UUID索引',
  KEY `idx_core_jobs_company_id` (`company_id`) COMMENT '公司ID索引',
  KEY `idx_core_jobs_uin` (`uin`) COMMENT '用户标识符索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='核心任务表，用于记录系统中各类异步任务的执行情况';

CREATE TABLE IF NOT EXISTS `ke_article`
(
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `title`        VARCHAR(255)    NOT NULL COMMENT '文章标题',
    `content`      MEDIUMTEXT   DEFAULT NULL COMMENT '文章详细内容',
    `public_scope` VARCHAR(63)  DEFAULT 'company' COMMENT '资源公开范围',
    `company_id`   BIGINT       DEFAULT 0 COMMENT '公司id',
    `uin`          BIGINT       DEFAULT 0 COMMENT 'uin',
    `template_id`  BIGINT       DEFAULT 0 COMMENT '模板id',
    `forest_ids`   VARCHAR(511) DEFAULT NULL COMMENT '知识库列表id',
    `created_at`   datetime(3)  DEFAULT NULL,
    `updated_at`   datetime(3)  DEFAULT NULL,
    `deleted_at`   datetime(3)  DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='写作空间文章表';

CREATE TABLE IF NOT EXISTS `ke_article_template`
(
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`       VARCHAR(255)    NOT NULL COMMENT '模板名',
    `description` VARCHAR(511) DEFAULT NULL COMMENT '描述',
    `content`    MEDIUMTEXT  DEFAULT NULL COMMENT '模板内容',
    `company_id` BIGINT      DEFAULT 0 COMMENT '公司id',
    `uin`        BIGINT      DEFAULT 0 COMMENT 'uin',
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='写作空间文章模板表';

CREATE TABLE IF NOT EXISTS `ke_article_history`
(
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `cmd` VARCHAR(127) NOT NULL COMMENT '撰写命令',
    `content`    MEDIUMTEXT  DEFAULT NULL COMMENT '原始内容',
    `result`    MEDIUMTEXT  DEFAULT NULL COMMENT '撰写结果',
    `company_id` BIGINT      DEFAULT 0 COMMENT '公司id',
    `uin`        BIGINT      DEFAULT 0 COMMENT 'uin',
    `created_at` datetime(3) DEFAULT NULL,
    `updated_at` datetime(3) DEFAULT NULL,
    `deleted_at` datetime(3) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='写作空间撰写历史';


CREATE TABLE `ke_project` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '项目名',
  `forest_id_list` text COLLATE utf8mb4_general_ci COMMENT '''森林ID列表''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_project_deleted_at` (`deleted_at`),
  KEY `idx_ke_project_uin` (`uin`),
  KEY `idx_ke_project_company_id` (`company_id`)
) ENGINE=InnoDB AUTO_INCREMENT=95 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


-- 同时添加两个字段和索引
ALTER TABLE `chat_sessions` 
ADD COLUMN `subject_id` bigint NOT NULL DEFAULT '0' COMMENT '主体id',
ADD COLUMN `external_token_id_list` text COLLATE utf8mb4_general_ci COMMENT '外部数据源id',
ADD KEY `idx_chat_sessions_subject_id` (`subject_id`);

CREATE TABLE `chat_chart` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `request_id` varchar(128) NOT NULL DEFAULT '' COMMENT '请求ID',
  `question_id` varchar(128) NOT NULL DEFAULT '' COMMENT '问题ID',
  `session_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '会话ID',
  `subject_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '主体 id',
  `subject_type` varchar(32) NOT NULL DEFAULT '' COMMENT '主体类型',
  `chart_type` varchar(32) NOT NULL DEFAULT '' COMMENT '图表类型',
  `chart_content` longtext COMMENT '内容',
  `company_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '公司ID',
  `uin` bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户uin',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_company_id` (`company_id`),
  KEY `idx_uin` (`uin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='chart图表';

CREATE TABLE `chat_chart_canvas` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `subject_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '主体 id',
  `subject_type` varchar(32) NOT NULL DEFAULT '' COMMENT '主体类型',
  `content` longtext COLLATE utf8mb4_general_ci NOT NULL COMMENT '画布内容(json)',
  `company_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '公司ID',
  `uin` bigint unsigned NOT NULL DEFAULT '0' COMMENT '用户uin',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
  PRIMARY KEY (`id`),
  KEY `idx_deleted_at` (`deleted_at`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_company_id` (`company_id`),
  KEY `idx_uin` (`uin`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='chart画布表';
