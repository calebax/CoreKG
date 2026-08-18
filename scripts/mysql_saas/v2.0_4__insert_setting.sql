SET NAMES utf8mb4;


INSERT INTO `core_settings`
(created_at, updated_at, deleted_at, `group`, `key`, name, `describe`, value_type, value, `default`)
VALUES ('2025-10-15 11:58:08.000', '2025-10-15 11:58:08.000', NULL, 'core', 'cos-company-logo', '组织logo存储',
        '组织logo存储配置', 'yaml',
        '
purpose: company-logo
s3:
  end_point: https://example.com:58081
  access_key_id: CHANGE_ME_S3_ACCESS_KEY
  secret_access_key: CHANGE_ME_S3_SECRET_KEY
  region: cn-xian
  bucket: test-knownow
  use_path_style: true
',
        NULL);