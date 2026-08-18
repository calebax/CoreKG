# API 接口参考

## 1. 路由结构总览

路由注册位于 `api/router/coze/api.go`，使用 Hertz 框架。分为以下路由组：

- `/api/workflow_api/` — 工作流核心 API（WebAPI，Session 认证）
- `/api/admin/` — 管理后台 API（Admin UIN 白名单）
- `/api/bot/` — Bot 相关
- `/api/common/upload/` — 文件上传
- `/api/conversation/` — 会话 API
- `/api/draftbot/` — Agent CRUD
- `/api/intelligence_api/` — 项目管理
- `/api/knowledge/` — 知识库 API
- `/api/memory/` — 记忆与变量
- `/api/plugin_api/` — 插件 API
- `/api/playground_api/` — Playground API
- `/open_api/` — OpenAPI（Bearer Token 认证）
- `/v1/` — 公开 OpenAPI v1
- `/v3/` — 公开 OpenAPI v3
- 静态文件 — SPA 前端

## 2. 工作流核心 API（/api/workflow_api/）

### 2.1 CRUD

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/workflow_api/create` | CreateWorkflow | 创建工作流 |
| POST | `/api/workflow_api/save` | SaveWorkflow | 保存草稿 |
| POST | `/api/workflow_api/update_meta` | UpdateWorkflowMeta | 更新元数据 |
| POST | `/api/workflow_api/delete` | DeleteWorkflow | 删除工作流 |
| POST | `/api/workflow_api/batch_delete` | BatchDeleteWorkflow | 批量删除 |
| POST | `/api/workflow_api/copy` | CopyWorkflow | 复制工作流 |
| POST | `/api/workflow_api/copy_wk_template` | CopyWkTemplateApi | 从模板复制 |
| POST | `/api/workflow_api/publish` | PublishWorkflow | 发布版本 |

### 2.2 查询

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/workflow_api/canvas` | GetCanvasInfo | 获取画布信息 |
| POST | `/api/workflow_api/workflow_detail` | WorkflowDetail | 工作流详情 |
| POST | `/api/workflow_api/workflow_detail_info` | WorkflowDetailInfo | 工作流详细信息 |
| POST | `/api/workflow_api/workflow_list` | WorkflowList | 工作流列表 |
| POST | `/api/workflow_api/list_publish_workflow` | ListPublishWorkflow | 已发布版本列表 |
| GET | `/api/workflow_api/released_workflows` | ReleasedWorkflows | 已发布工作流 |
| GET | `/api/workflow_api/history_schema` | HistorySchema | 历史版本 Schema |
| POST | `/api/workflow_api/workflow_references` | WorkflowReferences | 引用关系 |
| GET | `/api/workflow_api/delete_strategy` | DeleteStrategy | 删除策略检查 |
| GET | `/api/workflow_api/example_workflow_list` | ExampleWorkflowList | 示例工作流列表 |
| GET | `/api/workflow_api/sign_image_url` | SignImageURL | 签名图片 URL |

### 2.3 执行与调试

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/workflow_api/test_run` | WorkFlowTestRun | 测试运行 |
| POST | `/api/workflow_api/test_resume` | WorkFlowTestResume | 测试恢复 |
| POST | `/api/workflow_api/cancel` | CancelWorkFlow | 取消执行 |
| POST | `/api/workflow_api/nodeDebug` | WorkflowNodeDebugV2 | 单节点调试 |
| GET | `/api/workflow_api/get_process` | GetWorkFlowProcess | 获取执行进度 |
| GET | `/api/workflow_api/get_trace` | GetTrace | 获取执行追踪 |
| GET | `/api/workflow_api/get_node_execute_history` | GetNodeExecuteHistory | 节点执行历史 |

### 2.4 节点与 Schema

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/workflow_api/node_type` | NodeType | 节点类型列表 |
| GET | `/api/workflow_api/node_template_list` | NodeTemplateList | 节点模板列表 |
| GET | `/api/workflow_api/node_panel_search` | NodePanelSearch | 节点面板搜索 |
| POST | `/api/workflow_api/validate_tree` | ValidateTree | 验证工作流树 |
| GET | `/api/workflow_api/apiDetail` | ApiDetail | API 详情 |
| GET | `/api/workflow_api/llm_fc_setting_detail` | LLMFcSettingDetail | LLM FC 设置详情 |
| GET | `/api/workflow_api/llm_fc_setting_merged` | LLMFcSettingMerged | LLM FC 合并设置 |

### 2.5 ChatFlow

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/workflow_api/chat_flow_role/create` | CreateChatFlowRole | 创建 ChatFlow 角色 |
| POST | `/api/workflow_api/chat_flow_role/delete` | DeleteChatFlowRole | 删除 ChatFlow 角色 |
| GET | `/api/workflow_api/chat_flow_role/get` | GetChatFlowRole | 获取 ChatFlow 角色 |

### 2.6 项目会话

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/api/workflow_api/project_conversation/create` | CreateProjectConversation | 创建项目会话 |
| POST | `/api/workflow_api/project_conversation/delete` | DeleteProjectConversation | 删除项目会话 |
| GET | `/api/workflow_api/project_conversation/list` | ListProjectConversation | 项目会话列表 |
| POST | `/api/workflow_api/project_conversation/update` | UpdateProjectConversation | 更新项目会话 |

### 2.7 其他

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/api/workflow_api/upload/auth_token` | UploadAuthToken | 上传认证 Token |

## 3. OpenAPI v1（/v1/）

### 3.1 工作流执行

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST | `/v1/workflow/run` | OpenAPIRunFlow | 同步执行 |
| POST | `/v1/workflow/ygrun` | OpenAPIYgRunFlow | CoreKG 扩展执行 |
| POST | `/v1/workflow/stream_run` | OpenAPIStreamRunFlow | 流式执行（SSE） |
| POST | `/v1/workflow/stream_resume` | OpenAPIStreamResumeFlow | 流式恢复（SSE） |
| GET | `/v1/workflow/get_run_history` | OpenAPIGetWorkflowRunHistory | 执行历史 |
| GET | `/v1/workflows/:workflow_id` | OpenAPIGetWorkflowInfo | 工作流信息 |
| POST | `/v1/workflows/chat` | OpenAPIChatFlowRun | ChatFlow 执行（SSE） |
| POST | `/v1/workflow/conversation/create` | OpenAPICreateConversation | 创建会话 |

### 3.2 会话

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/v1/conversations` | 会话列表 |
| DELETE | `/v1/conversations/:id` | 删除会话 |
| PUT | `/v1/conversations/:id` | 更新会话 |
| POST | `/v1/conversation/create` | 创建会话 |
| GET | `/v1/conversation/retrieve` | 获取会话 |
| GET | `/v1/conversation/message/list` | 消息列表 |
| POST | `/v1/conversations/:id/clear` | 清空会话 |

### 3.3 知识库

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/v1/datasets` | 数据集列表 |
| POST | `/v1/datasets` | 创建数据集 |
| DELETE | `/v1/datasets/:id` | 删除数据集 |
| PUT | `/v1/datasets/:id` | 更新数据集 |
| GET | `/v1/datasets/:id/images` | 数据集图片 |
| POST | `/v1/datasets/:id/process` | 处理数据集 |

### 3.4 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/v1/apps/:app_id` | 应用信息 |
| GET | `/v1/bot/get_online_info` | Bot 在线信息 |
| GET | `/v1/bots/:bot_id` | Bot 信息 |
| POST | `/v1/files/upload` | 文件上传 |

## 4. OpenAPI v3（/v3/）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v3/chat` | Chat（统一对话接口） |
| POST | `/v3/chat/cancel` | 取消 Chat |
| GET | `/v3/chat/retrieve` | 获取 Chat 结果 |
| GET | `/v3/chat/message/list` | Chat 消息列表 |

## 5. OpenAPI（/open_api/）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/open_api/knowledge/document/create` | 创建文档 |
| GET | `/open_api/knowledge/document/list` | 文档列表 |
| POST | `/open_api/knowledge/document/delete` | 删除文档 |
| POST | `/open_api/knowledge/document/update` | 更新文档 |

## 6. CoreKG 自定义路由

文件：`api/router/coze/custom_corekg.go`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/internal/space_sync` | 空间同步 Webhook |
| GET | `/api/internal/agent/external_info` | Agent 短链码 |
| POST | `/api/internal/agent/set_external_status` | 设置 Agent 外部状态 |
| GET | `/api/permission_api/pat/get_personal_access_token` | 获取/创建 API Key |
| GET | `/api/public/agent/external_token` | 公开 Agent Token |
| POST | `/api/playground_api/produce/create_bot` | 生产创建 Bot（含权限） |
| POST | `/api/playground_api/bot_config/create` | 创建 Bot 配置 |

## 7. 管理后台 API（/api/admin/）

所有路由需 AdminAuthMW（UIN 白名单）：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/config/basic/get` | 获取基础配置 |
| POST | `/api/admin/config/basic/save` | 保存基础配置 |
| GET | `/api/admin/config/knowledge/get` | 获取知识库配置 |
| POST | `/api/admin/config/knowledge/save` | 保存知识库配置 |
| POST | `/api/admin/config/model/create` | 创建模型配置 |
| POST | `/api/admin/config/model/delete` | 删除模型配置 |
| GET | `/api/admin/config/model/list` | 模型配置列表 |

## 8. Agent CRUD API（/api/draftbot/）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/draftbot/create` | 创建 Agent |
| POST | `/api/draftbot/delete` | 删除 Agent |
| POST | `/api/draftbot/duplicate` | 复制 Agent |
| POST | `/api/draftbot/get_display_info` | 获取展示信息 |
| GET | `/api/draftbot/list_draft_history` | 草稿历史列表 |
| POST | `/api/draftbot/publish` | 发布 Agent |
| POST | `/api/draftbot/update_display_info` | 更新展示信息 |
| POST | `/api/draftbot/commit_check` | 提交检查 |
| POST | `/api/draftbot/publish/connector/list` | 发布连接器列表 |

## 9. 知识库 API（/api/knowledge/）

### 9.1 知识库 CRUD

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/knowledge/create` | 创建知识库 |
| POST | `/api/knowledge/delete` | 删除知识库 |
| POST | `/api/knowledge/detail` | 知识库详情 |
| POST | `/api/knowledge/list` | 知识库列表 |
| POST | `/api/knowledge/update` | 更新知识库 |

### 9.2 文档管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/knowledge/document/create` | 创建文档 |
| POST | `/api/knowledge/document/delete` | 删除文档 |
| POST | `/api/knowledge/document/list` | 文档列表 |
| POST | `/api/knowledge/document/resegment` | 重新分段 |
| POST | `/api/knowledge/document/update` | 更新文档 |
| GET | `/api/knowledge/document/progress/get` | 处理进度 |

### 9.3 切片管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/knowledge/slice/create` | 创建切片 |
| POST | `/api/knowledge/slice/delete` | 删除切片 |
| POST | `/api/knowledge/slice/list` | 切片列表 |
| POST | `/api/knowledge/slice/update` | 更新切片 |

### 9.4 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/knowledge/icon/get` | 知识库图标 |
| POST | `/api/knowledge/photo/*` | 图片相关操作 |
| POST | `/api/knowledge/review/*` | Review 操作 |
| POST | `/api/knowledge/table_schema/*` | 表格 Schema 操作 |

## 10. 插件 API（/api/plugin_api/）

共 31 个端点，涵盖插件的完整生命周期：注册、创建/删除/更新 API、发布、调试、OAuth、资源复制等。

## 11. 内存与变量 API（/api/memory/）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/memory/doc_table_info` | 文档表信息 |
| GET | `/api/memory/sys_variable_conf` | 系统变量配置 |
| GET | `/api/memory/table_mode_config` | 表模式配置 |
| POST | `/api/memory/database/*` | 数据库 CRUD（14 个端点） |
| GET | `/api/memory/project/variable/meta_list` | 变量元信息列表 |
| POST | `/api/memory/project/variable/meta_update` | 更新变量元信息 |
| GET | `/api/memory/table_file/get_progress` | 表文件处理进度 |
| POST | `/api/memory/table_file/submit` | 提交表文件 |
| GET | `/api/memory/table_schema/get` | 获取表 Schema |
| POST | `/api/memory/table_schema/validate` | 验证表 Schema |
| POST | `/api/memory/variable/delete` | 删除变量 |
| GET | `/api/memory/variable/get` | 获取变量 |
| GET | `/api/memory/variable/get_meta` | 获取变量元信息 |
| POST | `/api/memory/variable/upsert` | 创建/更新变量 |

## 12. Playground API（/api/playground_api/）

Prompt 资源、快捷命令、文件/图片 URL、用户信息、空间列表、Bot 操作等。

## 13. 会话 API（/api/conversation/）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/conversation/break_message` | 中断消息 |
| POST | `/api/conversation/chat` | Agent 对话 |
| POST | `/api/conversation/clear_message` | 清空消息 |
| POST | `/api/conversation/create_section` | 创建分段 |
| POST | `/api/conversation/delete_message` | 删除消息 |
| GET | `/api/conversation/get_message_list` | 消息列表 |

## 14. SSE 流式响应格式

所有流式端点（stream_run、stream_resume、chat）使用 Server-Sent Events：

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: message
data: {"id":"evt_001","event":"workflow.running","data":{"execute_id":"123"}}

event: message
data: {"id":"evt_002","event":"node.start","data":{"node_id":"node_1","node_type":"LLM"}}

event: message
data: {"id":"evt_003","event":"node.streaming_output","data":{"node_id":"node_1","content":"Hello"}}

event: message
data: {"id":"evt_004","event":"workflow.done","data":{"output":{...},"token_info":{...}}}
```

## 15. 静态文件服务

| 路径 | 说明 |
|------|------|
| `/` | SPA index.html |
| `/static/*` | SPA 静态资源 |
| `/admin/` | 管理面板 |
| `/favicon.png` | Favicon |
| `/sign` | SPA（签名页面） |
| 其他 API 路径 | 404 JSON |
| 其他非 API 路径 | SPA index.html（支持前端路由） |
