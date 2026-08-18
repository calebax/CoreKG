SET NAMES utf8mb4;

-- UPDATE `chat_agent` SET `name` = `` WHERE `name` = ``;


ALTER TABLE ke_forest_file ADD COLUMN file_config TEXT COLLATE utf8mb4_general_ci COMMENT '文件配置';

-- session 增加 excel 问答相关字段
ALTER TABLE  `chat_sessions` ADD COLUMN `db_list` text COLLATE utf8mb4_general_ci COMMENT '数据库名称列表';
ALTER TABLE  `chat_sessions` ADD COLUMN `excel_sheet_id_list` text COLLATE utf8mb4_general_ci COMMENT 'excel sheet id 列表';
ALTER TABLE  `chat_sessions` ADD  COLUMN `db_table_list` text COLLATE utf8mb4_general_ci COMMENT '数据表名称列表';
ALTER TABLE  `chat_sessions` ADD  COLUMN `base_type` varchar(32) DEFAULT 'standard' COLLATE utf8mb4_general_ci COMMENT '基础类型，standard：标准, data_excel：Excel, data_mysql：MySQL';
ALTER TABLE  `chat_sessions` ADD  COLUMN `excel_id_list` text COLLATE utf8mb4_general_ci COMMENT 'excelIDList';

-- 1. 创建只读用户（仅在不存在时创建）
CREATE USER IF NOT EXISTS 'readonly'@'%' IDENTIFIED BY 'gTxG6Kgq4YmDCji2';

-- 2. 授权全库只读权限
GRANT SELECT, SHOW VIEW ON *.* TO 'readonly'@'%';

-- 3. 刷新权限
FLUSH PRIVILEGES;

ALTER TABLE
    `ke_forest_db_instance`
ADD
    COLUMN `connection_status` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'valid' COMMENT '连接状态，valid：有效，invalid：无效'
AFTER
    `database`;


ALTER TABLE
    `ke_forest_db`
ADD
    `row_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '行数'
AFTER
    `size`;

INSERT INTO `core_settings` (`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES
	('2025-08-29 09:19:50.000', '2025-08-29 09:19:50.000', NULL, 'knowledge', 'mysql_excel_instance_readonly', 'excel问答 mysql 只读实例信息', NULL, 'yaml', 'host: mysql\nport: 3306\nusername: readonly\npassword: gTxG6Kgq4YmDCji2\ncharset: utf8mb4\n', NULL);
