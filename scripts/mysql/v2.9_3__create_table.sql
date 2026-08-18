CREATE TABLE `ke_keywords` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `subject_id` bigint NOT NULL COMMENT 'subject id',
  `word_type` varchar(32) NOT NULL COMMENT '词典类型 synonym:同义词；major：专业',
  `description` varchar(255) DEFAULT NULL COMMENT 'description 专业名词描述',
  `word` varchar(255) DEFAULT NULL COMMENT 'word 词内容',
  PRIMARY KEY (`id`),
  KEY `idx_core_task_deleted_at` (`deleted_at`),
  KEY `idx_core_task_uin` (`uin`),
  KEY `idx_core_task_company_id` (`company_id`)
) ENGINE=InnoDB AUTO_INCREMENT=134 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;