-- excel 相关表
CREATE TABLE `ke_forest_db_instance` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '公司ID',
    `forest_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '知识库ID',
    `ownership_type` VARCHAR(32) NOT NULL DEFAULT 'system' COMMENT '实例归属类型：system-系统自有，external-外部实例',
    `instance_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '数据库类型，oracle：oracle，mysql：mysql',
    `connect_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '数据库连接名称',
    `connect_mode` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '连接模式，standard：标准连接，ssh：ssh隧道模式',
    `host` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '数据库地址',
    `port` INT NOT NULL DEFAULT 0 COMMENT '端口号',
    `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '连接用户名',
    `password` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '连接密码（加密存储）',
    `database` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '默认数据库名',
    `uin` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户uin',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_forest_id` (`forest_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '数据库实例表';

CREATE TABLE `ke_company_db` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '数据库ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '公司ID',
    `db_instance_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据库实例ID',
    `db_name` VARCHAR(255) NOT NULL COMMENT '数据库名',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_unique` (
        `company_id`,
        `db_name`
    ),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_db_instance_id` (`db_instance_id`)
) COMMENT = '公司数据库表';

CREATE TABLE `ke_forest_db` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '数据库ID',
    `company_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '公司ID',
    `forest_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '知识库ID',
    `db_instance_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据库实例ID',
    `db_name` VARCHAR(255) NOT NULL COMMENT '数据库名',
    `db_meta` TEXT NOT NULL COMMENT '数据库元数据，字符集等',
    `size` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据库大小（Bytes）',
    `uin` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户uin',
    `synced_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '同步时间',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_company_id` (`company_id`),
    KEY `idx_forest_id` (`forest_id`),
    KEY `idx_db_instance_id` (`db_instance_id`)
) COMMENT = '数据库信息表';

CREATE TABLE `ke_forest_table` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '数据表ID',
    `forest_id` BIGINT UNSIGNED NOT NULL COMMENT '知识库ID',
    `db_instance_id` BIGINT UNSIGNED NOT NULL COMMENT '数据库实例ID',
    `db_id` BIGINT UNSIGNED NOT NULL COMMENT '数据库ID',
    `table_name` VARCHAR(255) NOT NULL COMMENT '表名',
    `table_meta` TEXT COMMENT '表元数据（如字段结构的JSON描述）',
    `size` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '数据表大小（Bytes）',
    `row_count` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '行数',
    `column_count` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '列数',
    `uin` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户uin',
    `synced_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '同步时间',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_forest_id` (`forest_id`),
    KEY `idx_db_instance_id` (`db_instance_id`),
    KEY `idx_db_id` (`db_id`)
) COMMENT = '知识库数据表信息表';

CREATE TABLE `ke_forest_excel_sheet` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
    `forest_id` BIGINT UNSIGNED NOT NULL COMMENT '知识库ID',
    `forest_file_id` BIGINT UNSIGNED NOT NULL COMMENT '文件ID',
    `forest_table_id` BIGINT UNSIGNED NOT NULL COMMENT '关联数据表ID',
    `sheet_name` VARCHAR(255) NOT NULL COMMENT 'Sheet名称',
    `header_row_num` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '表头行号（从1开始）',
    `data_start_row_num` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '数据开始行号',
    `data_end_row_num` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '数据结束行号',
    `total_row` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总行数（Excel数据行数）',
    `total_column` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总列数（Excel数据列数）',
    `header_mode` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '表头模式，row_title：行表头，column_title：列表头',
    `sheet_meta` TEXT COMMENT 'Sheet元数据（如字段结构的JSON描述）',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间，非 NULL 表示已删除',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_forest_id` (`forest_id`),
    KEY `idx_forest_file_id` (`forest_file_id`),
    KEY `idx_forest_table_id` (`forest_table_id`)
) COMMENT = 'Excel Sheet 与数据表映射';

-- 根据 excel 的 sheet 创建的动态表，用于保存数据行
-- 知识库增加数据源字段
ALTER TABLE
    `ke_forest`
ADD
    COLUMN `data_source_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '数据源类型，standard：标准数据源，excel：excel数据，db：数据库导入'
AFTER
    `forest_type`,
ADD
    COLUMN `data_source_subtype` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '数据源子类型，standard、excel、mysql、pg 等'
AFTER
    `data_source_type`;

-- session 增加 excel 问答相关字段
ALTER TABLE
    `ke_qa_session`
ADD
    COLUMN `excel_id_list` text COLLATE utf8mb4_general_ci COMMENT 'Excel ID 列表'
AFTER
    `forest_id_list`,
ADD
    COLUMN `excel_sheet_id_list` text COLLATE utf8mb4_general_ci COMMENT 'Excel Sheet ID 列表'
AFTER
    `excel_id_list`;

ALTER TABLE
    `ke_qa_session`
ADD
    COLUMN `base_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'standard' COMMENT '基础类型，standard：标准, data_excel：Excel, data_mysql：MySQL '
AFTER
    `company_id`;