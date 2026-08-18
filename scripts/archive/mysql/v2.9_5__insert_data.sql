SET NAMES utf8mb4;

INSERT INTO ke_message_template (name, description, type, title_template, content_template, module,
                                           route_path, status, created_at, updated_at, deleted_at)
VALUES ('already_upload_same_file', '已上传相同文件提醒', 'system', '文档重复提醒',
        '您在知识库系统上传{{.newFileName}}文档时，系统检测到该文档与知识库({{.forestName}})中已存在的{{.oldFileName}}重复，请核对文档信息后再进行操作',
        'system', '', 'enable', '2026-01-08 09:32:50.000', '2026-01-08 11:16:06.339', null);