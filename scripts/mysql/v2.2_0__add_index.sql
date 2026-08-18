SET NAMES utf8mb4;

ALTER TABLE ke_resource_scope
ADD INDEX idx_scope (`resource_type`, `scope_type`, `scope_id`);


ALTER TABLE ke_forest_file
ADD INDEX idx_forest_id_status (`forest_id`, `status`, `is_dir`);

ALTER TABLE ke_article
ADD INDEX idx_update_at (`updated_at`),
ADD INDEX idx_company (`company_id`),
ADD INDEX idx_uin (`uin`);