INSERT INTO `core_settings` (`created_at`, `updated_at`, `deleted_at`, `group`, `key`, `name`, `describe`, `value_type`, `value`, `default`)
VALUES
	(now(), now(), NULL, 'corekg', 'official_website_wechat_webhook_url', '官网留资企微机器人', '', 'text', 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=6de1b354-46e4-4c80-998a-85f6968d3d62', '');
