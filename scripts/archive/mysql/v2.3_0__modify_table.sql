SET NAMES utf8mb4;


ALTER TABLE ke_forest
ADD COLUMN graph_status VARCHAR(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'uncreated' COMMENT '知识库图谱状态';


RENAME TABLE ke_forest_graph TO ke_forest_graph_version;

ALTER TABLE ke_forest_graph_version
ADD COLUMN graph_id bigint NOT NULL DEFAULT '0' COMMENT '图谱id';

CREATE TABLE `ke_forest_graph` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '图谱名称',
  `description` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '知识森林描述',
  `public_scope` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'company' COMMENT '公开范围',
  `forest_id` bigint NOT NULL DEFAULT '0' COMMENT '''知识库id''',
  `version_id` bigint NOT NULL DEFAULT '0' COMMENT '''图谱版本id''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_forest_graph_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_graph_uin` (`uin`),
  KEY `idx_ke_forest_graph_company_id` (`company_id`),
  KEY `idx_ke_forest_graph_forest_id` (`forest_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;




ALTER TABLE ke_graph_tag
ADD COLUMN graph_version_id bigint NOT NULL DEFAULT '0' COMMENT '图谱版本ID';

ALTER TABLE ke_graph_edge_tag
ADD COLUMN graph_version_id bigint NOT NULL DEFAULT '0' COMMENT '图谱版本ID';

ALTER TABLE ke_graph_node
ADD COLUMN graph_version_id bigint NOT NULL DEFAULT '0' COMMENT '图谱版本ID';

ALTER TABLE ke_graph_tag_node
ADD COLUMN graph_version_id bigint NOT NULL DEFAULT '0' COMMENT '图谱版本ID';

ALTER TABLE ke_graph_edge
ADD COLUMN graph_version_id bigint NOT NULL DEFAULT '0' COMMENT '图谱版本ID';

ALTER TABLE ke_forest_graph
ADD COLUMN avatar_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '文章头像';