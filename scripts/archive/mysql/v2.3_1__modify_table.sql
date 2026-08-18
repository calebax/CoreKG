SET NAMES utf8mb4;

ALTER TABLE user_identification ADD COLUMN last_login_at datetime(3) NULL COMMENT '最后登录时间';