SET NAMES utf8mb4;

UPDATE ke_article SET avatar_url =CONCAT('default', FLOOR(1 + RAND() * 3)) WHERE deleted_at IS NULL AND avatar_url = '';

UPDATE ke_forest_graph SET avatar_url =CONCAT('default', FLOOR(1 + RAND() * 6)) WHERE deleted_at IS NULL AND avatar_url = '';

