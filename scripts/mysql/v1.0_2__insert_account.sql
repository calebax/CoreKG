INSERT INTO account_api_key ( id, created_at, updated_at, deleted_at, uin, company_id, name, api_key, purpose, status, expired_at)
VALUES
	( 1, '2025-04-24 10:48:50.029', '2025-04-24 10:48:50.029', NULL, 1, 1, 'prod', @yg_AGENT_MODEL_APIKEY, 'prod', 'normal', NULL);

INSERT INTO account_privilege ( id, created_at, updated_at, deleted_at, description, api, action, action_path, parent_id, type, company_id )
VALUES
	( 1, '2025-06-13 16:50:32.000', '2025-06-13 16:50:34.000', NULL, 'agent', 'chat.Agent/chat/completions', '', '', 0, '5', 0 );

INSERT INTO account_api_key_privilege (id, created_at, updated_at, deleted_at, api_key_id, api_id) VALUES (1, '2025-04-24 11:48:57.000', '2025-04-24 11:48:57.000', NULL, 1, 1);
  
INSERT INTO admin_login_setting (
  created_at,
  updated_at,
  deleted_at,
  domain_name,
  path,
  env,
  is_enable_we_chat,
  is_enable_we_chat_com,
  is_enable_phone,
  is_enable_email,
  title,
  image_url,
  app_id,
  app_id_com,
  agent_id,
  allow_register,
  issuer,
  auth_key,
  is_enable_password,
  login_url
) VALUES (
  NOW(),              
  NOW(),               
  NULL,               
  CONCAT(SUBSTRING_INDEX(@yg_BASE_HOST, '://', -1), ':30000'),      
  NULL,            
  'prod',         
  0,                    
  0,                   
  0,                  
  0,                   
  '言古AI',        
  'https://prod-roc-1251908240.cos.ap-beijing.myqcloud.com//em-image/20250612/128-I2wKIJMbE.png',  
  '1111',        
  '1111',    
  '1111111111',    
  1,                    
  'yygu',       
  'wechat_web_oauth',    
  1,                    
  CONCAT(SUBSTRING_INDEX(@yg_BASE_HOST, '://', -1), ':30000')       
);

INSERT INTO user (id, created_at, updated_at, deleted_at, identify, name, bio, avatar_url, email, phone, password, github_id, work_wechat_user_id, wechat_union_id, wechat_web_open_id) VALUES (1, '2024-08-02 14:06:58.225', '2025-07-11 15:20:41.480', NULL, 'admin', 'admin', '', 'https://thirdwx.qlogo.cn/mmopen/vi_32/oAdEUk0YiaJaibAFvFwewct8YggbYiajeahTjliaP5AQb2YwxYs1kW6Cmib9oVonfGtibb2zQADibq0HiaWYgHmMAPFAgmliarWwfM0vcMKRZQ7FfOHg/132', 'admin@admin.com', '13800000000', @yg_ADMIN_PASSWORD_HASH, NULL, NULL, 'zzzzzzz', 'zzzzz');



INSERT INTO user_identification 
(id, created_at, updated_at, deleted_at, user_id, subject_type, subject_id, uin_status, issuer, name) 
VALUES 
(1, '2024-08-02 14:06:58.225', '2024-08-02 14:06:58.225', NULL, 1, 'company', 1, 'normal', 'yygu', 'admin');

INSERT INTO company (id, created_at, updated_at, deleted_at, name, alias, description, logo, address, tel, email, website, company_status, version, quota) VALUES (1, '2025-01-07 16:53:47.000', '2025-01-07 16:53:49.000', NULL, '默认组织', 'default-organization', '系统默认组织', '', '', '', '', '', 'passed', 'free_trail', '{\"qa_quota\": 1000, \"disk_quota\": 10737418240, \"agent_quota\": 5, \"employee_quota\": 5, \"graph_quota\": 5, \"article_quota\": 5}');


INSERT INTO account_employee (id, created_at, updated_at, deleted_at, company_id, user_id, sys_role, uin) VALUES (1, '2025-01-07 16:55:52.000', '2025-03-11 11:40:22.082', NULL, 1, 1, 'sys_admin', 1);

INSERT INTO account_department (id, created_at, updated_at, deleted_at, name, parent_id, sort, company_id) VALUES (1, '2025-01-07 16:55:52.000', '2025-01-07 16:55:52.000', NULL, '默认组织', 0, 1000, 1);

INSERT INTO account_rel_employee_department (id, created_at, updated_at, deleted_at, uin, department_id, is_primary, employee_id, company_id) VALUES (1, '2025-01-07 16:55:52.000', '2025-01-07 16:55:52.000', NULL, 1, 1, 1, 1, 1);