SET NAMES utf8mb4;

ALTER TABLE ke_graph_tag_node ADD COLUMN created_type VARCHAR(63) NOT NULL DEFAULT 'algorithm' COMMENT '创建类型 algorithm:算法创建,manual:手动创建';