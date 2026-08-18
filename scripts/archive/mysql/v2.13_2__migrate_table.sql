SET NAMES utf8mb4;

INSERT INTO `ke_article` (
    `created_at`, `updated_at`, `deleted_at`,
    `type`,
    `title`, `description`, `content`,
    `avatar_url`,
    `company_id`, `uin`,
    `source_type`, `source_id`,
    `template_id`,
    `forest_ids`, `public_scope`
)
SELECT
    `created_at`, `updated_at`, `deleted_at`,

    CASE `template_type`
        WHEN 'system' THEN 'template_system'
        WHEN 'user'   THEN 'template_user'
        ELSE 'template_user'
    END                         AS `type`,

    `name`                      AS `title`,
    COALESCE(`description`, '') AS `description`,
    `content`,
    ''                          AS `avatar_url`,
    `company_id`, `uin`,
    `source_type`,
    `source_id`,
    0                           AS `template_id`,
    NULL                        AS `forest_ids`,
    'company'                   AS `public_scope`
FROM `ke_article_template`
WHERE `deleted_at` IS NULL;        -- 只迁移未软删除的记录


-- 验证
SELECT
    (SELECT COUNT(*) FROM `ke_article_template` WHERE deleted_at IS NULL) AS template_cnt,
    (SELECT COUNT(*) FROM `ke_article` WHERE type IN ('template_system', 'template_user')) AS migrated_cnt;