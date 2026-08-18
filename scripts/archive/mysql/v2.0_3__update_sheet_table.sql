SET NAMES utf8mb4;

ALTER TABLE `ke_forest_excel_sheet`
ADD COLUMN `parent_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '父 Sheet ID（0 表示顶层 Sheet）' AFTER `forest_table_id`,
ADD COLUMN `sheet_type` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'normal' COMMENT 'Sheet 类型：normal=普通 Sheet，sub=内嵌子 Sheet' AFTER `sheet_name`;