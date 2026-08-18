SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `ke_message_template` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` VARCHAR(64) DEFAULT NULL COMMENT '模板名称，如公告、下单等，具有唯一性',
    `description` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '模板描述',
    `type` VARCHAR(32) NOT NULL COMMENT '模板类型：system-系统消息，announcement-公告消息',
    `title_template` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '标题模板，支持 {{variable}} 占位符',
    `content_template` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '内容模板，支持 {{variable}} 占位符',
    `module` VARCHAR(64) DEFAULT NULL DEFAULT '' COMMENT '功能模块，表示所属哪个功能模块',
    `route_path` VARCHAR(256) DEFAULT NULL DEFAULT '' COMMENT '前端路由路径模板，支持 {{variable}} 占位符，如：/order/detail?id={{order_id}}',
    `status` VARCHAR(16) NOT NULL DEFAULT 'draft' COMMENT '模板状态：draft-草稿，enable-启用，disable-禁用',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间, NULL表示未删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_module` (`module`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='消息模板表';

CREATE TABLE IF NOT EXISTS `ke_uin_message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属公司ID',
    `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    `uin` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'uin',
    `template_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '模板ID',
    `title` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '消息标题',
    `content` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '渲染后的消息内容',
    `template_type` VARCHAR(32) NOT NULL COMMENT '模板类型：system-系统消息，announcement-公告消息',
    `source_type` VARCHAR(64) DEFAULT NULL DEFAULT '' COMMENT '业务关联类型',
    `source_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '业务关联ID',
    `route_path` VARCHAR(256) DEFAULT NULL DEFAULT '' COMMENT '实际跳转路由路径',
    `read_status` VARCHAR(16) NOT NULL DEFAULT 'unread' COMMENT '已读状态：unread-未读，read-已读',
    `read_at` datetime(3) DEFAULT NULL COMMENT '阅读时间',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间, NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_uin` (`uin`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_template_id` (`template_id`),
    KEY `idx_source` (`source_type`, `source_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户消息表';