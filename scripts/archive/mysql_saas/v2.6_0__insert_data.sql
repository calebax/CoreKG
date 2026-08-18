SET NAMES utf8mb4;

-- 插入套餐数据
INSERT INTO `ke_package` (`name`, `description`, `price`, `sale_price`, `agent_quota`, `qa_quota`, `disk_quota`, `employee_quota`, `edition`, `level`, `period_type`, `source_type`, `status`, `extra`) VALUES
('社区版', '核心功能免费试用', 0, 0, 5, 100, 10737418240, 5, 'free_trail', 1, 'lifetime', 'system', 'online', NULL),
('专业版', '解锁更强大的AI功能', 9900, 2990, 200, 1000, 536870912000, 50, 'professional', 2, 'month', 'manual', 'online', '{"additional_notes":["完整的RBAC","专属客户支持","API响应"]}');
