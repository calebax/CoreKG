ALTER TABLE ke_application
    ADD COLUMN forest_id BIGINT UNSIGNED DEFAULT NULL AFTER id,
    ADD COLUMN crawl_config JSON DEFAULT NULL AFTER config;

ALTER TABLE ke_web_resource
    ADD COLUMN forest_file_id BIGINT UNSIGNED DEFAULT NULL AFTER id,
    ADD COLUMN raw_content LONGTEXT DEFAULT NULL AFTER metadata;

ALTER TABLE ke_crawl_task
    ADD COLUMN pages_new INT DEFAULT 0 AFTER pages_total,
    ADD COLUMN pages_updated INT DEFAULT 0 AFTER pages_new,
    ADD COLUMN pages_skipped INT DEFAULT 0 AFTER pages_updated;
