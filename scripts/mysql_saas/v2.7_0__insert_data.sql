SET NAMES utf8mb4;

INSERT INTO `ke_message_template` (`name`, `description`, `type`, `title_template`, `content_template`, `module`, `route_path`, `status`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	('order_about_to_expire', '订单即将过期提醒', 'system', '温馨提示', '您当前使用的升级版即将于{{.expire_date}}到期，为避免服务到期后对您的使用造成影响，请及时前往「个人信息」页面进行续费', 'system', '', 'enable', now(), now(), NULL);
