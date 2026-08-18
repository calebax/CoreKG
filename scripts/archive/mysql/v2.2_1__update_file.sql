SET NAMES utf8mb4;

ALTER TABLE chat_agent
    ADD COLUMN coze_space_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Coze空间ID',
    ADD COLUMN coze_workflow_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Coze工作流ID',
    ADD COLUMN workflow_version VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Coze工作流版本';
