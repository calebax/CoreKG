SET NAMES utf8mb4;

-- 简体中文/繁体中文 prompt：将「言小古」替换为「智能助手」
UPDATE chat_agent_version 
SET prompt_template = REPLACE(prompt_template, '言小古', '智能助手')
WHERE prompt_template LIKE '%言小古%';

-- 英文 prompt：将 'Yan Xiaogu' 替换为 'Intelligent Assistant'
UPDATE chat_agent_version 
SET prompt_template = REPLACE(prompt_template, 'Yan Xiaogu', 'Intelligent Assistant')
WHERE prompt_template LIKE '%Yan Xiaogu%';