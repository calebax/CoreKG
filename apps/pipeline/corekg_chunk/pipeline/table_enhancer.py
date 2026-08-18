#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
表格描述增强模块

为 table 类型的 chunk 生成简短描述，结合文档上下文，方便检索时识别表格内容。
"""

import re
import asyncio
import logging
from typing import List, Dict, Any, Optional

from corekg_chunk.prompt.prompt import PROMPTS

logger = logging.getLogger(__name__)


async def _describe_single_table(table_content: str, context: str = None,
                                   model: str = None, api_key: str = None,
                                   base_url: str = None,
                                   semaphore: asyncio.Semaphore = None) -> Optional[str]:
    """调用 LLM 为单个表格生成简短描述"""
    from tools.llm_chat import chat_with_llm

    async def _call():
        try:
            prompt = PROMPTS["table_description"].format(
                table=table_content[:3000],
                context=context or "无"
            )
            result = await chat_with_llm(prompt=prompt, model=model, api_key=api_key, base_url=base_url)
            result = result.strip()
            if result:
                logger.info(f"表格描述生成成功: {result[:60]}...")
            return result
        except Exception as e:
            logger.warning(f"表格描述生成失败: {e}")
            return None

    if semaphore:
        async with semaphore:
            return await _call()
    return await _call()


async def enhance_table_chunks(chunks_dic: dict,
                                model: str = None, api_key: str = None,
                                base_url: str = None,
                                max_concurrency: int = 3) -> dict:
    """
    为所有 table 类型的 chunk 生成简短描述，结合上文上下文。
    参数:
        chunks_dic: corekg_chunk 格式的 {uid: {type, description, title_level, ...}}
        model/api_key/base_url: LLM 配置
        max_concurrency: 最大并发数
    返回:
        增强后的 chunks_dic（原地修改 + 返回）
    """
    chunk_ids = list(chunks_dic.keys())
    table_tasks = []

    for idx, chunk_id in enumerate(chunk_ids):
        chunk = chunks_dic[chunk_id]
        if chunk.get('type') != 'table':
            continue

        context = None
        table_title = chunk.get('title_level')
        for i in range(idx - 1, -1, -1):
            prev = chunks_dic[chunk_ids[i]]
            if prev.get('type') == 'chunk' and prev.get('description', '').strip():
                if table_title and prev.get('title_level') == table_title:
                    context = prev.get('description', '')[-500:]
                    break
                elif context is None:
                    context = prev.get('description', '')[-500:]

        table_tasks.append((chunk_id, chunk['description'], context))

    if not table_tasks:
        logger.info("未发现需要处理的表格")
        return chunks_dic

    logger.info(f"共发现 {len(table_tasks)} 个表格待处理")

    semaphore = asyncio.Semaphore(max_concurrency)
    tasks = [_describe_single_table(
        table_content=tc, context=ctx,
        model=model, api_key=api_key, base_url=base_url,
        semaphore=semaphore
    ) for _, tc, ctx in table_tasks]

    descriptions = await asyncio.gather(*tasks)

    for (chunk_id, _, _), desc in zip(table_tasks, descriptions):
        if desc:
            original = chunks_dic[chunk_id]["description"]
            table_pos = original.find('<table')
            if table_pos != -1:
                chunks_dic[chunk_id]["description"] = (
                    original[:table_pos] + f"[表格摘要: {desc}]\n" + original[table_pos:]
                )
                chunks_dic[chunk_id]["table"] = original
            else:
                chunks_dic[chunk_id]["description"] = f"[表格摘要: {desc}]\n{original}"
                chunks_dic[chunk_id]["table"] = original

    logger.info(f"表格描述生成完成，共处理 {sum(1 for d in descriptions if d)}/{len(descriptions)} 个表格")
    return chunks_dic
