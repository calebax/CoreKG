SET NAMES utf8mb4;

ALTER TABLE `ke_article_template`
ADD COLUMN `template_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板类型：system=系统模板，user=用户模板',
ADD COLUMN `source_type` VARCHAR(32) NOT NULL DEFAULT 'manual' COMMENT '来源类型：manual=手动创建，article=基于文章创建',
ADD COLUMN `source_id` BIGINT NOT NULL DEFAULT 0 COMMENT '来源资源id，配合source_type使用';

UPDATE `ke_article_template` SET `template_type`='system' where `template_type`='';

ALTER TABLE `chat_model`
ADD COLUMN `priority` TINYINT NOT NULL DEFAULT 0 COMMENT '优先级，值越大优先级越高';

ALTER TABLE `ke_article`
ADD COLUMN `description` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文章描述';