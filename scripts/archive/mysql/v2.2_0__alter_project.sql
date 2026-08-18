SET NAMES utf8mb4;

ALTER TABLE `ke_project`
ADD COLUMN `project_type` varchar(32) NOT NULL DEFAULT 'custom' COMMENT '项目类型：custom：自定义, kb_qa=知识库问答, agent_qa=智能体问答' AFTER `forest_id_list`,
ADD COLUMN `sort` tinyint NOT NULL DEFAULT 0 COMMENT '排序权重：kb_qa=2, agent_qa=1, custom=0' AFTER `project_type`;
