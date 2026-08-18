SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS chat_coze_mapping (
                                                 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                                 created_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) DEFAULT NULL,
    uin BIGINT NOT NULL DEFAULT 0 COMMENT 'uin',
    type VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '类型',
    corekg_id BIGINT NOT NULL DEFAULT 0 COMMENT 'corekg侧ID',
    coze_id BIGINT NOT NULL DEFAULT 0 COMMENT 'coze侧ID',
    PRIMARY KEY (id),
    UNIQUE KEY uniq_mapping (uin, type, corekg_id),
    INDEX idx_corekg (uin, type, corekg_id),
    INDEX idx_coze (uin, type, coze_id)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;