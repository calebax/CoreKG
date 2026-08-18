SET NAMES utf8mb4;

UPDATE `ke_forest_excel_sheet`
SET 
  `parent_id` = `id`
WHERE 
  `parent_id` = 0;