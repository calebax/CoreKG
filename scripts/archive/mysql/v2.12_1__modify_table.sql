SET NAMES utf8mb4;

ALTER TABLE `ke_company_quota`
ADD COLUMN `article_quota` int unsigned NOT NULL DEFAULT 0 COMMENT '文章数量' AFTER `employee_quota`;

ALTER TABLE `ke_package`
ADD COLUMN `article_quota` int unsigned NOT NULL DEFAULT 0 COMMENT '文章数量' AFTER `employee_quota`;


UPDATE ke_company_quota SET article_quota=5 WHERE source_type='manual' AND article_quota=0;
UPDATE ke_company_quota SET article_quota=20 WHERE source_type='order' AND article_quota=0;
UPDATE ke_package SET `article_quota`=5 WHERE edition='free_trail' AND `level`=1 AND article_quota=0;
UPDATE ke_package SET `article_quota`=20 WHERE edition='professional' AND `level`=2 AND article_quota=0;