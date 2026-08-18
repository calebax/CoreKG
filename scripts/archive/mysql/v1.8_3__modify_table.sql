SET NAMES utf8mb4;


ALTER TABLE `ke_forest_graph`
    ADD COLUMN `public_scope` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'company' COMMENT '公开范围';
