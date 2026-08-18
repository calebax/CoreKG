SET NAMES utf8mb4;

INSERT INTO `ke_message_template` (
    `name`, 
    `description`, 
    `type`, 
    `title_template`, 
    `content_template`, 
    `module`, 
    `route_path`, 
    `status`, 
    `created_at`, 
    `updated_at`, 
    `deleted_at`
) VALUES 
(
    'announcement_new_release', 
    '系统发布新功能公告模板', 
    'announcement', 
    '发版公告', 
    '{{.tag}}版本已更新，请查看更新内容', 
    'system', 
    '/announcement/{{.announcement_id}}', 
    'enable', 
    '2025-12-04 12:16:08.416', 
    '2025-12-05 11:02:07.504', 
    NULL
);