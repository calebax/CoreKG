SET NAMES utf8mb4;

CREATE TABLE `ke_forest_graph` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '图谱名称',
  `status` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'draft' COMMENT '知识图谱状态',
  `description` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '知识森林描述',
  `file_id_list` text COLLATE utf8mb4_general_ci COMMENT '''文件ID列表''',
  `space_name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '知识图谱空间名称',
  `parse_mode` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'auto' COMMENT '解析模式',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_ke_forest_graph_space_name` (`space_name`),
  KEY `idx_ke_forest_graph_deleted_at` (`deleted_at`),
  KEY `idx_ke_forest_graph_uin` (`uin`),
  KEY `idx_ke_forest_graph_company_id` (`company_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;



CREATE TABLE `ke_graph_tag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `graph_id` bigint NOT NULL DEFAULT '0' COMMENT '''图谱ID''',
  `tag_name` varchar(256) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '''标签名称''',
  `tag_type` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT '''标签类型 tag',
  `properties` mediumtext COLLATE utf8mb4_general_ci COMMENT '''标签属性''',
  `description` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '知识森林描述',
  `tag_status` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'not_synced' COMMENT '''标签状态''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_graph_tag_deleted_at` (`deleted_at`),
  KEY `idx_ke_graph_tag_uin` (`uin`),
  KEY `idx_ke_graph_tag_company_id` (`company_id`),
  KEY `idx_ke_graph_tag_graph_id` (`graph_id`),
  KEY `idx_ke_graph_tag_tag_name` (`tag_name`),
  KEY `idx_ke_graph_tag_tag_type` (`tag_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;



CREATE TABLE `ke_graph_node` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `uin` bigint NOT NULL DEFAULT '0' COMMENT '''用户ID''',
  `company_id` bigint NOT NULL DEFAULT '0' COMMENT '''公司ID''',
  `graph_id` bigint NOT NULL DEFAULT '0' COMMENT '''图谱ID''',
  `name` varchar(512) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '''节点ID''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_graph_node_deleted_at` (`deleted_at`),
  KEY `idx_ke_graph_node_uin` (`uin`),
  KEY `idx_ke_graph_node_company_id` (`company_id`),
  KEY `idx_ke_graph_node_graph_id` (`graph_id`),
  KEY `idx_ke_graph_node_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


CREATE TABLE `ke_graph_tag_node` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `tag_id` bigint NOT NULL DEFAULT '0' COMMENT '''标签ID''',
  `node_id` bigint NOT NULL DEFAULT '0' COMMENT '''节点ID''',
  `file_id_list` text COLLATE utf8mb4_general_ci COMMENT '''文件ID列表''',
  `chunk_id_list` text COLLATE utf8mb4_general_ci COMMENT '''分块ID列表''',
  `properties_values` mediumtext COLLATE utf8mb4_general_ci COMMENT '''标签属性''',
  `graph_id` bigint NOT NULL DEFAULT '0' COMMENT '''图谱ID''',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_tag_node` (`tag_id`,`node_id`),
  KEY `idx_ke_graph_tag_node_deleted_at` (`deleted_at`),
  KEY `idx_ke_graph_tag_node_graph_id` (`graph_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


CREATE TABLE `ke_graph_edge` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `graph_id` bigint NOT NULL DEFAULT '0' COMMENT '''图谱ID''',
  `src_id` bigint NOT NULL DEFAULT '0' COMMENT '''起始节点ID''',
  `dst_id` bigint NOT NULL DEFAULT '0' COMMENT '''结束节点ID''',
  `src_tag_id` bigint NOT NULL DEFAULT '0' COMMENT '''起始节点标签ID''',
  `dst_tag_id` bigint NOT NULL DEFAULT '0' COMMENT '''结束节点标签ID''',
  `tag_id` bigint NOT NULL DEFAULT '0' COMMENT '''标签ID''',
  `file_id_list` text COLLATE utf8mb4_general_ci COMMENT '''文件ID列表''',
  `chunk_id_list` text COLLATE utf8mb4_general_ci COMMENT '''分块ID列表''',
  `properties_values` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT '''标签属性''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_graph_edge_deleted_at` (`deleted_at`),
  KEY `idx_ke_graph_edge_graph_id` (`graph_id`),
  KEY `idx_ke_graph_edge_src_id` (`src_id`),
  KEY `idx_ke_graph_edge_dst_id` (`dst_id`),
  KEY `idx_ke_graph_edge_src_tag_id` (`src_tag_id`),
  KEY `idx_ke_graph_edge_dst_tag_id` (`dst_tag_id`),
  KEY `idx_ke_graph_edge_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


CREATE TABLE `ke_graph_edge_tag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `graph_id` bigint NOT NULL DEFAULT '0' COMMENT '''图谱ID''',
  `edge_type_id` bigint NOT NULL DEFAULT '0' COMMENT '''边类型ID''',
  `src_tag_id` bigint NOT NULL DEFAULT '0' COMMENT '''起始节点标签ID''',
  `dst_tag_id` bigint NOT NULL DEFAULT '0' COMMENT '''结束节点标签ID''',
  PRIMARY KEY (`id`),
  KEY `idx_ke_graph_edge_tag_graph_id` (`graph_id`),
  KEY `idx_ke_graph_edge_tag_edge_type_id` (`edge_type_id`),
  KEY `idx_ke_graph_edge_tag_src_tag_id` (`src_tag_id`),
  KEY `idx_ke_graph_edge_tag_dst_tag_id` (`dst_tag_id`),
  KEY `idx_ke_graph_edge_tag_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;