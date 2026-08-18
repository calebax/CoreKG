## 统一约定

- 统一请求包体：

```json
{
  "request": {}
}
```

- 统一返回包体：

```json
{
  "code": 0,
  "message": "",
  "response": {}
}
```

- 分页查询参数

```json
{
  "request": {
    "offset": 0,
    "limit": 10,
    "orderBy": ["created_at desc"]
  }
}
```

- 分页返回统一使用：

```json
{
  "total": 1,
  "offset": 0,
  "limit": 10
}
```

## API Key

##### 以下接口实现apiKey鉴权

- 认证方式：`Authorization: Bearer <api_key>`

- 当前建议纳入 API Key 鉴权的接口：
  - `keapi.ListForest`
  - `keapi.CreateForest`
  - `keapi.UpdateForest`
  - `keapi.DeleteForest`
  - `keapi.ListFile`
  - `keapi.UploadFile`
  - `keapi.PreviewFileByURL`
  - `keapi.CreateDir`
  - `keapi.RenamePath`
  - `keapi.DeletePath`
  - `keapi.CreateChat`
  - `keapi.BatchGetChatInfo`
  - `keapi.UpdateChatName`
  - `keapi.DeleteChat`
  - `keapi.CreateChatMessage`
  - `keapi.ListChatMessages`
  - `keapi.Search`
  - `keapi.chat/chat/completions`

## 知识库

- ##### 列出知识库列表
  - 接口名称：`keapi.ListForest`
  - 请求结构：

```json
{
  "request": {
    "offset": 0,
    "limit": 10,
    "orderBy": ["created_at desc"]
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "total": 1,
    "offset": 0,
    "limit": 10,
    "data": [
      {
        "forest_id": 1,
        "name": "测试知识库",
        "knowledge_status": "success",
        "avatar_url": "",
        "description": "知识库描述",
        "file_count": 12,
        "total_size": 34078720
      }
    ]
  }
}
```

- ##### 创建知识库
  - 接口名称：`keapi.CreateForest`
  - 请求结构：

```json
{
  "request": {
    "name": "测试知识库",
    "avatar_url": "",
    "description": "知识库描述",
    "forest_type": "file"
  }
}
```

- 支持的知识库类型：
  - 标准知识库：`forest_type=file`
  - Excel 知识库：`forest_type=data`
- 默认值：
  - 未传 `forest_type` 时默认为标准知识库
  - `public_scope` 固定为 `private`
  - `keapi` 会根据 `forest_type` 自动映射数据源类型

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "forest_id": 1
  }
}
```

- ##### 更新知识库
  - 接口名称：`keapi.UpdateForest`
  - 请求结构：

```json
{
  "request": {
    "forest_id": 1,
    "name": "测试知识库-已更新",
    "description": "更新后的知识库描述"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": ""
}
```

- ##### 删除知识库
  - 接口名称：`keapi.DeleteForest`
  - 请求结构：

```json
{
  "request": {
    "forest_id": 1
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {}
}
```

## 文档

- ##### 列出文档列表
  - 接口名称：`keapi.ListFile`
  - 请求结构：

```json
{
  "request": {
    "forest_id": 1,
    "offset": 0,
    "limit": 10,
    "orderBy": ["created_at desc"]
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "total": 2,
    "offset": 0,
    "limit": 10,
    "data": [
      {
        "forest_file_id": 11,
        "forest_id": 1,
        "is_dir": -1,
        "parent_id": 0,
        "name": "测试文档.pdf",
        "size": 102400,
        "ext": ".pdf",
        "file_status": "success"
      }
    ]
  }
}
```

- ##### 文档上传
  - 接口名称：`keapi.UploadFile`
  - 请求结构：`multipart/form-data`

| 字段名      | 类型   | 必填 | 说明                      |
| ----------- | ------ | ---- | ------------------------- |
| `file`      | file   | 是   | 上传文件                  |
| `forest_id` | string | 是   | 知识库 ID                 |
| `parent_id` | string | 否   | 父目录 ID，不传默认根目录 |

- 返回结构：

```json
{
  "code": 0,
  "message": "kecore_upload_success",
  "response": {
    "forest_file_id": 11
  }
}
```

- ##### 下载文档
  - 接口名称：`keapi.PreviewFileByURL`
  - 请求结构：

```json
{
  "request": {
    "forest_file_id": 11
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "url": "https://xxx/xxx.pdf?signature=..."
  }
}
```

- ##### 创建文件夹
  - 接口名称：`keapi.CreateDir`
  - 请求结构：

```json
{
  "request": {
    "forest_id": 1,
    "parent_id": 0,
    "name": "一级目录"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "forest_file_id": 21
  }
}
```

- ##### 更新知识库节点信息
  - 接口名称：`keapi.RenamePath`
  - 请求结构：

```json
{
  "request": {
    "forest_file_id": 21,
    "name": "一级目录-已更新"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {}
}
```

- ##### 删除节点
  - 接口名称：`keapi.DeletePath`
  - 请求结构：

```json
{
  "request": {
    "forest_file_id": [21]
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {}
}
```

## 检索

- ##### 内容检索
  - 接口名称：`keapi.Search`
  - 请求结构：

```json
{
  "request": {
    "forest_ids": [1],
    "query": "查询关键词"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "doc_search_result": [
      {
        "forest_file_id": 11,
        "forest_id": 1,
        "file_name": "测试文档.pdf",
        "created_at": "2026-03-30T10:00:00+08:00",
        "_score": 12.34,
        "highlights": [
          {
            "_score": 12.34,
            "description": "原始分片内容",
            "highlighted_description": "包含 <em>查询关键词</em> 的分片内容",
            "image_url": "",
            "location": [1, 0, 0, 0, 0]
          }
        ]
      }
    ],
    "image_search_result": [],
    "video_search_result": []
  }
}
```

## 对话

- ##### 创建对话会话
  - 接口名称：`keapi.CreateChat`
  - 请求参数：

| 参数              | 类型   | 必填 | 默认值     | 说明               |
| ----------------- | ------ | ---- | ---------- | ------------------ |
| `forest_file_ids` | Array  | 否   | -          | 知识库文档 ID 列表 |
| `name`            | String | 否   | `New Chat` | 会话名称           |

- 请求结构：

```json
{
  "request": {
    "forest_file_ids": [21],
    "name": "可选会话名"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "session_id": 123,
    "name": "New Chat",
    "forest_file_id": [21],
    "forest_id": [1],
    "model_name": "xxx",
    "created_at": "2026-04-18T10:00:00+08:00"
  }
}
```

- ##### 批量查询对话会话
  - 接口名称：`keapi.BatchGetChatInfo`
  - 请求结构：

```json
{
  "request": {
    "session_ids": [123, 124]
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "total": 1,
    "offset": 0,
    "limit": 2,
    "data": [
      {
        "session_id": 123,
        "name": "New Chat",
        "forest_file_id": [21],
        "forest_id": [1],
        "model_name": "xxx",
        "created_at": "2026-04-18T10:00:00+08:00"
      }
    ]
  }
}
```

- ##### 更新对话会话名称
  - 接口名称：`keapi.UpdateChatName`
  - 请求结构：

```json
{
  "request": {
    "session_id": 123,
    "name": "新的会话名称"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "session_id": 123,
    "name": "新的会话名称",
    "forest_file_id": [21],
    "forest_id": [1],
    "model_name": "xxx",
    "created_at": "2026-04-18T10:00:00+08:00"
  }
}
```

- ##### 删除对话会话
  - 接口名称：`keapi.DeleteChat`
  - 说明：删除会话记录；与 `chat.RemoveChatSession` 保持一致，暂不删除会话下的问答记录。
  - 请求结构：

```json
{
  "request": {
    "session_id": 123
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "session_id": 123,
    "deleted": true
  }
}
```

- ##### 创建用户消息
  - 接口名称：`keapi.CreateChatMessage`
  - 说明：向指定会话追加一条用户消息；不触发模型回复，不更新会话名称。
  - 请求结构：

```json
{
  "request": {
    "session_id": 123,
    "content": "请总结这篇文档"
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "message_id": "8f3a1c2d9b4e6f0012a7c8d3e5f6a9b0",
    "session_id": 123,
    "role": "user",
    "content": "请总结这篇文档",
    "created_at": "2026-04-20T10:00:00+08:00"
  }
}
```

- ##### 查询对话会话消息列表
  - 接口名称：`keapi.ListChatMessages`
  - 请求结构：

```json
{
  "request": {
    "session_id": 123
  }
}
```

- 返回结构：

```json
{
  "code": 0,
  "message": "",
  "response": {
    "data": [
      {
        "message_id": "8f3a1c2d9b4e6f0012a7c8d3e5f6a9b0",
        "session_id": 123,
        "role": "user",
        "content": "请总结这篇文档",
        "created_at": "2026-04-18T10:00:00+08:00"
      },
      {
        "message_id": "b2e4a6c8d0f1357911ace02468bd3579",
        "session_id": 123,
        "role": "assistant",
        "content": "这篇文档主要介绍了……",
        "created_at": "2026-04-18T10:00:01+08:00"
      }
    ],
    "has_more": false
  }
}
```

- ##### 对话
  - 接口名称：`keapi.chat/chat/completions`
  - 请求参数：

| 参数               | 类型    | 必填 | 默认值  | 说明                                          |
| ------------------ | ------- | ---- | ------- | --------------------------------------------- |
| `session_id`       | Number  | 否   | -       | 对话会话 ID，传入时复用该会话及其历史消息     |
| `forest_file_id`   | Array   | 否   | -       | 知识库文档 ID 列表                            |
| `messages`         | Array   | 是   | -       | Conversation messages，格式与 OpenAI 协议一致 |
| `stream`           | Boolean | 否   | `false` | Enable streaming                              |
| `temperature`      | Number  | 否   | `0.2`   | OpenAI 协议采样温度                           |
| `top_p`            | Number  | 否   | -       | OpenAI 协议 nucleus sampling 参数             |
| `presence_penalty` | Number  | 否   | -       | OpenAI 协议话题重复惩罚参数                   |

- 说明：
  - 传入 `session_id` 时，服务端使用该会话已有消息作为上下文，本次 `forest_file_id` 会被忽略。
  - 不传 `session_id` 时，服务端创建临时会话；`forest_file_id` 可省略。
  - 未提供 `forest_file_id` 的会话不会进行知识库范围检索。

- 请求结构：

```json
{
  "request": {
    "forest_file_id": [21],
    "messages": [
      {
        "role": "user",
        "content": "请总结这篇文档的主要内容"
      }
    ],
    "stream": false,
    "temperature": 0.2,
    "top_p": 1,
    "presence_penalty": 0
  }
}
```

- 返回说明：

当 `stream=false` 时，返回 JSON

```json
{
  "id": "b2e4a6c8d0f1357911ace02468bd3579",
  "object": "chat.completion",
  "created": 1774915200,
  "model": "forest-chat",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "这篇文档主要介绍了……"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 123,
    "completion_tokens": 56,
    "total_tokens": 179
  }
}
```

当 `stream=true` 时，返回 `text/event-stream`，采用标准 SSE 格式

```text
data: {"id":"b2e4a6c8d0f1357911ace02468bd3579","object":"chat.completion.chunk","created":1774915200,"model":"forest-chat","choices":[{"index":0,"delta":{"role":"assistant","content":"这"},"finish_reason":null}]}

data: {"id":"b2e4a6c8d0f1357911ace02468bd3579","object":"chat.completion.chunk","created":1774915200,"model":"forest-chat","choices":[{"index":0,"delta":{"content":"篇文档"},"finish_reason":null}]}

data: {"id":"b2e4a6c8d0f1357911ace02468bd3579","object":"chat.completion.chunk","created":1774915200,"model":"forest-chat","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

## 参考

阿里百炼

https://help.aliyun.com/zh/model-studio/rag-knowledge-base-api-guide#1b153d3113d77

LightRAG

https://github.com/HKUDS/LightRAG/blob/main/lightrag/api/README.md

ragflow

https://ragflow.io/docs/http_api_reference#download-file

乐享知识库

https://lexiang.tencent.com/wiki/api/15007.html

pageindex

https://docs.pageindex.ai/endpoints
