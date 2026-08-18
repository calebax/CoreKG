SET NAMES utf8mb4;

START TRANSACTION;

CREATE TABLE `account_department`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`       varchar(100)    NOT NULL COMMENT '部门名称，全局唯一',
    `parent_id`  bigint unsigned NOT NULL DEFAULT 0 COMMENT '父部门ID，0表示顶层部门',
    `sort`       bigint unsigned NOT NULL DEFAULT 0 COMMENT '同级排序',
    `company_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '组织ID',
    `created_at` datetime(3)     NULL COMMENT '创建时间',
    `updated_at` datetime(3)     NULL COMMENT '更新时间',
    `deleted_at` datetime(3)     NULL COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='部门信息表';

CREATE TABLE `account_rel_employee_department`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `uin`           bigint unsigned NOT NULL COMMENT 'uin',
    `department_id` bigint unsigned NOT NULL COMMENT '部门ID',
    `is_primary`    tinyint(1)      NOT NULL DEFAULT -1 COMMENT '是否为主部门 (1:是, -1:否)',
    `employee_id`   bigint unsigned NOT NULL COMMENT 'employee_id',
    `company_id`    bigint unsigned NOT NULL COMMENT '组织ID',
    `created_at`    datetime(3)     NULL COMMENT '创建时间',
    `updated_at`    datetime(3)     NULL COMMENT '更新时间',
    `deleted_at`    datetime(3)     NULL COMMENT '删除时间（软删除）',
    PRIMARY KEY (`id`),
    INDEX `idx_department_id` (`department_id`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci COMMENT ='员工与部门关系表';

ALTER TABLE user
    ADD COLUMN password_changed tinyint NOT NULL DEFAULT -1 COMMENT '是否修改过密码(1:是 -1:否)';



COMMIT;