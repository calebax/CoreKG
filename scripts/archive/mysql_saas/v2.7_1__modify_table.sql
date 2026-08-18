SET NAMES utf8mb4;

ALTER TABLE admin_license ADD COLUMN version_key VARCHAR(127) NOT NULL DEFAULT 'all' COMMENT '版本类型';