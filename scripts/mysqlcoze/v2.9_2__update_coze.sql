SET NAMES utf8mb4;
START TRANSACTION;

-- Create "single_agent_short_link" table
CREATE TABLE `single_agent_short_link` (
    `id` BIGINT UNSIGNED NOT NULL COMMENT 'id',
    `bot_id` BIGINT UNSIGNED NOT NULL COMMENT 'bot id',
    `short_code` VARCHAR(16) NOT NULL COMMENT 'short route code',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'user id',
    `space_id` BIGINT UNSIGNED NOT NULL COMMENT 'space id',
    `user_token` VARCHAR(255) NOT NULL COMMENT 'user token',
    `created_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'create time (ms)',
    `updated_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'update time (ms)',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT 'status 0: active 1: deleted 3: disabled',
    `last_used_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'last used time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_short_code` (`short_code`),
    KEY `idx_bot_id` (`bot_id`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_general_ci
  COMMENT='Bot short link table';

COMMIT;
