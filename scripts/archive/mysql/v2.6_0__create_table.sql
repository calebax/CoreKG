SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `admin_announcement`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `uin`        bigint          NOT NULL COMMENT '创建人uin',
    `company_id` bigint          NOT NULL DEFAULT 0 COMMENT '公司ID',
    `creator`    varchar(511)    NOT NULL COMMENT '创建人昵称',
    `tag`        varchar(127)    NOT NULL COMMENT '版本tag',
    `content`    longtext COMMENT '公告内容',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
    `deleted_at` datetime                 DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',

    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='发版公告';


