SET NAMES utf8mb4;

-- 新增 type 字段，区分文章和模板
ALTER TABLE `ke_article`
    ADD COLUMN `type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'article'
        COMMENT '类型：article=普通文章，template_system=系统模板，template_user=用户模板'
        AFTER `deleted_at`;

-- 新增模板来源字段（原 template 表的 source_type / source_id）
ALTER TABLE `ke_article`
    ADD COLUMN `source_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'manual'
        COMMENT '记录来源：manual=手动创建，article=基于文章创建，template=基于模板创建'
        AFTER `type`,
    ADD COLUMN `source_id` bigint NOT NULL DEFAULT '0'
        COMMENT '来源资源id，配合 source_type 使用'
        AFTER `source_type`;

ALTER TABLE `ke_article_history`
    ADD COLUMN `article_id` bigint NOT NULL DEFAULT '0'
        COMMENT '关联文章id（ke_article.id）'
        AFTER `deleted_at`,
    ADD INDEX `idx_article_id` (`article_id`);