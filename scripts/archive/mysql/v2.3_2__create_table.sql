SET NAMES utf8mb4;

CREATE TABLE `admin_api_sample_data` (
       `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
       `created_at` DATETIME(3) NULL DEFAULT NULL COMMENT '创建时间',
       `updated_at` DATETIME(3) NULL DEFAULT NULL COMMENT '更新时间',
       `deleted_at` DATETIME(3) NULL DEFAULT NULL COMMENT '软删除时间',
       `api_usage_id` INT NOT NULL DEFAULT 0 COMMENT 'API使用记录ID',
       `sample_time` DATETIME NOT NULL COMMENT '抽样时间',
       `api_call_time` DATETIME NOT NULL COMMENT 'api调用时间',
       `api` VARCHAR(255) NOT NULL COMMENT 'API接口',
       `review_uin` INT NOT NULL DEFAULT 0 COMMENT '审查人id',
       `review_time` DATETIME COMMENT '审查时间',
       `review_status` VARCHAR(64) NOT NULL default 'undone' COMMENT '审查状态',
       `desc` TEXT NOT NULL COMMENT '评语',
       `review_result` VARCHAR(64) NOT NULL COMMENT '审查结果',
       `score` INT NOT NULL DEFAULT 0 COMMENT '评分',
       `tag` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标签',

       PRIMARY KEY (`id`),
       KEY `idx_api_usage_id` (`api_usage_id`),
       KEY `idx_api` (`api`),
       KEY `idx_review_uin` (`review_uin`),
       KEY `idx_review_status` (`review_status`),
       KEY `idx_review_result` (`review_result`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='API抽样数据表';

CREATE TABLE `admin_api_sample_log` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
      `created_at` DATETIME(3) NULL DEFAULT NULL COMMENT '创建时间',
      `updated_at` DATETIME(3) NULL DEFAULT NULL COMMENT '更新时间',
      `deleted_at` DATETIME(3) NULL DEFAULT NULL COMMENT '软删除时间',
      `start_time` DATETIME NOT NULL COMMENT '抽样范围-开始时间',
      `end_time` DATETIME NOT NULL COMMENT '抽样范围-结束时间',
      `req_body` TEXT COMMENT '筛选',
      `resp_body` TEXT COMMENT '结果',
      `sample_num` INT NOT NULL DEFAULT 0 COMMENT '抽样数量',

      PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='API 抽样日志表';