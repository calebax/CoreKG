import re
import uuid
import hashlib
import asyncio
import yaml

from loguru import logger 

from corekg_chunk.pipeline.knowchunk import get_knowchunks, extract_yg_pos_mapping, find_yg_locations_for_chunks, compute_page_bboxes
from corekg_chunk.utils import get_content, simple_clean
from tools.llm_chat import chat_with_embedding
from corekg_chunk.pipeline.image_vision_enhancer import enhance_chunks_with_vision
from corekg_chunk.pipeline.table_enhancer import enhance_table_chunks

# ===================== 加载配置 =====================
with open("./config/chunk_config.yaml", "r", encoding="utf-8") as file:
    config = yaml.safe_load(file)

# 算法版本配置
version = config["General"]["VERSION"]

def knowchunk_format_conversion(chunks: list, chunk_metas: list, yg_locations: list, forest_id, company_id, uin, file_id, file_name, file_ext):
    """
    将 knowchunk 策略产出的分块转换为 corekg_chunk 的 ES 入库格式
    """
    chunks_dic = {}
    index = 1
    for i, chunk_text in enumerate(chunks):
        if not chunk_text or chunk_text in ['\n', '\n\n', '\t', '\t\t', ' ', '  ']:
            continue
        uid = str(uuid.uuid4())
        meta = chunk_metas[i] if i < len(chunk_metas) else {}

        title_level = None
        headers = meta.get('headers', {})
        if headers:
            sorted_levels = sorted(headers.keys())
            title_level = [" -> ".join(headers[l] for l in sorted_levels)]

        yg_loc = yg_locations[i] if i < len(yg_locations) else None

        chunks_dic[uid] = {
            "forest_id": forest_id,
            "company_id": company_id,
            "uin": uin,
            "file_id": file_id,
            "version": version,
            "file_name": file_name,

            "type": None,
            "tokens": len(chunk_text),
            "chunk_size": len(chunk_text),

            "sequence": index,
            "location": None,
            "yg_location": yg_loc,

            "description": chunk_text,
            "description_hash": hashlib.sha256(chunk_text.encode('utf-8')).hexdigest(),
            "embedding": None,

            "image_url": None,
            "image_embedding": None,

            "formula": None,

            "table": None,

            "title_level_ids": None,
            "title_level": title_level,
            "references": None,

            "graph_info": None,
            "graph_status": None,
        }
        if file_ext == ".mp4":
            chunks_dic[uid]['type'] = 'video'
        elif bool(re.search(r'<table>.*</table>', chunk_text)):
            chunks_dic[uid]['type'] = 'table'
        elif bool(re.search(r'(!\[.*?\]\(.*?\))', chunk_text)):
            chunks_dic[uid]['type'] = 'image'
        else:
            chunks_dic[uid]['type'] = 'chunk'
        index += 1

    return chunks_dic


async def chunk_emb(chunks_dic: dict, eb_max_concurrency: int = 100,
                    embedding_model: str = None, embedding_api_key: str = None,
                    embedding_base_url: str = None):
    """
    异步处理所有 chunks，限制最大并发数
    """
    _embedding_model = embedding_model or config['Embedding']['MODEL']
    _embedding_api_key = embedding_api_key or config['Embedding']['API_KEY']
    _embedding_base_url = embedding_base_url or config['Embedding']['BASE_URL']

    semaphore = asyncio.Semaphore(eb_max_concurrency)
    total = len(chunks_dic)
    completed = 0
    report_interval = max(1, total // 10)

    async def process_with_progress(chunk_id, chunk):
        nonlocal completed
        async with semaphore:
            description_embedding = await chat_with_embedding(
                text=chunk.get("description"),
                model=_embedding_model,
                api_key=_embedding_api_key,
                base_url=_embedding_base_url,
            )
        chunk["embedding"] = description_embedding
        completed += 1
        if completed % report_interval == 0 or completed == total:
            logger.info(f"  [5/5] 向量化进度: {completed}/{total} ({completed*100//total}%)")
        return chunk_id, chunk

    tasks = [
        process_with_progress(chunk_id, chunks_dic[chunk_id])
        for chunk_id in chunks_dic
    ]

    results = await asyncio.gather(*tasks)

    return {cid: c for cid, c in results}

async def chunk_vision_enhance(chunks_dic: dict, clean_content: str,
                               vllm_model: str = None, vllm_api_key: str = None,
                               vllm_base_url: str = None,
                               image_width: int = None, image_height: int = None,
                               max_concurrency: int = None) -> dict:
    """
    图片描述增强：提取 chunk 中的 <img> 和 ![]() 图片引用，调用 chat_with_llm 多模态能力
    获取 AI 描述，将描述插入到图片标签之后，同时利用文档上下文提升描述准确性。
    处理完成后将 chunk 中的图片 URL 以换行拼接存入 image_url 字段。
    """
    _model = vllm_model or config['VLLM']['MODEL']
    _api_key = vllm_api_key or config['VLLM']['API_KEY']
    _base_url = vllm_base_url or config['VLLM']['BASE_URL']
    _concurrency = max_concurrency or config["Concurrency"]["LLM_WORKS"]

    chunks_list = []
    uid_order = []
    for uid, chunk in chunks_dic.items():
        chunks_list.append(chunk)
        uid_order.append(uid)

    enhanced_list = await enhance_chunks_with_vision(
        chunks=chunks_list,
        model=_model,
        api_key=_api_key,
        base_url=_base_url,
        markdown_content=clean_content,
        content_field="description",
        max_concurrency=_concurrency,
        image_width=image_width,
        image_height=image_height,
    )

    img_pattern = re.compile(r'(?:<img\s+src="([^"]+)"[^>]*>|!\[[^\]]*\]\(([^)]+)\))')

    result = {}
    for i, enhanced_chunk in enumerate(enhanced_list):
        uid = uid_order[i]
        desc = enhanced_chunk.get("description", enhanced_chunk.get("content", ""))
        chunks_dic[uid]["description"] = desc

        # 提取所有图片 URL，换行拼接存入 image_url
        urls = [u1 or u2 for u1, u2 in img_pattern.findall(desc)]
        if urls:
            chunks_dic[uid]["image_url"] = '\n'.join(urls)

        result[uid] = chunks_dic[uid]

    return result


def _resolve(config_value, param_value):
    """参数解析：优先使用传入值，回退到 config 文件值"""
    return param_value if param_value is not None else config_value


async def chunk_process(
    url, forest_id, company_id, uin, file_id, file_name, file_ext,
    index_name=None,
    # === 预处理 ===
    remove_email=True,
    remove_URL=True,
    remove_empty_line=True,
    # === 分块策略 ===
    mode="smart",
    chunk_token_num=1024,
    min_chunk_tokens=10,
    split_level=2,
    overlap_ratio=0.0,
    regex_pattern=None,
    delimiter="\n!?。；！？",
    enable_heading_in_content=False,
    # === 表格增强 LLM 配置 ===
    llm_enabled=None,
    llm_model=None,
    llm_api_key=None,
    llm_base_url=None,
    llm_timeout=None,
    # === 图片增强 VLLM 配置 ===
    vllm_enabled=None,
    vllm_model=None,
    vllm_api_key=None,
    vllm_base_url=None,
    image_width=None,
    image_height=None,
    # === Embedding 配置 ===
    embedding_model=None,
    embedding_api_key=None,
    embedding_base_url=None,
    # === 并发控制 ===
    eb_max_concurrency=None,
    llm_max_concurrency=None,
):
    """chunk 处理主流程：获取内容 → 清洗 → 分块 → 图片/表格增强 → 向量化

    所有参数均可通过接口传入，未传入时使用 config/chunk_config.yaml 中的默认值。

    Args:
        url: 文件下载地址
        forest_id/company_id/uin/file_id: 知识库标识
        file_name: 文件名 / file_ext: 扩展名
        index_name: ES 索引名

        remove_email: 是否移除邮箱
        remove_URL: 是否移除 URL
        remove_empty_line: 是否合并空行

        mode: 分块策略 smart/basic/advanced/title/strict_regex
        chunk_token_num: 目标 chunk token 数
        min_chunk_tokens: 最小 chunk token 数
        split_level: 标题分块级别 (仅 title 策略)
        overlap_ratio: 相邻 chunk 重叠比例 0.0~1.0
        regex_pattern: 正则表达式 (仅 strict_regex 策略)
        delimiter: 分隔符 (仅 basic 策略)
        enable_heading_in_content: 是否在 chunk 内容中补全父标题

        llm_enabled: 是否启用表格 LLM 增强
        llm_model/api_key/base_url/timeout: 表格增强 LLM 配置
        vllm_enabled: 是否启用图片视觉增强
        vllm_model/api_key/base_url: 视觉增强 VLLM 配置
        image_width/height: 图片缩放尺寸

        embedding_model/api_key/base_url: 向量化配置
        eb_max_concurrency: Embedding 最大并发
        llm_max_concurrency: LLM 最大并发

    Returns:
        {uuid: {fields...}} 含 embedding 向量的 chunks 字典
    """
    logger.info(f"[1/5] 正在获取文件内容... {file_name}")
    content: str = get_content(url)
    logger.info(f"[1/5] 文件获取完成，长度 {len(content)} 字符")

    logger.info(f"[2/5] 正在清洗文本...")
    content: str = simple_clean(
        content=content,
        remove_email=remove_email,
        remove_URL=remove_URL,
        remove_empty_line=remove_empty_line
    )

    # 自动选择分块策略
    if mode == 'auto':
        from corekg_chunk.prompt.prompt import PROMPTS
        from tools.llm_chat import chat_with_llm

        _model = llm_model or config['LLM']['MODEL']
        _api_key = llm_api_key or config['LLM']['API_KEY']
        _base_url = llm_base_url or config['LLM']['BASE_URL']
        _sample = content[:3000]

        try:
            _resp = await chat_with_llm(
                prompt=PROMPTS["auto_select_mode"].format(sample=_sample),
                model=_model,
                api_key=_api_key,
                base_url=_base_url,
            )
            _chosen = _resp.strip().lower()
            if _chosen in ('smart', 'basic', 'advanced', 'title', 'strict_regex', 'slide', 'resume', 'paper', 'laws', 'llm'):
                mode = _chosen
                logger.info(f"[3/5] auto 模式 LLM 选择策略: {mode}")
            else:
                logger.warning(f"[3/5] auto 模式 LLM 返回未知策略 '{_resp[:50]}'，回退 smart")
                mode = 'smart'
        except Exception as e:
            logger.warning(f"[3/5] auto 模式 LLM 调用失败: {e}，回退 smart")
            mode = 'smart'

    # 分块参数解析
    if mode not in ('smart', 'basic', 'advanced', 'title', 'strict_regex', 'slide', 'resume', 'paper', 'laws', 'llm', 'auto'):
        logger.warning(f"未知分块策略 '{mode}'，回退到 smart")
        mode = 'smart'

    logger.info(f"[3/5] 正在分块... 策略: {mode}")

    if mode == 'slide':
        # 按 yg_pos 页码分页
        chunks_list, chunk_metas = get_knowchunks(content=content, strategy=mode)
        clean_content, yg_mapping = extract_yg_pos_mapping(content)
        page_bboxes = compute_page_bboxes(yg_mapping)
        yg_locations = [
            page_bboxes.get(p, None)
            for p in range(len(chunks_list))
        ]
    elif mode == 'resume':
        # 全文不切分
        chunks_list, chunk_metas = get_knowchunks(content=content, strategy=mode)
        clean_content, yg_mapping = extract_yg_pos_mapping(content)
        yg_locations = find_yg_locations_for_chunks(chunks_list, clean_content, yg_mapping)
    elif mode == 'llm':
        # 大模型语义切块
        from corekg_chunk.pipeline.knowchunk import split_markdown_to_chunks_llm
        clean_content, yg_mapping = extract_yg_pos_mapping(content)
        _model = llm_model or config['LLM']['MODEL']
        _api_key = llm_api_key or config['LLM']['API_KEY']
        _base_url = llm_base_url or config['LLM']['BASE_URL']
        chunks_list = await split_markdown_to_chunks_llm(
            content=clean_content,
            model=_model,
            api_key=_api_key,
            base_url=_base_url,
            chunk_token_num=chunk_token_num,
            min_chunk_tokens=min_chunk_tokens,
        )
        chunk_metas = [{} for _ in chunks_list]
        yg_locations = find_yg_locations_for_chunks(chunks_list, clean_content, yg_mapping)
    else:
        clean_content, yg_mapping = extract_yg_pos_mapping(content)
        chunks_list, chunk_metas = get_knowchunks(
            content=clean_content,
            strategy=mode,
            chunk_token_num=chunk_token_num,
            min_chunk_tokens=min_chunk_tokens,
            split_level=split_level,
            overlap_ratio=overlap_ratio,
            regex_pattern=regex_pattern,
            delimiter=delimiter,
            enable_heading_in_content=enable_heading_in_content,
        )
    logger.info(f"[3/5] 分块完成，共 {len(chunks_list)} 个 chunk")

    yg_locations = find_yg_locations_for_chunks(chunks_list, clean_content, yg_mapping)
    logger.info(f"[3/5] 坐标映射完成，{sum(1 for y in yg_locations if y)}/{len(yg_locations)} 命中")

    chunks_dic = knowchunk_format_conversion(
        chunks_list, chunk_metas, yg_locations,
        forest_id, company_id, uin, file_id, file_name, file_ext
    )
    logger.info(f"[3/5] 格式转换完成")

    # 图片增强
    vllm_status = _resolve(config.get("VLLM", {}).get("STATUS", False), vllm_enabled)
    if vllm_status:
        logger.info(f"[4/5] 正在图片增强...")
        chunks_dic = await chunk_vision_enhance(
            chunks_dic=chunks_dic,
            clean_content=clean_content,
            vllm_model=vllm_model,
            vllm_api_key=vllm_api_key,
            vllm_base_url=vllm_base_url,
            image_width=image_width,
            image_height=image_height,
            max_concurrency=llm_max_concurrency or config["Concurrency"]["LLM_WORKS"],
        )
        logger.info(f"[4/5] 图片增强完成")
    else:
        logger.info(f"[4/5] 图片增强已跳过，仅提取 image_url")
        img_pattern = re.compile(r'(?:<img\s+src="([^"]+)"[^>]*>|!\[[^\]]*\]\(([^)]+)\))')
        for chunk in chunks_dic.values():
            urls = [u1 or u2 for u1, u2 in img_pattern.findall(chunk.get("description", ""))]
            if urls:
                chunk["image_url"] = '\n'.join(urls)

    # 表格增强
    llm_status = _resolve(config.get("LLM", {}).get("STATUS", False), llm_enabled)
    if llm_status:
        logger.info(f"[4.2/5] 正在表格增强...")
        chunks_dic = await enhance_table_chunks(
            chunks_dic=chunks_dic,
            model=_resolve(config['LLM']['MODEL'], llm_model),
            api_key=_resolve(config['LLM']['API_KEY'], llm_api_key),
            base_url=_resolve(config['LLM']['BASE_URL'], llm_base_url),
            max_concurrency=llm_max_concurrency or config["Concurrency"]["LLM_WORKS"],
        )
        logger.info(f"[4.2/5] 表格增强完成")

    # 向量化
    logger.info(f"[5/5] 正在向量化... 共 {len(chunks_dic)} 个 chunk")
    chunks_emb_dic = await chunk_emb(
        chunks_dic=chunks_dic,
        eb_max_concurrency=eb_max_concurrency or config["Concurrency"]["EB_WORKS"],
        embedding_model=embedding_model,
        embedding_api_key=embedding_api_key,
        embedding_base_url=embedding_base_url,
    )
    logger.info(f"[5/5] 向量化完成，共 {len(chunks_emb_dic)} 个 chunk")

    return chunks_emb_dic

