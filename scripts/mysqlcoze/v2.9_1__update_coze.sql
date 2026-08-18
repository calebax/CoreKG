SET NAMES utf8mb4;
START TRANSACTION;

-- Ensure kv flag to avoid using old model key exists
INSERT INTO kv_entries (namespace, key_data, value_data)
SELECT 'kv_model_ns', 'do_not_use_old_model_key', '{}'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM kv_entries WHERE namespace = 'kv_model_ns' AND key_data = 'do_not_use_old_model_key'
);

COMMIT;
