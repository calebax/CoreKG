SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `core_daily_log`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `date`          datetime(3)     NOT NULL COMMENT "记录日期",
    `previous_hash` varchar(511) DEFAULT '' COMMENT "前置hash",
    `current_hash`  varchar(511) DEFAULT '' COMMENT "当前hash",
    `valid`         tinyint(1)   NOT NULL DEFAULT -1 COMMENT "有效状态",
    `message`       text COMMENT "备注信息",
    `created_at`    datetime(3)  DEFAULT NULL,
    `updated_at`    datetime(3)  DEFAULT NULL,
    `deleted_at`    datetime(3)  DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_date` (`date`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 1
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='日志对象表';

CREATE TABLE IF NOT EXISTS `chat_migrate` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `resource_type` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '''资源类型''',
  `resource_id` bigint NOT NULL DEFAULT '0' COMMENT '''资源ID''',
  `target_id` bigint NOT NULL DEFAULT '0' COMMENT '''目标ID''',
  `target_id_str` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '''目标ID''',
  PRIMARY KEY (`id`),
  KEY `idx_chat_migrate_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=4735 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;