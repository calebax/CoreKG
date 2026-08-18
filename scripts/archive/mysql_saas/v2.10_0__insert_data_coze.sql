SET NAMES utf8mb4;
START TRANSACTION;

-- 在 yygu_test 或 yygu_db（存在其一）中创建 history_data_migration_sync_record 表
SET @target_yygu_db := (
    SELECT `schema_name`
    FROM `information_schema`.`schemata`
    WHERE `schema_name` IN ('yygu_test', 'yygu_db')
    ORDER BY FIELD(`schema_name`, 'yygu_test', 'yygu_db')
    LIMIT 1
);

SET @target_coze_db := (
    SELECT `schema_name`
    FROM `information_schema`.`schemata`
    WHERE `schema_name` IN ('opencoze_test', 'opencoze')
    ORDER BY FIELD(`schema_name`, 'opencoze_test', 'opencoze')
    LIMIT 1
);

SET @create_sql := IF(
    @target_coze_db IS NULL,
    'SELECT ''No target database (opencoze_test/opencoze) found'' AS message',
    CONCAT(
    'CREATE TABLE IF NOT EXISTS `', @target_coze_db, '`.`history_data_migration_sync_record` (',
    ' `id` BIGINT UNSIGNED NOT NULL COMMENT ''用户ID'',',
    ' `old_space_id` BIGINT UNSIGNED NOT NULL COMMENT ''旧空间ID'',',
    ' `space_id` BIGINT UNSIGNED NOT NULL COMMENT ''新空间ID'',',
    ' `sync_flag` TINYINT NOT NULL DEFAULT 0 COMMENT ''同步标记 0:未同步 1:已同步'',',
    ' `migration_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''迁移时间（毫秒）'',',
    ' `migration_completed_time` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''迁移完成时间（毫秒）'',',
    ' PRIMARY KEY (`id`),',
    ' KEY `idx_space_id` (`space_id`)',
    ') ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT=''历史数据迁移同步记录'''
    )
);

PREPARE stmt FROM @create_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 汇总待处理记录：
-- oldSpaceId = space_user.space_id，userId = space_user.user_id，
-- newSpaceId = user_identification.subject_id（匹配条件：id 与 subject_type=company）。
SET @query_sql := IF(
    @target_coze_db IS NULL OR @target_yygu_db IS NULL,
    'SELECT ''No target database (yygu_test/yygu_db or opencoze_test/opencoze) found'' AS message',
    CONCAT(
        'INSERT INTO `', @target_coze_db, '`.`history_data_migration_sync_record` ',
        '(`id`, `old_space_id`, `space_id`, `sync_flag`, `migration_time`, `migration_completed_time`) ',
        'SELECT su.`user_id` AS userId, su.`space_id` AS oldSpaceId, ui.`subject_id` AS newSpaceId, 0, 0, 0 ',
        'FROM `', @target_coze_db, '`.`space_user` su ',
        'JOIN `', @target_yygu_db, '`.`user_identification` ui ',
        '  ON ui.`id` = su.`user_id` AND ui.`subject_type` = ''company'' ',
        'LEFT JOIN `', @target_yygu_db, '`.`company` c ON c.`id` = su.`space_id` ',
        'WHERE c.`id` IS NULL'
    )
);

PREPARE stmt FROM @query_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

COMMIT;
