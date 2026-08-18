SET NAMES utf8mb4;

ALTER TABLE `ke_graph_tag_node`
ADD COLUMN `uin` bigint NOT NULL DEFAULT '0' COMMENT '用户ID',
ADD COLUMN `company_id` bigint NOT NULL DEFAULT '0' COMMENT '公司ID',
ADD COLUMN `name` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '节点ID';

ALTER TABLE ke_graph_tag_node ADD INDEX idx_ke_graph_tag_node_uin (`uin`);
ALTER TABLE ke_graph_tag_node ADD INDEX idx_ke_graph_tag_node_company_id (`company_id`);
ALTER TABLE ke_graph_tag_node ADD INDEX idx_ke_graph_tag_node_name (`name`);

ALTER TABLE ke_graph_tag_node
    DROP INDEX idx_tag_node;

UPDATE ke_graph_tag_node t
        JOIN ke_graph_node n ON t.node_id = n.id
        SET 
            t.name = n.name,
            t.company_id = n.company_id,
			t.uin = n.uin
        WHERE t.id IN (
            SELECT id FROM (
                SELECT t2.id
                FROM ke_graph_tag_node t2
                WHERE t2.name = ''
            ) AS tmp
        );

ALTER TABLE ke_graph_tag_node
    ADD UNIQUE INDEX idx_tag_node (tag_id, name);


ALTER TABLE chat_sessions ADD prompt_mode varchar(32) NOT NULL DEFAULT '' COMMENT 'prompt模式';