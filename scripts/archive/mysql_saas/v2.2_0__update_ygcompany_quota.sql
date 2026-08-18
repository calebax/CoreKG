SET NAMES utf8mb4;


UPDATE `company`
SET
    `quota` = JSON_SET(
            `quota`,
            '$.graph_quota', 10000,
            '$.article_quota', 10000
              ),
    `updated_at` = NOW()
WHERE
    `id` = 2;