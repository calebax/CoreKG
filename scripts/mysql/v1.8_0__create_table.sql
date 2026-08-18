SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `chat_question_db_dataset` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `request_id` varchar(128) NOT NULL DEFAULT '' COMMENT '请求ID',
  `question_id` varchar(128) NOT NULL DEFAULT '' COMMENT '问题ID',
  `session_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '会话ID',
  `database_type` varchar(32) NOT NULL COMMENT '数据库类型，mysql',
  `table_list` longtext DEFAULT NULL COMMENT '相关表',
  `query_statement` text COMMENT '生成的查询语句（SQL/NoSQL/API等）',
  `query_result` longtext DEFAULT NULL COMMENT '查询执行结果数据集',
  `echarts_config` longtext DEFAULT NULL COMMENT 'echarts图表配置',
  `echarts_dataset` longtext DEFAULT NULL COMMENT 'echarts数据集',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_request_id` (`request_id`),
  KEY `idx_question_id` (`question_id`),
  KEY `idx_session_id` (`session_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库问答数据集';