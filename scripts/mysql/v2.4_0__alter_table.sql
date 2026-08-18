SET NAMES utf8mb4;

ALTER TABLE chat_sessions MODIFY COLUMN file_id_list LONGTEXT;

ALTER TABLE chat_sessions MODIFY COLUMN forest_id_list LONGTEXT;

ALTER TABLE chat_sessions MODIFY COLUMN excel_id_list LONGTEXT;

ALTER TABLE chat_sessions MODIFY COLUMN excel_sheet_id_list LONGTEXT;

ALTER TABLE chat_sessions MODIFY COLUMN db_list LONGTEXT;

ALTER TABLE chat_sessions MODIFY COLUMN db_table_list LONGTEXT;

ALTER TABLE chat_sessions MODIFY COLUMN external_token_id_list LONGTEXT;