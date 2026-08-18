SET NAMES utf8mb4;

CREATE TABLE `ke_graph_resource_chunk` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `graph_id` bigint NOT NULL DEFAULT 0 COMMENT '图谱id',
  `graph_version_id` bigint NOT NULL DEFAULT 0 COMMENT '图谱版本id',
  `chunk_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'chunk id',
  `resource_type` varchar(24) NOT NULL DEFAULT '' COMMENT '资源类型，node：图谱节点，edge：图谱边',
  `resource_id` bigint NOT NULL DEFAULT 0 COMMENT '资源id',
  PRIMARY KEY (`id`),
  KEY `idx_graph_id` (`graph_id`),
  KEY `idx_graph_version_id` (`resource_type`, `resource_id`),
  KEY `idx_resource_id` (`resource_type`, `resource_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='chunk 和图谱节点、边的关系';