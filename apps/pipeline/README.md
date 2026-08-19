# CoreKG Pipeline

CoreKG 知识引擎文档知识摄入管线，负责将上传的各类文档和知识文件自动解析、切块、向量化并存入知识库。系统包含两个核心微服务工作模块：

- **Document Analyser（`corekg_analyser`）**：文档解析服务，从 S3 下载文档，通过 MinerU API 将 PDF、图片等格式转换为结构化 Markdown，并将结果上传回 S3。同时支持 TXT、CSV、JSON、日志文件及视频关键帧提取等非 PDF 格式的处理。
- **Chunk Splitter（`corekg_chunk`）**：知识切块与增强服务，拉取 Markdown 内容后使用多种策略（智能 AST 切分、基础贪婪切分、高级动态切分、标题切分、正则切分）进行语义分块，并通过大模型对图片生成描述、对表格生成摘要，生成文本向量嵌入，最终写入 Elasticsearch。

两个服务均以任务队列 Worker 模式运行（通过 HTTP API 拉取/回调任务），支持 Docker 容器化部署和 Nuitka 代码加密私有化交付。

---

## 架构

```
                    Task Queue API (GetPendingTask / TaskCallBack)
                                |
        +-----------------------+-----------------------+
        |                                               |
  doc_worker_main.py                          chunk_worker_main.py
  (corekg_analyser)                           (corekg_chunk)
        |                                               |
  S3 下载 -> MinerU 解析 ->                    下载 MD -> 文本清洗 ->
  JSON 转 Markdown -> S3 上传                  语义切块 -> 图片增强 ->
                                              表格增强 -> Embedding -> ES 写入
```

---

## 项目结构

```
corekg-pipeline/
├── config/
│   ├── chunk_config.yaml          # Chunk 服务配置（LLM/ES/并发）
│   └── analyser_config.yaml       # 文档解析服务配置（API/S3）
│
├── tools/                         # 共享工具模块
│   ├── es_storage.py              # Elasticsearch 连接管理与 CRUD
│   ├── llm_chat.py                # LLM / Embedding 调用封装
│   ├── s3_common.py               # S3 客户端工厂
│   ├── s3_minio.py                # MinIO S3 操作
│   ├── s3_weedfs.py               # SeaweedFS S3 操作
│   └── temp_file.py               # 临时文件管理
│
├── corekg_chunk/                  # 知识切块与增强包
│   ├── chunk.py                   # 主处理管线入口
│   ├── utils.py                   # 工具函数（图片Base64、文本清洗、tiktoken）
│   ├── pipeline/
│   │   ├── knowchunk.py           # 多种切块策略实现（7种模式）
│   │   ├── image_vision_enhancer.py      # 视觉大模型图片描述增强
│   │   ├── image_context_extractor.py    # 图片上下文提取
│   │   └── table_enhancer.py            # 大模型表格摘要生成
│   ├── prompt/
│   │   └── prompt.py              # LLM Prompt 模板
│   ├── task/
│   │   └── work_chunk.py          # 任务消费者（拉取/处理/回调）
│   └── tiktoken_cache/
│       └── cl100k_base.tiktoken   # BPE 分词编码缓存
│
├── corekg_analyser/               # 文档解析包
│   ├── task/
│   │   └── work_pdf.py            # 任务消费者入口
│   ├── analyser_process.py        # MinerU API 调用与 ZIP 解压
│   ├── json_to_md.py              # JSON content_list 转 Markdown
│   ├── bbox_render.py             # bbox 坐标可视化渲染
│   ├── gpuinfo.py                 # GPU 监控与心跳
│   ├── video/
│   │   └── video_process.py       # 视频关键帧提取
│   └── others/
│       └── content_process.py     # 非 PDF 格式处理
│
├── private/docker/                # Docker 构建文件
│   ├── chunker/
│   │   ├── Dockerfile             # Chunker 加密镜像
│   │   └── Dockerfile.private     # Chunker 私有化镜像
│   └── analyser/
│       ├── Dockerfile             # Analyser 加密镜像
│       └── Dockerfile.private     # Analyser 私有化镜像
│
├── chunk_worker_main.py           # Chunk 服务启动入口
├── doc_worker_main.py             # 文档解析服务启动入口
├── Makefile                       # Docker 镜像构建
├── requirements.txt               # Python 主依赖
├── requirements_analyser.txt      # 文档解析依赖
└── .gitignore
```

---

## 依赖环境

- **Python**：3.10+
- **Elasticsearch**：8.x
- **对象存储**：S3 兼容（MinIO / SeaweedFS）
- **PDF 解析**：MinerU REST API
- **大模型**：OpenAI 兼容 API（DeepSeek-V3 / Qwen3-VL / Qwen3-Embedding）

---

## 安装

```bash
pip install -r requirements.txt
pip install -r requirements_analyser.txt
```

---

## 配置

### Chunk 服务配置（`config/chunk_config.yaml`）

| 配置项 | 说明 |
|---|---|
| `LLM` | 文本大模型配置（API Key / Model / Base URL） |
| `VLLM` | 多模态视觉模型配置 |
| `Embedding` | 向量嵌入模型配置 |
| `ES` | Elasticsearch 连接信息（Host / Account / Password / Index） |
| `Concurrency` | 模型并发控制参数（EB_WORKS / LLM_WORKS / LLM_TIMEOUT） |
| `General` | 基础配置（VERSION） |
| `Work` | 任务队列 API 地址与任务类型 |

### Analyser 服务配置（`config/analyser_config.yaml`）

| 配置项 | 说明 |
|---|---|
| `api_url` | 任务队列 API 地址 |
| `analyser_api_url` | MinerU PDF 解析 API 地址 |
| `s3` | S3 对象存储凭证 |

---

## 运行

### 文档解析 Worker

```bash
python doc_worker_main.py
```

通过环境变量可指定配置文件路径：
```bash
COREKG_CONFIGPATH=./config/analyser_config.yaml python doc_worker_main.py
```

### 知识切块 Worker

```bash
python chunk_worker_main.py
```

---

## Docker 构建

### 构建并推送镜像

```bash
# 私有化加密镜像
APP=chunker USE=private make push-image
APP=analyser USE=private make push-image
```

> 也可从仓库根目录统一入口调用（转发到本目录的 Makefile）：
> `make pipeline-push-image MODULE=graphrag APP=chunker USE=private`，等价于上面第一条命令。

### 构建流程说明

通过 `make` 构建时会：
1. 基于 `Dockerfile` 编译加密镜像（Nuitka 编译 Python 为 .so）
2. 从加密容器中提取编译产物
3. 将编译产物打入 `Dockerfile.private` 私有化部署镜像
4. 推送最终镜像至私有镜像仓库

---

## 切块策略

| 策略 | 说明 |
|---|---|
| `smart` | 基于 AST 的智能切分，在 Markdown 标题边界处切分 |
| `basic` | 基于 token 数量的行级贪婪合并切分 |
| `advanced` | 多轮动态大小优化切分（50-800 tokens） |
| `title` | 按标题级别（H1-H6）严格切分 |
| `strict_regex` | 基于自定义正则表达式的切分 |
| `slide` | 按 yg_pos 页码切分，一页一个 chunk，适用于 PPT/PDF 等按页解析的文档 |
| `resume` | 全文不切分，整个文档作为一个完整 chunk，适用于简历等短文档 |
| `paper` | 论文智能切分：摘要独立成块，按最频繁标题级别切分，相邻同级别合并 |
| `laws` | 法文切分：识别编/章/节/条层级，条不可拆分，按节或章聚合 |
| `llm` | 大模型语义切分：预分段为句子，LLM 判断合并边界后生成 chunks |
| `auto` | 自动选择：截取文档片段，LLM 分析后自动选择上述最合适的策略 |

---

## Chunk Worker 入参说明

Worker 通过 `knowledge.GetPendingTask` 拉取任务，`payload` 为 JSON 字符串，解析后结构如下：

### 顶层字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `file_url` | string | 是 | Markdown 文档的 S3 下载地址 |
| `file_name` | string | 否 | 源文件名，用于日志与 ES 入库标识 |
| `file_ext` | string | 否 | 源文件扩展名，影响 chunk type 判定（如 `.mp4` → `video`） |
| `forest_id` | string | 是 | 知识森林 ID |
| `company_id` | string | 是 | 公司/租户 ID |
| `uin` | string | 是 | 用户 ID |
| `file_id` | string | 是 | 文件唯一 ID |
| `es_index` | string | 否 | 目标 ES 索引名，不传则使用 config 默认值 |
| `llm` | object | 否 | 全局 LLM 配置，作为 `split_config` 中 LLM 参数的回退值 |
| `llm.model_name` | string | 否 | 文本模型名称 |
| `llm.api_key` | string | 否 | LLM API Key |
| `llm.base_url` | string | 否 | LLM API 地址 |
| `split_config` | object | 是 | 切块详细配置，见下方 |
| `task_type` | string | 是 | 任务类型标识 |

### `split_config` — 切块配置

#### 预处理规则 `preprocessing_rules`

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `remove_email` | bool | `true` | 是否清洗文本中的邮箱地址 |
| `remove_url` | bool | `true` | 是否清洗文本中未被 `![]()` 引用的裸 URL |
| `remove_empty_line` | bool | `true` | 是否合并连续空格与多余空白行 |

#### 分块策略参数

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `split_mode` | string | `"smart"` | 分块策略：`smart` / `basic` / `advanced` / `title` / `strict_regex` / `slide` / `resume` / `paper` / `laws` / `llm` / `auto` |
| `chunk_token_num` | int | `1024` | 目标 chunk 的 token 数上限，各策略均以此为粒度参考 |
| `chunk_size` | int | — | `chunk_token_num` 的旧字段名，兼容（优先级低于 `chunk_token_num`） |
| `min_chunk_tokens` | int | `10` | 最小 chunk token 数，低于此值的 chunk 会合并或注入上下文 |
| `split_level` | int | `2` | **仅 `title` 策略**：按第几级标题（1-6）切分，若仅产生 1 块则自动降级 |
| `overlap_ratio` | float | `0.0` | **所有策略**：相邻 chunk 尾部内容的重复比例（0.0 ~ 1.0），如 0.1 表示将前一个 chunk 倒数 10% 的内容追加到下一个 chunk 开头，增强语义连续性 |
| `split_overlap` | float | — | `overlap_ratio` 的旧字段名，兼容（优先级低于 `overlap_ratio`） |
| `regex_pattern` | string | `null` | **仅 `strict_regex` 策略**：自定义正则表达式，匹配行首时在此断开前一个 chunk（如 `"第[一二三四五六七八九十]+条"`） |
| `split_mark` | string | — | `regex_pattern` 的旧字段名，兼容（优先级低于 `regex_pattern`） |
| `delimiter` | string | `"\n!?。；！？"` | **仅 `basic` 策略**：分隔符控制（当前实现中保留兼容） |
| `enable_heading_in_content` | bool | `false` | **`smart` / `title` 策略**：是否在 chunk 内容头部自动补全缺失的父级标题路径，使孤立 chunk 保留上下文语义 |

#### 表格增强 LLM 配置

调用文本大模型为 `type: "table"` 的 chunk 生成表格摘要，插入 `<table>` 标签之前。

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `llm_enabled` | bool | config 文件 `LLM.STATUS` | 是否开启表格摘要生成。优先级高于 config 文件 |
| `llm_model` | string | config 文件 `LLM.MODEL` | 文本模型名称，如 `"deepseek-v3"`。未传时回退到 `llm.model_name` |
| `llm_api_key` | string | config 文件 `LLM.API_KEY` | 文本模型 API Key。未传时回退到 `llm.api_key` |
| `llm_base_url` | string | config 文件 `LLM.BASE_URL` | 文本模型 API 地址。未传时回退到 `llm.base_url` |
| `llm_timeout` | int | config 文件 `Concurrency.LLM_TIMEOUT` | 单次 LLM 调用超时（秒） |

> 优先级：`split_config` 直接字段 > `payload.llm` > `config/chunk_config.yaml`

#### 图片视觉增强 VLLM 配置

调用视觉多模态大模型为 chunk 中的图片生成 AI 描述，插入到图片标签之后。

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `vllm_enabled` | bool | config 文件 `VLLM.STATUS` | 是否开启图片视觉增强。关闭时仅提取 `image_url` 不做 AI 描述 |
| `vllm_model` | string | config 文件 `VLLM.MODEL` | 视觉模型名称，如 `"qwen3-vl-plus"` |
| `vllm_api_key` | string | config 文件 `VLLM.API_KEY` | 视觉模型 API Key |
| `vllm_base_url` | string | config 文件 `VLLM.BASE_URL` | 视觉模型 API 地址 |
| `image_width` | int | `1024` | 发送给视觉模型的图片缩放宽度（px），LLM 只识别该分辨率 |
| `image_height` | int | `null` | 发送给视觉模型的图片缩放高度（px），为 null 时按宽度等比缩放 |

#### 向量嵌入 Embedding 配置

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `embedding_model` | string | config 文件 `Embedding.MODEL` | Embedding 模型名称，如 `"Qwen3-Embedding-0.6B"` |
| `embedding_api_key` | string | config 文件 `Embedding.API_KEY` | Embedding 模型 API Key |
| `embedding_base_url` | string | config 文件 `Embedding.BASE_URL` | Embedding 模型 API 地址 |

#### 并发控制

| 字段 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `eb_max_concurrency` | int | config 文件 `Concurrency.EB_WORKS` | Embedding 向量化最大并发数，控制同时进行的向量化请求数 |
| `llm_max_concurrency` | int | config 文件 `Concurrency.LLM_WORKS` | LLM/VLLM 调用最大并发数，同时控制图片增强和表格增强的并发上限 |

### 请求示例

```json
{
  "task_id": "task_001",
  "payload": {
    "file_url": "https://s3.example.com/bucket/path/doc.md",
    "file_name": "产品手册.md",
    "file_ext": ".md",
    "forest_id": "434",
    "company_id": "147",
    "uin": "user_001",
    "file_id": "2936",
    "es_index": "ke_0",
    "task_type": "ke.knowledge_task",
    "llm": {
      "model_name": "deepseek-v3",
      "api_key": "yg-xxx",
      "base_url": "https://yygu.cn/v3/llm.chat"
    },
    "split_config": {
      "preprocessing_rules": {
        "remove_email": true,
        "remove_url": true,
        "remove_empty_line": true
      },
      "split_mode": "smart",
      "chunk_token_num": 1024,
      "min_chunk_tokens": 10,
      "overlap_ratio": 0.1,
      "enable_heading_in_content": true,
      "llm_enabled": true,
      "vllm_enabled": true,
      "image_width": 512,
      "eb_max_concurrency": 30,
      "llm_max_concurrency": 32
    }
  }
}
```

---

## 处理管线

### 文档解析管线
1. 从任务队列拉取待处理文档（S3 URL）
2. 从 S3 下载文档到本地临时目录
3. 调用 MinerU API 进行 PDF/图片解析，返回 ZIP 流
4. 解压 ZIP，提取 `content_list.json`、图片、原始文件
5. bbox 坐标转换（0-1000 归一化 → 300 DPI 像素），写回 JSON
6. 将 `content_list.json` 转换为结构化 Markdown
7. 为 Markdown 中的本地图片路径添加 S3 公网前缀
8. 上传解析结果到 S3
9. 清理临时文件，回调任务队列

### 知识切块管线
1. 从任务队列拉取待处理的 Markdown 文档 URL
2. 下载并清洗文本（移除 URL/邮箱/空行等）
3. 按配置的策略进行语义切块
4. 对每个 chunk 中的图片调用视觉大模型生成描述
5. 对每个 chunk 中的表格调用大模型生成摘要
6. 调用 Embedding 模型生成文本向量
7. 清除该文件在 ES 中的历史 chunks
8. 批量写入新的 chunks 到 Elasticsearch
9. 回调任务队列报告结果

---

## License

Copyright (c) YYGU. All rights reserved.
