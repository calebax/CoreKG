#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
图片视觉增强模块

在分块完成后，为 chunks 中的图片添加 AI 生成的描述。
使用 chat_with_llm 多模态能力生成图片描述，支持从文档中提取图片上下文。
"""

import re
import os
import asyncio
import logging
from typing import List, Dict, Any, Optional

from corekg_chunk.utils import image_url_base64
from corekg_chunk.prompt.prompt import PROMPTS

logger = logging.getLogger(__name__)


def extract_image_references(text: str) -> List[tuple]:
    """
    提取文本中的图片引用
    支持两种格式：
    1. HTML: <img src="/minio/xxx.jpg" ...>
    2. Markdown: ![alt](path)
    返回: [(full_tag, img_path, start_pos, end_pos), ...]
    """
    matches = []

    html_pattern = r'<img\s+src="([^"]+)"[^>]*>'
    for match in re.finditer(html_pattern, text):
        matches.append((match.group(0), match.group(1), match.start(), match.end()))

    md_pattern = r'!\[([^\]]*)\]\(([^)]+)\)'
    for match in re.finditer(md_pattern, text):
        matches.append((match.group(0), match.group(2), match.start(), match.end()))

    matches.sort(key=lambda x: x[2])
    return matches


def _find_caption_line_end(content: str, img_tag_end: int) -> int:
    """
    查找图片标签后标题行的结束位置。
    利用 alt 属性来判断下一行是否是标题。
    """
    img_tag_start = content.rfind('<img', 0, img_tag_end)
    if img_tag_start == -1:
        return img_tag_end

    img_tag = content[img_tag_start:img_tag_end]
    alt_match = re.search(r'alt="([^"]*)"', img_tag)
    if not alt_match:
        return img_tag_end

    alt_text = alt_match.group(1)
    if not alt_text or alt_text == '图片':
        return img_tag_end

    remaining = content[img_tag_end:]
    if not remaining.startswith('\n'):
        return img_tag_end
    if len(remaining) > 1 and remaining[1] == '\n':
        return img_tag_end

    line_end = remaining.find('\n', 1)
    if line_end == -1:
        next_line = remaining[1:].strip()
    else:
        next_line = remaining[1:line_end].strip()

    alt_normalized = ' '.join(alt_text.split())
    next_line_normalized = ' '.join(next_line.split())

    if alt_normalized == next_line_normalized:
        caption_end = img_tag_end + (line_end if line_end != -1 else len(remaining))
        logger.debug(f"检测到图片标题 (通过 alt 匹配): {next_line[:50]}...")
        return caption_end

    return img_tag_end


async def _describe_single_image(img_url: str, context: str = None,
                                  model: str = None, api_key: str = None,
                                  base_url: str = None, semaphore: asyncio.Semaphore = None,
                                  image_width: int = None, image_height: int = None) -> Optional[str]:
    """下载单张图片并调用多模态 LLM 获取描述"""
    from tools.llm_chat import chat_with_llm

    async def _call():
        try:
            img_base64, mime_type = image_url_base64(img_url, width=image_width or 1024, height=image_height)
            prompt_text = PROMPTS["image_description"].format(context=context or "无")

            prompt = [
                {"type": "image_url", "image_url": {"url": f"data:{mime_type};base64,{img_base64}"}},
                {"type": "text", "text": prompt_text}
            ]

            result = await chat_with_llm(prompt=prompt, model=model, api_key=api_key, base_url=base_url)
            result = result.replace('\n', '').strip()
            logger.info(f"图片描述生成成功: {img_url[:60]}... -> {result[:80]}...")
            return result
        except Exception as e:
            logger.warning(f"图片描述生成失败 {img_url[:60]}: {e}")
            return None

    if semaphore:
        async with semaphore:
            return await _call()
    return await _call()


async def call_llm_vision_batch(image_paths: List[str],
                                 image_contexts: Dict[str, str] = None,
                                 model: str = None, api_key: str = None,
                                 base_url: str = None,
                                 max_concurrency: int = 3,
                                 image_width: int = None, image_height: int = None) -> Dict[str, Optional[str]]:
    """
    批量调用 chat_with_llm 多模态能力获取图片描述
    参数:
        image_paths: 图片 URL 列表
        image_contexts: {img_url: context_summary} 上下文映射
        model/api_key/base_url: LLM 配置
        max_concurrency: 最大并发数
    返回: {img_url: description} 字典
    """
    if not image_paths:
        return {}
    if image_contexts is None:
        image_contexts = {}

    semaphore = asyncio.Semaphore(max_concurrency)
    tasks = [_describe_single_image(
        img_url=path,
        context=image_contexts.get(path),
        model=model, api_key=api_key, base_url=base_url,
        semaphore=semaphore,
        image_width=image_width, image_height=image_height,
    ) for path in image_paths]

    results = await asyncio.gather(*tasks)
    return {path: desc for path, desc in zip(image_paths, results)}


async def enhance_chunks_with_vision(
    chunks: List[Dict[str, Any]],
    model: str = None,
    api_key: str = None,
    base_url: str = None,
    description_format: str = "[图片描述: {desc}]",
    max_concurrency: int = None,
    markdown_content: str = None,
    content_field: str = "content",
    image_width: int = None,
    image_height: int = None,
) -> List[Dict[str, Any]]:
    """
    为分块添加图片视觉增强描述（批量处理，使用 chat_with_llm 多模态）
    参数:
        chunks: 分块列表
        model/api_key/base_url: LLM 配置，默认从 chunk_config.yaml 读取
        description_format: 描述格式模板
        max_concurrency: 最大并发数
        markdown_content: 完整的 Markdown 文档内容，用于提取上下文
        content_field: chunk 中存储文本内容的字段名
    返回: 增强后的分块列表
    """
    if not chunks:
        return chunks
    if max_concurrency is None:
        max_concurrency = int(os.getenv('VISION_MAX_CONCURRENCY', '3'))

    logger.info(f"开始图片视觉增强，共 {len(chunks)} 个分块，最大并发: {max_concurrency}")

    # 第一步：收集所有需要处理的图片
    chunk_images = {}
    image_to_tag = {}

    for i, chunk in enumerate(chunks):
        if not isinstance(chunk, dict) or content_field not in chunk:
            continue
        img_refs = extract_image_references(chunk[content_field])
        if img_refs:
            chunk_images[i] = img_refs
            for full_tag, img_path, _, _ in img_refs:
                if img_path not in image_to_tag:
                    image_to_tag[img_path] = full_tag

    if not image_to_tag:
        logger.info("未发现需要处理的图片")
        return chunks

    logger.info(f"共发现 {len(image_to_tag)} 个唯一图片")

    # 第二步：提取图片上下文信息
    image_contexts = {}
    if markdown_content:
        try:
            from corekg_chunk.pipeline.image_context_extractor import ImageContextExtractor
            extractor = ImageContextExtractor(markdown_content)
            for img_path, full_tag in image_to_tag.items():
                context_info = extractor.extract_context_for_image(full_tag, img_path)
                if context_info['context_summary']:
                    image_contexts[img_path] = context_info['context_summary']
            logger.info(f"成功提取 {len(image_contexts)} 个图片的上下文信息")
        except Exception as e:
            logger.error(f"提取图片上下文失败: {e}")

    # 第三步：批量调用 LLM 获取图片描述
    all_image_paths = list(image_to_tag.keys())
    all_descriptions = await call_llm_vision_batch(
        image_paths=all_image_paths,
        image_contexts=image_contexts,
        model=model, api_key=api_key, base_url=base_url,
        max_concurrency=max_concurrency,
        image_width=image_width, image_height=image_height,
    )

    # 第四步：将描述应用到对应的分块
    enhanced_chunks = []
    enhanced_count = 0

    for i, chunk in enumerate(chunks):
        chunk_img_refs = chunk_images.get(i)
        if chunk_img_refs is None:
            enhanced_chunks.append(chunk)
            continue

        enhanced_content = chunk[content_field]
        for full_tag, img_path, start, end in reversed(chunk_img_refs):
            desc = all_descriptions.get(img_path)
            if desc:
                insert_pos = _find_caption_line_end(enhanced_content, end)
                enhancement = "<br>\n" + description_format.format(desc=desc) + "\n"
                enhanced_content = enhanced_content[:insert_pos] + enhancement + enhanced_content[insert_pos:]
                enhanced_count += 1

        enhanced_chunk = chunk.copy()
        enhanced_chunk[content_field] = enhanced_content
        enhanced_chunks.append(enhanced_chunk)

    logger.info(f"图片视觉增强完成，共增强 {enhanced_count} 个图片")
    return enhanced_chunks
