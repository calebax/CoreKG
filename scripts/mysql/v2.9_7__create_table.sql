SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `ke_uin_likes`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',

    `uin`           BIGINT UNSIGNED NOT NULL COMMENT '用户uin',
    `company_id`    BIGINT UNSIGNED NOT NULL COMMENT '公司id',

    `resource_id`   BIGINT UNSIGNED NOT NULL COMMENT '资源id',
    `resource_type` VARCHAR(255)         NOT NULL COMMENT '资源类型',

    `created_at`    datetime(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`    datetime(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at`    datetime(3)              DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_target` (`resource_id`, `resource_type`)
) ENGINE = InnoDB COMMENT ='点赞表';


CREATE TABLE IF NOT EXISTS `ke_uin_collections`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',

    `uin`           BIGINT UNSIGNED NOT NULL COMMENT '用户uin',
    `company_id`    BIGINT UNSIGNED NOT NULL COMMENT '公司id',

    `resource_id`   BIGINT UNSIGNED NOT NULL COMMENT '资源id',
    `resource_type` VARCHAR(255)         NOT NULL COMMENT '资源类型',

    `created_at`    datetime(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`    datetime(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at`    datetime(3)              DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_target` (`resource_id`, `resource_type`)
) ENGINE = InnoDB COMMENT ='收藏表';
