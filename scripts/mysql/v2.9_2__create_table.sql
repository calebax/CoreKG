SET NAMES utf8mb4;

-- 标签分组表
CREATE TABLE IF NOT EXISTS `ke_tag_group` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司 id',
    `name` VARCHAR(64) NOT NULL COMMENT '分组名称',
    `status` VARCHAR(16) NOT NULL DEFAULT 'enable' COMMENT '状态：enable-启用，disable-禁用',
    `created_uin` BIGINT UNSIGNED NOT NULL COMMENT '创建人uin',
    `updated_uin` BIGINT UNSIGNED NOT NULL COMMENT '更新人uin',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '删除时间, NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='标签分组表';

-- 标签表
CREATE TABLE IF NOT EXISTS `ke_tag` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司 id',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '标签分组ID',
    `name` VARCHAR(64) NOT NULL COMMENT '标签名称',
    `status` VARCHAR(16) NOT NULL DEFAULT 'enable' COMMENT '状态：enable-启用，disable-禁用',
    `created_uin` BIGINT UNSIGNED NOT NULL COMMENT '创建人uin',
    `updated_uin` BIGINT UNSIGNED NOT NULL COMMENT '更新人uin',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '删除时间, NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_group_id` (`group_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='标签表';

-- 文档标签关联表
CREATE TABLE IF NOT EXISTS `ke_resource_tag` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司 id',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '标签分组ID',
    `resource_type` VARCHAR(127) NOT NULL COMMENT '资源类型，file：知识库文件',
    `resource_id` BIGINT UNSIGNED NOT NULL COMMENT '资源 id',
    `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
    `created_uin` BIGINT UNSIGNED NOT NULL COMMENT '创建人uin',
    `updated_uin` BIGINT UNSIGNED NOT NULL COMMENT '更新人uin',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '删除时间, NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_resource` (`resource_type`, `resource_id`),
    KEY `idx_tag_id` (`tag_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='资源标签关联表';

-- 用户最近使用标签记录表
CREATE TABLE IF NOT EXISTS `ke_recent_used_tag` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司 id',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '标签分组ID',
    `uin` BIGINT UNSIGNED NOT NULL COMMENT '用户uin',
    `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
    `last_used_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '最后使用时间',
    `usage_count` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '使用次数',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at` DATETIME(3) DEFAULT NULL COMMENT '删除时间, NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_uin` (`uin`),
    KEY `idx_last_used_at` (`last_used_at`),
    KEY `idx_tag_id` (`tag_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户最近使用标签记录表';