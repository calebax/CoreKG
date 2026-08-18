SET NAMES utf8mb4;
START TRANSACTION;

-- 设置 yygu 数据库（存在其一）
SET @target_yygu_db := (
    SELECT `schema_name`
    FROM `information_schema`.`schemata`
    WHERE `schema_name` IN ('yygu_test', 'yygu_db')
    ORDER BY FIELD(`schema_name`, 'yygu_test', 'yygu_db')
    LIMIT 1
);

-- 逐条处理记录。操作完成后更新 sync_flag。
DROP PROCEDURE IF EXISTS `sp_history_data_migration_process`;
DELIMITER //
CREATE PROCEDURE `sp_history_data_migration_process`()
proc: BEGIN
    DECLARE v_last_id BIGINT UNSIGNED DEFAULT 0;
    DECLARE v_user_id BIGINT UNSIGNED;
    DECLARE v_old_space_id BIGINT UNSIGNED;
    DECLARE v_new_space_id BIGINT UNSIGNED;

    IF @target_yygu_db IS NULL THEN
        SELECT 'No target database (yygu_test/yygu_db) found' AS message;
        LEAVE proc;
    END IF;

    read_loop: LOOP
        SET v_user_id = NULL;
        SET v_old_space_id = NULL;
        SET v_new_space_id = NULL;

        SELECT `id`, `old_space_id`, `space_id`
        INTO v_user_id, v_old_space_id, v_new_space_id
        FROM `history_data_migration_sync_record`
        WHERE `sync_flag` = 0 AND `id` > v_last_id
        ORDER BY `id` ASC
        LIMIT 1;

        IF v_user_id IS NULL THEN
            LEAVE read_loop;
        END IF;

        -- 1) 根据 oldSpaceId 查询 workflow_meta
        -- 2) 插入 ke_resource_scope 两条记录（view/manage）
        -- 3) 更新 workflow_meta.space_id 为 newSpaceId
        SET @insert_scope_sql := CONCAT(
            'INSERT INTO `', @target_yygu_db, '`.`ke_resource_scope` ',
            '(`resource_type`, `resource_id`, `scope_type`, `scope_id`, `action`) ',
            'SELECT ''workflow'', wm.`id`, ''user'', ', v_user_id, ', ''view'' ',
            'FROM `workflow_meta` wm ',
            'WHERE wm.`space_id` = ', v_old_space_id, ' AND wm.`creator_id` = ', v_user_id, ' ',
            'UNION ALL ',
            'SELECT ''workflow'', wm.`id`, ''user'', ', v_user_id, ', ''manage'' ',
            'FROM `workflow_meta` wm ',
            'WHERE wm.`space_id` = ', v_old_space_id, ' AND wm.`creator_id` = ', v_user_id
        );
        PREPARE stmt FROM @insert_scope_sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;

        UPDATE `workflow_meta`
        SET `space_id` = v_new_space_id
        WHERE `space_id` = v_old_space_id AND `creator_id` = v_user_id;

        -- 处理 single_agent_draft：插入 ke_resource_scope（view/manage）并更新 space_id
        SET @insert_agent_scope_sql := CONCAT(
            'INSERT INTO `', @target_yygu_db, '`.`ke_resource_scope` ',
            '(`resource_type`, `resource_id`, `scope_type`, `scope_id`, `action`) ',
            'SELECT ''agent'', sad.`agent_id`, ''user'', ', v_user_id, ', ''view'' ',
            'FROM `single_agent_draft` sad ',
            'WHERE sad.`space_id` = ', v_old_space_id, ' AND sad.`creator_id` = ', v_user_id, ' ',
            'UNION ALL ',
            'SELECT ''agent'', sad.`agent_id`, ''user'', ', v_user_id, ', ''manage'' ',
            'FROM `single_agent_draft` sad ',
            'WHERE sad.`space_id` = ', v_old_space_id, ' AND sad.`creator_id` = ', v_user_id
        );
        PREPARE stmt FROM @insert_agent_scope_sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;

        UPDATE `single_agent_draft`
        SET `space_id` = v_new_space_id
        WHERE `space_id` = v_old_space_id AND `creator_id` = v_user_id;

        UPDATE `history_data_migration_sync_record`
        SET `sync_flag` = 1,
            `migration_completed_time` = CAST(UNIX_TIMESTAMP(NOW(3)) * 1000 AS UNSIGNED)
        WHERE `id` = v_user_id;

        SET v_last_id = v_user_id;
    END LOOP;
END//
DELIMITER ;

CALL `sp_history_data_migration_process`();
DROP PROCEDURE IF EXISTS `sp_history_data_migration_process`;

COMMIT;
