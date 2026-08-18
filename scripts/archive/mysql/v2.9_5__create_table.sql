CREATE TABLE `ke_forest_hot_word` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `hot_words` text COMMENT '''热词列表''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_hot_word_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_hot_word_uin` (`uin`),
  KEY `idx_ke_forest_hot_word_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;