SET NAMES utf8mb4;

START TRANSACTION;

# Add enable column to ke_forest_file, ke_forest_db
ALTER TABLE `ke_forest_file`
    ADD COLUMN `enable` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用';

ALTER TABLE `ke_forest_db`
    ADD COLUMN `enable` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用';

ALTER TABLE `ke_forest_excel_sheet`
    ADD COLUMN `enable` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用';

ALTER TABLE `ke_forest_table`
    ADD COLUMN `enable` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用';

COMMIT;
