SET NAMES utf8mb4;

UPDATE chat_model
SET model_group = CASE
    WHEN model_name LIKE '%deepseek%' THEN 'DeepSeek'
    WHEN model_name LIKE '%qwen%' THEN '通义千问'
    WHEN model_name LIKE '%doubao%' THEN '豆包'
    ELSE model_group  -- 保持原值不变
END;
