SET NAMES utf8mb4;

ALTER TABLE user ADD COLUMN company_quota BIGINT DEFAULT 2  COMMENT 'company_quota';

UPDATE company c
JOIN (
    -- 步骤2：通过最小ID找到对应的 user_id 和 subject_id
    SELECT u.user_id, u.subject_id
    FROM user_identification u
    JOIN (
        -- 步骤1：按 subject_id 分组，找出每个 company 对应的最小 uin_id (第一条)
        SELECT MIN(id) as min_id
        FROM user_identification
        WHERE subject_type = 'company' AND deleted_at IS NULL
        GROUP BY subject_id
    ) AS first_u ON u.id = first_u.min_id
    WHERE u.deleted_at IS NULL
) AS valid_data ON c.id = valid_data.subject_id
SET c.user_id = valid_data.user_id
WHERE c.deleted_at IS NULL;