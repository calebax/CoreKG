# -*- coding: utf-8 -*-
"""
AST 智能分块、标题分块、高级分块、正则分块等策略，

依赖:
    pip install markdown-it-py tiktoken
"""
import os
import re
import logging
import tiktoken

from markdown import markdown as md_to_html

try:
    from markdown_it import MarkdownIt
    from markdown_it.tree import SyntaxTreeNode
    MARKDOWN_IT_AVAILABLE = True
except ImportError:
    MARKDOWN_IT_AVAILABLE = False

logger = logging.getLogger(__name__)

# ===================== token 计数 =====================
def load_local_tiktoken():
    """从本地路径加载tiktoken编码器"""
    local_file_path = "./corekg_chunk/tiktoken_cache/cl100k_base.tiktoken"
    
    if not os.path.exists(local_file_path):
        raise FileNotFoundError(f"本地编码文件不存在: {local_file_path}")
    
    # 读取本地文件
    with open(local_file_path, 'rb') as f:
        contents = f.read()
    
    # 使用tiktoken的加载函数
    from tiktoken.load import load_tiktoken_bpe
    from tiktoken.registry import get_encoding
    
    # 直接使用Encoding对象
    mergeable_ranks = load_tiktoken_bpe(local_file_path)
    
    # 创建Encoding实例
    from tiktoken.core import Encoding
    encoder = Encoding(
        name="cl100k_base",
        pat_str=r"""'(?:[sdmt]|ll|ve|re)| ?\p{L}+| ?\p{N}+| ?[^\s\p{L}\p{N}]+|\s+(?!\S)|\s+""",
        mergeable_ranks=mergeable_ranks,
        special_tokens={
            "<|endoftext|>": 100257,
            "<|fim_prefix|>": 100258,
            "<|fim_middle|>": 100259,
            "<|fim_suffix|>": 100260,
            "<|endofprompt|>": 100276
        }
    )
    
    return encoder

# 使用本地文件
try:
    encoder = load_local_tiktoken()
    logger.info("成功从本地加载编码器")
except Exception as e:
    logger.error(f"加载失败: {e}")
    # 降级方案
    encoder = None


def num_tokens_from_string(string: str) -> int:
    """使用 tiktoken cl100k_base 编码计算文本的 token 数量"""
    try:
        return len(encoder.encode(string))
    except Exception:
        return 0


# ===================== AST 辅助函数 =====================

def _extract_text_from_node(node):
    """从 markdown-it AST 节点递归提取纯文本，保留行内格式（粗体/斜体/链接/行内代码）"""
    if hasattr(node, 'content') and node.content:
        return node.content

    text_parts = []
    if hasattr(node, 'children') and node.children:
        for child in node.children:
            if child.type == "text":
                text_parts.append(child.content)
            elif child.type == "code_inline":
                text_parts.append("`" + child.content + "`")
            elif child.type == "strong":
                text_parts.append("**" + _extract_text_from_node(child) + "**")
            elif child.type == "em":
                text_parts.append("*" + _extract_text_from_node(child) + "*")
            elif child.type == "link":
                link_text = _extract_text_from_node(child)
                text_parts.append("[" + link_text + "](" + (child.attrGet('href') or '') + ")")
            else:
                text_parts.append(_extract_text_from_node(child))

    return "".join(text_parts)


def _render_table_from_ast(table_node):
    """将 AST 表格节点渲染为 HTML table 字符串"""
    try:
        table_md = []
        cells = []
        for child in table_node.children:
            if child.type == "thead":
                for row in child.children:
                    if row.type == "tr":
                        cells = []
                        for cell in row.children:
                            if cell.type in ["th", "td"]:
                                cells.append(_extract_text_from_node(cell))
                        table_md.append("| " + " | ".join(cells) + " |")

                if table_md:
                    separator = "| " + " | ".join(["---"] * len(cells)) + " |"
                    table_md.append(separator)

            elif child.type == "tbody":
                for row in child.children:
                    if row.type == "tr":
                        cells = []
                        for cell in row.children:
                            if cell.type in ["th", "td"]:
                                cells.append(_extract_text_from_node(cell))
                        table_md.append("| " + " | ".join(cells) + " |")

        table_markdown = "\n".join(table_md)
        return md_to_html(table_markdown, extensions=['markdown.extensions.tables'])

    except Exception:
        return _extract_text_from_node(table_node)


def _render_list_from_ast(list_node):
    """将 AST 列表节点渲染为 markdown 列表文本（有序/无序）"""
    list_items = []
    list_type = list_node.attrGet('type') or 'bullet'

    for i, item in enumerate(list_node.children):
        if item.type == "list_item":
            item_content = _extract_text_from_node(item)
            if list_type == 'ordered':
                list_items.append(str(i + 1) + ". " + item_content)
            else:
                list_items.append("- " + item_content)

    return "\n".join(list_items)


def _render_blockquote_from_ast(blockquote_node):
    """将 AST 引用块节点渲染为 markdown blockquote 文本"""
    content = _extract_text_from_node(blockquote_node)
    lines = content.split('\n')
    return '\n'.join("> " + line for line in lines)


def _add_missing_parent_headings(chunk_content, headers):
    """为分块内容补充缺失的父级标题路径，使 chunk 保持上下文完整性
    Args:
        chunk_content: 当前分块的文本内容
        headers: 父级标题字典 {level: title}，如 {1: "第一章", 2: "1.1 背景"}
    Returns:
        补充了父级标题的文本（已存在的标题不重复添加）
    """
    if not headers:
        return chunk_content

    existing_headings = set()
    for line in chunk_content.split('\n'):
        line = line.strip()
        if line.startswith('#'):
            level = len(line) - len(line.lstrip('#'))
            if 0 < level <= 6:
                heading_text = line.lstrip('#').strip()
                existing_headings.add(str(level) + ":" + heading_text)

    missing_heading_lines = []
    for level in sorted(headers.keys()):
        heading_key = str(level) + ":" + headers[level]
        if heading_key not in existing_headings:
            heading_prefix = '#' * level
            missing_heading_lines.append(heading_prefix + " " + headers[level])

    if missing_heading_lines:
        missing_heading_text = '\n'.join(missing_heading_lines)
        chunk_content = missing_heading_text + "\n\n" + chunk_content

    return chunk_content


def _update_context_stack(context_stack, level, title):
    """更新标题上下文栈：移除 >= 当前级别的旧标题，压入新标题"""
    while context_stack and context_stack[-1]['level'] >= level:
        context_stack.pop()
    context_stack.append({'level': level, 'title': title})


def _process_non_heading_node(node, chunk_token_num):
    """处理非标题的 AST 节点，根据节点类型渲染并判断是否需要在此断开
    返回: (content, should_break)
    """
    node_type = node.type
    should_break = False
    content = ""

    if node_type == "table":
        content = _render_table_from_ast(node)
        table_tokens = num_tokens_from_string(content)
        if table_tokens > chunk_token_num:
            should_break = True

    elif node_type == "code_block":
        content = "```" + (node.info or '') + "\n" + node.content + "```"

    elif node_type == "blockquote":
        content = _render_blockquote_from_ast(node)

    elif node_type in ("list", "bullet_list", "ordered_list"):
        content = _render_list_from_ast(node)

    elif node_type == "paragraph":
        content = _extract_text_from_node(node)

    elif node_type == "hr":
        content = "---"
        should_break = True

    else:
        content = _extract_text_from_node(node)

    return content, should_break


def _finalize_ast_chunk(chunk_parts, context_stack, enable_heading_in_content=False):
    """完成 AST chunk 的格式化输出：拼接内容、提取标题元数据"""
    chunk_content = "\n\n".join(chunk_parts).strip()
    headers = {item['level']: item['title'] for item in context_stack}

    if enable_heading_in_content and headers:
        chunk_content = _add_missing_parent_headings(chunk_content, headers)

    return {
        'content': chunk_content,
        'heading_metadata': {
            'headers': headers,
            'level': max(headers.keys()) if headers else 0
        }
    }


# ===================== 标题/高级分块辅助函数 =====================

def _render_node_content(node):
    """渲染单个 AST 节点为文本，根据节点类型分发到对应的渲染函数"""
    if node.type == "heading":
        title_text = _extract_text_from_node(node)
        return node.markup + " " + title_text
    elif node.type == "table":
        return _render_table_from_ast(node)
    elif node.type == "code_block":
        return "```" + (node.info or '') + "\n" + node.content + "```"
    elif node.type == "blockquote":
        return _render_blockquote_from_ast(node)
    elif node.type in ["bullet_list", "ordered_list"]:
        return _render_list_from_ast(node)
    elif node.type == "paragraph":
        return _extract_text_from_node(node)
    elif node.type == "hr":
        return "---"
    else:
        return _extract_text_from_node(node)


def _extract_nodes_with_header_info(tree, headers_to_split_on):
    """遍历 AST 树，为每个节点挂载当前所处的标题路径和是否为分块边界
    返回: [{'node', 'type', 'level', 'title', 'headers', 'is_split_boundary', 'content'}, ...]
    """
    nodes_with_headers = []
    current_headers = {}

    for node in tree.children:
        if node.type == "heading":
            level = int(node.tag[1])
            title = _extract_text_from_node(node)
            current_headers = {k: v for k, v in current_headers.items() if k < level}
            current_headers[level] = title
            is_split_boundary = level in headers_to_split_on

            nodes_with_headers.append({
                'node': node,
                'type': 'heading',
                'level': level,
                'title': title,
                'headers': current_headers.copy(),
                'is_split_boundary': is_split_boundary,
                'content': node.markup + " " + title
            })
        else:
            content = _render_node_content(node)
            if content.strip():
                nodes_with_headers.append({
                    'node': node,
                    'type': node.type,
                    'headers': current_headers.copy(),
                    'is_split_boundary': False,
                    'content': content
                })

    return nodes_with_headers


def _split_by_header_levels(nodes_with_headers, headers_to_split_on):
    """根据标题边界将带标题信息的节点列表切分为多个 chunk
    连续标题后无正文内容的，不作为分块边界（避免产生空块）
    """
    chunks = []
    current_chunk = {'headers': {}, 'nodes': []}

    i = 0
    while i < len(nodes_with_headers):
        node_info = nodes_with_headers[i]

        if node_info['is_split_boundary']:
            if node_info['type'] == 'heading':
                has_following_content = False
                j = i + 1
                while j < len(nodes_with_headers):
                    next_node = nodes_with_headers[j]
                    if next_node.get('type') == 'heading':
                        j += 1
                        continue
                    if next_node.get('content', '').strip():
                        has_following_content = True
                        break
                    j += 1

                if not has_following_content:
                    current_chunk['nodes'].append(node_info)
                    if node_info['headers']:
                        current_chunk['headers'] = node_info['headers'].copy()
                    i += 1
                    continue

            if (current_chunk['nodes'] and
                    any(n for n in current_chunk['nodes'] if n['content'].strip())):
                chunks.append(current_chunk)
                current_chunk = {'headers': {}, 'nodes': []}

        if node_info['headers']:
            current_chunk['headers'] = node_info['headers'].copy()

        current_chunk['nodes'].append(node_info)
        i += 1

    if current_chunk['nodes'] and any(n for n in current_chunk['nodes'] if n['content'].strip()):
        chunks.append(current_chunk)

    return chunks


def _render_header_chunk(chunk_info):
    """渲染基于标题切分的 chunk：若块内无标题，在块首自动注入最近上下文标题"""
    content_parts = []
    chunk_has_header = any(node['type'] == 'heading' for node in chunk_info.get('nodes', []))

    if not chunk_has_header and chunk_info.get('headers'):
        max_level = max(chunk_info['headers'].keys())
        context_header = '#' * max_level + " " + chunk_info['headers'][max_level]
        if context_header:
            content_parts.append(context_header)

    for node_info in chunk_info.get('nodes', []):
        if node_info.get('content', '').strip():
            content_parts.append(node_info['content'])

    return "\n\n".join(content_parts).strip()


# ===================== 基础分块 =====================

def _extract_tables_and_remainder_md(txt):
    lines = txt.split('\n')
    tables = []
    remainder_lines = []
    in_table = False
    current_table = []

    for idx, line in enumerate(lines):
        stripped_line = line.strip()
        is_table_line = stripped_line.startswith('|') and stripped_line.endswith('|')

        is_separator_line = False
        if is_table_line and '-' in stripped_line:
            parts = [p.strip() for p in stripped_line[1:-1].split('|')]
            if all(set(p) <= set('-:') for p in parts if p) and parts:
                is_separator_line = True

        if is_table_line or (in_table and stripped_line):
            if not in_table and is_table_line and not is_separator_line:
                next_is_table = False
                if idx + 1 < len(lines):
                    next_stripped = lines[idx + 1].strip()
                    if next_stripped.startswith('|') and next_stripped.endswith('|') and '-' in next_stripped:
                        parts_next = [p.strip() for p in next_stripped[1:-1].split('|')]
                        if all(set(p) <= set('-:') for p in parts_next if p) and parts_next:
                            next_is_table = True
                if next_is_table:
                    in_table = True
                    current_table.append(line)
                else:
                    remainder_lines.append(line)
            elif in_table:
                current_table.append(line)
                if not is_table_line and not stripped_line:
                    tables.append("\n".join(current_table))
                    current_table = []
                    in_table = False
                    remainder_lines.append(line)
            else:
                remainder_lines.append(line)

        elif in_table and not stripped_line:
            tables.append("\n".join(current_table))
            current_table = []
            in_table = False
            remainder_lines.append(line)
        elif in_table and not is_table_line:
            tables.append("\n".join(current_table))
            current_table = []
            in_table = False
            remainder_lines.append(line)
        else:
            remainder_lines.append(line)

    if current_table:
        tables.append("\n".join(current_table))

    return "\n".join(remainder_lines), tables


def split_markdown_to_chunks(txt, chunk_token_num=128, delimiter="\n!?。；！？"):
    """基础分块策略：提取 markdown 表格渲染为 HTML 独立成块，剩余文本按换行符切分后贪心合并
    Args:
        txt: markdown 文本
        chunk_token_num: 目标 token 数上限，默认 128
        delimiter: 分隔符字符串（未在当前实现中使用，保留兼容）
    Returns:
        字符串列表，每个元素为一个 chunk
    """
    if not txt or not txt.strip():
        return []

    remainder_text, extracted_tables = _extract_tables_and_remainder_md(txt)

    processed_chunks = []
    for table_md in extracted_tables:
        if table_md.strip():
            try:
                table_html = md_to_html(table_md, extensions=['markdown.extensions.tables'])
                processed_chunks.append(table_html)
            except Exception:
                processed_chunks.append(table_md)

    initial_sections = []
    if remainder_text and remainder_text.strip():
        for sec_line in remainder_text.split("\n"):
            line_content = sec_line.strip()
            if not line_content:
                initial_sections.append(sec_line)
                continue

            if num_tokens_from_string(sec_line) > 3 * chunk_token_num:
                mid_point = len(sec_line) // 2
                initial_sections.append(sec_line[:mid_point])
                initial_sections.append(sec_line[mid_point:])
            else:
                initial_sections.append(sec_line)

    final_text_chunks = []
    current_chunk_parts = []
    current_token_count = 0

    for section_text in initial_sections:
        section_token_count = num_tokens_from_string(section_text)

        if not section_text.strip() and not current_chunk_parts:
            continue

        if current_token_count + section_token_count <= chunk_token_num:
            current_chunk_parts.append(section_text)
            current_token_count += section_token_count
        else:
            if current_chunk_parts:
                final_text_chunks.append("\n".join(current_chunk_parts).strip())

            if section_token_count > chunk_token_num and section_token_count <= 3 * chunk_token_num:
                final_text_chunks.append(section_text.strip())
                current_chunk_parts = []
                current_token_count = 0
            elif section_token_count > 3 * chunk_token_num:
                mid = len(section_text) // 2
                final_text_chunks.append(section_text[:mid].strip())
                final_text_chunks.append(section_text[mid:].strip())
                current_chunk_parts = []
                current_token_count = 0
            else:
                current_chunk_parts = [section_text]
                current_token_count = section_token_count

    if current_chunk_parts:
        final_text_chunks.append("\n".join(current_chunk_parts).strip())

    all_chunks = [chunk for chunk in processed_chunks if chunk.strip()]
    all_chunks.extend([chunk for chunk in final_text_chunks if chunk.strip()])

    return all_chunks


# ===================== 智能分块（基于AST语法树） =====================

def split_markdown_to_chunks_smart(txt, chunk_token_num=256, min_chunk_tokens=10,
                                    enable_heading_in_content=False):
    """智能分块策略：基于 markdown-it-py AST 语法树的智能分块
    标题作为分块边界，逐节点遍历，超过 chunk_token_num(256) 时在前一个标题处断开。
    表格/代码块/列表保持完整不拆分。每个 chunk 携带 heading_metadata 记录标题层级路径。
    Args:
        txt: markdown 文本
        chunk_token_num: 目标 token 数上限，默认 256
        min_chunk_tokens: 最小 token 阈值，低于此值不独立成块，默认 10
        enable_heading_in_content: 是否在 chunk 内容中补全父标题路径
    Returns:
        dict 列表 [{'content': str, 'heading_metadata': {'headers': {level: title}, 'level': int}}, ...]
        当 markdown-it-py 不可用时回退到基础策略，异常时也回退到基础策略
    """
    if not MARKDOWN_IT_AVAILABLE:
        return split_markdown_to_chunks(txt, chunk_token_num)

    if not txt or not txt.strip():
        return []

    md = MarkdownIt("commonmark", {"breaks": True, "html": True})
    md.enable(['table'])

    try:
        tokens = md.parse(txt)
        tree = SyntaxTreeNode(tokens)

        chunks = []
        current_chunk = []
        current_tokens = 0
        context_stack = []

        for node in tree.children:
            node_type = node.type

            if node_type == "heading":
                if current_chunk and current_tokens >= min_chunk_tokens:
                    chunk_content = _finalize_ast_chunk(current_chunk, context_stack,
                                                        enable_heading_in_content)
                    if isinstance(chunk_content, dict) and chunk_content.get('content', '').strip():
                        chunks.append(chunk_content)
                    elif isinstance(chunk_content, str) and chunk_content.strip():
                        chunks.append(chunk_content)
                    current_chunk = []
                    current_tokens = 0

                level = int(node.tag[1])
                title_text = _extract_text_from_node(node)
                _update_context_stack(context_stack, level, title_text)

                chunk_data = node.markup + " " + title_text
                current_chunk.append(chunk_data)
                current_tokens = num_tokens_from_string(chunk_data)
            else:
                chunk_data, _ = _process_non_heading_node(node, chunk_token_num)

                if chunk_data:
                    chunk_tokens = num_tokens_from_string(chunk_data)

                    if (current_tokens + chunk_tokens > chunk_token_num and
                            current_chunk and current_tokens >= min_chunk_tokens):
                        chunk_content = _finalize_ast_chunk(current_chunk, context_stack,
                                                            enable_heading_in_content)
                        if isinstance(chunk_content, dict) and chunk_content.get('content', '').strip():
                            chunks.append(chunk_content)
                        elif isinstance(chunk_content, str) and chunk_content.strip():
                            chunks.append(chunk_content)
                        current_chunk = []
                        current_tokens = 0

                    current_chunk.append(chunk_data)
                    current_tokens += chunk_tokens

        if current_chunk:
            chunk_content = _finalize_ast_chunk(current_chunk, context_stack,
                                                enable_heading_in_content)
            if isinstance(chunk_content, dict) and chunk_content.get('content', '').strip():
                chunks.append(chunk_content)
            elif isinstance(chunk_content, str) and chunk_content.strip():
                chunks.append(chunk_content)

        result = []
        for chunk in chunks:
            if isinstance(chunk, dict):
                if chunk.get('content', '').strip():
                    result.append(chunk)
            elif isinstance(chunk, str) and chunk.strip():
                result.append(chunk)
        return result

    except Exception:
        return split_markdown_to_chunks(txt, chunk_token_num)


# ===================== 标题分块（按标题级别切割） =====================

def split_markdown_to_chunks_title(txt, chunk_token_num=256, min_chunk_tokens=10,
                                    split_level=2, enable_heading_in_content=False):
    """标题分块策略：严格按指定标题级别切割，不合并小块也不拆分大块
    若指定级别只产生 1 个块，自动回退到上一级（H2→H1）。
    Args:
        txt: markdown 文本
        chunk_token_num: token 数（此策略仅作参考，不控制大小）
        min_chunk_tokens: 最小 token（此策略仅作参考，不控制大小）
        split_level: 切割的标题级别 1-6，默认 2（即 H2）
        enable_heading_in_content: 是否在内容中补全父标题
    Returns:
        dict 列表 [{'content': str, 'heading_metadata': {'headers': {level: title}, 'level': int}}, ...]
        当 markdown-it-py 不可用或异常时回退到智能策略
    """
    if not MARKDOWN_IT_AVAILABLE:
        return split_markdown_to_chunks(txt, chunk_token_num)

    if not txt or not txt.strip():
        return []

    md = MarkdownIt("commonmark", {"breaks": True, "html": True})
    md.enable(['table'])

    try:
        tokens = md.parse(txt)
        tree = SyntaxTreeNode(tokens)

        chunks = None
        current_level = split_level

        while current_level >= 1:
            headers_to_split_on = [current_level]
            nodes_with_headers = _extract_nodes_with_header_info(tree, headers_to_split_on)
            chunks = _split_by_header_levels(nodes_with_headers, headers_to_split_on)

            if len(chunks) > 1 or current_level == 1:
                break

            current_level -= 1

        final_chunks = []
        for chunk_info in chunks:
            content = _render_header_chunk(chunk_info)
            if content.strip():
                headers = chunk_info.get('headers', {})

                if enable_heading_in_content and headers:
                    content = _add_missing_parent_headings(content, headers)

                chunk_data = {
                    'content': content,
                    'heading_metadata': {
                        'headers': headers,
                        'level': max(headers.keys()) if headers else 0
                    }
                }
                final_chunks.append(chunk_data)

        return final_chunks

    except Exception:
        return split_markdown_to_chunks_smart(txt, chunk_token_num, min_chunk_tokens)


# ===================== 高级分块（动态大小优化） =====================

def _has_special_content(chunk):
    """检查分块是否包含表格、代码块或数学公式等不可被拆分打断的特殊内容"""
    for node_info in chunk.get('nodes', []):
        node_type = node_info.get('type', '')
        content = node_info.get('content', '')
        if node_type in ['table', 'code_block']:
            return True
        if '$$' in content or '$' in content:
            return True
        if '<table>' in content and '</table>' in content:
            return True
    return False


def _split_oversized_chunk(chunk, target_tokens, max_tokens):
    """拆分超大块：在节点边界累积 tokens，超过 target_tokens 时断开，同时维护标题上下文"""
    split_chunks = []
    nodes = chunk.get('nodes', [])
    headers = chunk.get('headers', {})

    current_nodes = []
    current_tokens = 0

    for node_info in nodes:
        node_content = node_info.get('content', '')
        node_tokens = num_tokens_from_string(node_content)
        is_heading = node_info.get('type') == 'heading'

        if current_tokens + node_tokens > target_tokens and current_nodes:
            new_chunk = {
                'headers': headers.copy(),
                'nodes': current_nodes.copy(),
                'chunk_type': 'split_from_oversized',
                'has_special_content': any(
                    _has_special_content({'nodes': [n]}) for n in current_nodes)
            }
            split_chunks.append(new_chunk)

            current_nodes = [node_info]
            current_tokens = node_tokens

            if is_heading:
                level = node_info.get('level', 3)
                title = node_info.get('title', '')
                new_headers = {k: v for k, v in headers.items() if k < level}
                new_headers[level] = title
                headers = new_headers
        else:
            current_nodes.append(node_info)
            current_tokens += node_tokens

            if is_heading:
                level = node_info.get('level', 3)
                title = node_info.get('title', '')
                headers = {k: v for k, v in headers.items() if k < level}
                headers[level] = title

    if current_nodes:
        final_chunk = {
            'headers': headers.copy(),
            'nodes': current_nodes,
            'chunk_type': 'split_from_oversized',
            'has_special_content': any(
                _has_special_content({'nodes': [n]}) for n in current_nodes)
        }
        split_chunks.append(final_chunk)

    return split_chunks


def _try_merge_with_next(current_chunk, all_chunks, current_index, target_tokens):
    """尝试将过小块与下一个块合并，合并后总 tokens <= target_tokens * 1.2 则允许合并"""
    if current_index >= len(all_chunks) - 1:
        return None

    next_chunk = all_chunks[current_index + 1]
    current_content = _render_header_chunk(current_chunk)
    next_content = _render_header_chunk(next_chunk)
    merged_tokens = num_tokens_from_string(current_content + "\n\n" + next_content)

    if merged_tokens <= target_tokens * 1.2:
        return {
            'headers': next_chunk.get('headers', current_chunk.get('headers', {})),
            'nodes': current_chunk.get('nodes', []) + next_chunk.get('nodes', []),
            'chunk_type': 'merged_small',
            'has_special_content': (_has_special_content(current_chunk) or
                                    _has_special_content(next_chunk)),
            'merged_count': 2,
            'source_sections': 2
        }

    return None


def _enhance_small_chunk_with_context(chunk):
    """为无法合并的过小块注入父标题上下文路径，增强语义完整性"""
    enhanced_chunk = chunk.copy()
    enhanced_chunk['chunk_type'] = 'small_enhanced'
    enhanced_chunk['has_special_content'] = _has_special_content(chunk)

    headers = chunk.get('headers', {})
    if headers:
        context_parts = []
        for level in sorted(headers.keys()):
            context_parts.append('#' * level + " " + headers[level])

        if context_parts:
            context_node = {
                'type': 'context',
                'content': '\n'.join(context_parts),
                'headers': headers.copy(),
                'is_split_boundary': False
            }
            enhanced_chunk['nodes'] = [context_node] + enhanced_chunk.get('nodes', [])

    return enhanced_chunk


def _apply_size_control_and_optimization(chunks, min_tokens, target_tokens, max_tokens):
    """高级策略的核心优化循环：遍历初步分块，按 token 大小执行保留/拆分/合并/增强"""
    optimized_chunks = []
    i = 0
    while i < len(chunks):
        chunk = chunks[i]
        chunk_content = _render_header_chunk(chunk)
        chunk_tokens = num_tokens_from_string(chunk_content)
        has_special = _has_special_content(chunk)

        if chunk_tokens <= max_tokens and chunk_tokens >= min_tokens:
            chunk['chunk_type'] = 'normal'
            chunk['has_special_content'] = has_special
            optimized_chunks.append(chunk)

        elif chunk_tokens > max_tokens and not has_special:
            split_chunks = _split_oversized_chunk(chunk, target_tokens, max_tokens)
            optimized_chunks.extend(split_chunks)

        elif chunk_tokens < min_tokens:
            merged_chunk = _try_merge_with_next(chunk, chunks, i, target_tokens)
            if merged_chunk:
                optimized_chunks.append(merged_chunk)
                i += merged_chunk.get('merged_count', 1) - 1
            else:
                enhanced_chunk = _enhance_small_chunk_with_context(chunk)
                optimized_chunks.append(enhanced_chunk)
        else:
            chunk['chunk_type'] = 'oversized_special'
            chunk['has_special_content'] = has_special
            optimized_chunks.append(chunk)

        i += 1

    return optimized_chunks


def _render_header_chunk_advanced(chunk_info):
    """渲染高级策略的 chunk：对拆分/增强类型块首自动补充最近标题上下文"""
    content_parts = []
    chunk_has_header = any(node['type'] == 'heading' for node in chunk_info.get('nodes', []))
    headers = chunk_info.get('headers', {})
    chunk_type = chunk_info.get('chunk_type', 'normal')

    if chunk_type in ['split_from_oversized', 'small_enhanced'] and headers and not chunk_has_header:
        max_level = max(headers.keys())
        context_header = '#' * max_level + " " + headers[max_level]
        if context_header:
            content_parts.append(context_header)

    for node_info in chunk_info.get('nodes', []):
        if node_info.get('content', '').strip():
            content_parts.append(node_info['content'])

    return "\n\n".join(content_parts).strip()


def split_markdown_to_chunks_advanced(txt, chunk_token_num=256, min_chunk_tokens=10,
                                       overlap_ratio=0.0, include_metadata=False):
    """高级分块策略：标题初步分块 + 三轮动态大小优化
    1. 按 H1/H2/H3 边界初步分块
    2. 遍历每个块：合适(50-800 tokens)直接保留，过大(>800)在段落边界二次拆分，
       过小(<50)尝试与下一个块合并，包含表格/代码块的超大块保持完整性不拆
    动态阈值: target_min=50, target=300-600, target_max=800 tokens
    Args:
        txt: markdown 文本
        chunk_token_num: 参考 token 数，实际上下限由内部动态阈值控制
        min_chunk_tokens: 参考最小 token 数
        overlap_ratio: 重叠比例（预留，当前未实际使用）
        include_metadata: 是否附带 metadata
    Returns:
        字符串列表（默认）或 dict 列表（include_metadata=True）
        异常时回退到智能策略
    """
    if not MARKDOWN_IT_AVAILABLE:
        return split_markdown_to_chunks(txt, chunk_token_num)

    if not txt or not txt.strip():
        return []

    target_min_tokens = max(50, min_chunk_tokens // 2)
    target_tokens = min(600, chunk_token_num)
    target_max_tokens = min(800, int(chunk_token_num * 1.5))

    headers_to_split_on = [1, 2, 3]

    md = MarkdownIt("commonmark", {"breaks": True, "html": True})
    md.enable(['table'])

    try:
        tokens = md.parse(txt)
        tree = SyntaxTreeNode(tokens)

        nodes_with_headers = _extract_nodes_with_header_info(tree, headers_to_split_on)
        initial_chunks = _split_by_header_levels(nodes_with_headers, headers_to_split_on)

        optimized_chunks = _apply_size_control_and_optimization(
            initial_chunks, target_min_tokens, target_tokens, target_max_tokens
        )

        final_chunks = []
        for chunk_info in optimized_chunks:
            content = _render_header_chunk_advanced(chunk_info)
            if content.strip():
                if include_metadata:
                    chunk_data = {
                        'content': content,
                        'metadata': chunk_info.get('headers', {}),
                        'token_count': num_tokens_from_string(content),
                        'chunk_type': chunk_info.get('chunk_type', 'header_based'),
                        'has_special_content': chunk_info.get('has_special_content', False),
                        'source_sections': chunk_info.get('source_sections', 1)
                    }
                    final_chunks.append(chunk_data)
                else:
                    final_chunks.append(content)

        return final_chunks

    except Exception:
        return split_markdown_to_chunks_smart(txt, chunk_token_num, min_chunk_tokens)


# ===================== 正则分块（严格正则匹配） =====================

def split_markdown_to_chunks_strict_regex(txt, chunk_token_num=256, min_chunk_tokens=10,
                                            regex_pattern=''):
    """严格正则分块策略：逐行扫描，行首匹配自定义正则时断开前一个 chunk
    完全不关心 token 数，100% 由正则控制切割点。
    Args:
        txt: markdown 文本
        chunk_token_num: token 数（此策略不控制大小）
        min_chunk_tokens: 最小 token（此策略不控制大小）
        regex_pattern: 自定义正则表达式，如 "第[一二三四五六七八九十]+条"
    Returns:
        字符串列表
        正则为空或语法错误时回退到智能策略
    """
    if not txt or not txt.strip():
        return []

    if not regex_pattern or not regex_pattern.strip():
        return split_markdown_to_chunks_smart(txt, chunk_token_num, min_chunk_tokens)

    try:
        precise_pattern = r'^\s*' + regex_pattern
        lines = txt.split('\n')
        chunks = []
        current_chunk = []

        for line in lines:
            if re.search(precise_pattern, line) and current_chunk:
                chunk_content = '\n'.join(current_chunk).strip()
                if chunk_content:
                    chunks.append(chunk_content)
                current_chunk = [line]
            else:
                current_chunk.append(line)

        if current_chunk:
            chunk_content = '\n'.join(current_chunk).strip()
            if chunk_content:
                chunks.append(chunk_content)

        return [chunk for chunk in chunks if chunk.strip()]

    except re.error:
        return split_markdown_to_chunks_smart(txt, chunk_token_num, min_chunk_tokens)
    except Exception:
        return split_markdown_to_chunks_smart(txt, chunk_token_num, min_chunk_tokens)


# ===================== yg_pos 坐标提取与匹配 =====================

YG_POS_PATTERN = re.compile(r'<!--yg_pos(\d+),(\d+),(\d+),(\d+),(\d+)yg_pos-->')


def extract_yg_pos_mapping(content: str):
    """
    从原始内容中提取所有 yg_pos 标记，构建清洗后文本和位置映射。

    Returns:
        clean_content: 去掉所有 yg_pos 标记的清洗文本
        yg_mapping: [(position_in_clean, yg_location_str), ...]
    """
    matches = list(YG_POS_PATTERN.finditer(content))
    if not matches:
        return content, []

    mapping = []
    clean_parts = []
    last_end = 0

    for m in matches:
        clean_parts.append(content[last_end:m.start()])
        position_in_clean = sum(len(p) for p in clean_parts)
        mapping.append((position_in_clean, m.group(0)))
        last_end = m.end()

    clean_parts.append(content[last_end:])
    clean_content = ''.join(clean_parts)

    return clean_content, mapping


def find_yg_location(chunk_text: str, clean_content: str, yg_mapping: list):
    """
    为单个 chunk 找到对应的 yg_location 标记。

    策略：在清洗文本中定位 chunk 内容首次出现的位置，
         然后向前查找最近的 yg_pos 标记。

    Args:
        chunk_text: 分块文本内容
        clean_content: 清洗后的完整文档文本
        yg_mapping: yg_pos 位置映射列表

    Returns:
        yg_location 字符串或 None
    """
    if not yg_mapping:
        return None

    pos = clean_content.find(chunk_text)
    if pos == -1:
        first_line = chunk_text.split('\n')[0].strip()
        if first_line:
            pos = clean_content.find(first_line)

    if pos == -1:
        return None

    for i in range(len(yg_mapping) - 1, -1, -1):
        if yg_mapping[i][0] <= pos:
            return yg_mapping[i][1]

    return yg_mapping[0][1] if yg_mapping else None


def find_yg_locations_for_chunks(chunks: list, clean_content: str, yg_mapping: list):
    """
    为一组分块查找 yg_location，并将 chunk 文本范围内同一页的所有 yg_pos 合并。

    每个 chunk 取其覆盖范围内的所有 yg_pos 标记，按页合并 bbox 后
    生成该 chunk 独享的 yg_location。不同 chunk 在同一页上互不相干。

    Returns:
        yg_locations: [yg_location_str or None, ...] 与 chunks 等长
    """
    if not yg_mapping:
        return [None] * len(chunks)

    yg_locations = []
    last_pos = 0

    for chunk_text in chunks:
        if isinstance(chunk_text, dict):
            chunk_text = chunk_text.get('content', '')
        if not chunk_text or not chunk_text.strip():
            yg_locations.append(None)
            continue

        pos = clean_content.find(chunk_text, last_pos)
        if pos == -1:
            first_line = chunk_text.split('\n')[0].strip()
            if first_line:
                pos = clean_content.find(first_line, last_pos)

        if pos != -1:
            chunk_end = pos + len(chunk_text)

            # 收集 chunk 范围内所有 yg_pos，按页合并 bbox
            page_bboxes = {}
            for mp in yg_mapping:
                if not (pos <= mp[0] <= chunk_end):
                    continue
                m = YG_POS_PATTERN.match(mp[1])
                if not m:
                    continue
                page_idx = int(m.group(1))
                x0, y0, x1, y1 = int(m.group(2)), int(m.group(3)), int(m.group(4)), int(m.group(5))

                if page_idx not in page_bboxes:
                    page_bboxes[page_idx] = [x0, y0, x1, y1]
                else:
                    pb = page_bboxes[page_idx]
                    pb[0] = min(pb[0], x0)
                    pb[1] = min(pb[1], y0)
                    pb[2] = max(pb[2], x1)
                    pb[3] = max(pb[3], y1)

            if page_bboxes:
                # 有多个页面时取第一页（正常情况下一个 chunk 只跨一页）
                page_idx, pb = next(iter(page_bboxes.items()))
                found = f"<!--yg_pos{page_idx},{pb[0]},{pb[1]},{pb[2]},{pb[3]}yg_pos-->"
            else:
                # 回退：取 chunk 结束后第一个 yg_pos
                found = None
                for mp in yg_mapping:
                    if mp[0] >= chunk_end:
                        found = mp[1]
                        break
                if found is None:
                    found = yg_mapping[-1][1]

            yg_locations.append(found)
            last_pos = pos
        else:
            yg_locations.append(None)

    return yg_locations


# ===================== 论文切块 (Paper) =====================

ABSTRACT_KEYWORDS = ['abstract', '摘要', '概要', '总结', 'conclusion']


def _group_nodes_to_sections(nodes_with_headers: list) -> list:
    """将带标题信息的节点按标题分组成段落。

    返回:
        [(heading_level, section_content, is_abstract), ...]
    """
    sections = []
    current_level = 0
    current_parts = []

    for n in nodes_with_headers:
        if n['type'] == 'heading':
            if current_parts:
                content = '\n\n'.join(current_parts).strip()
                sections.append((current_level, content, False))
            current_level = n['level']
            current_parts = [n['content']]
        else:
            current_parts.append(n['content'])

    if current_parts:
        content = '\n\n'.join(current_parts).strip()
        sections.append((current_level, content, False))

    return sections


def _mark_abstract_sections(sections: list) -> list:
    """标记摘要段落（匹配前几个段落中标题含有关键词的）"""
    for i, (level, content, _) in enumerate(sections):
        if i >= 3:  # 只在前 3 个段落中找
            break
        title_lower = content.split('\n')[0].lower()
        for kw in ABSTRACT_KEYWORDS:
            if kw in title_lower:
                sections[i] = (level, content, True)
                break
    return sections


def _get_most_frequent_level(sections: list) -> int:
    """统计非摘要段落的标题级别频率，返回出现最多的级别"""
    level_counts = {}
    for level, _, is_abs in sections:
        if is_abs or level == 0:
            continue
        level_counts[level] = level_counts.get(level, 0) + 1

    if not level_counts:
        return 1

    return max(level_counts, key=level_counts.get)  # type: ignore[arg-type]


def split_markdown_to_chunks_paper(
    content: str,
    chunk_token_num: int = 1024,
    min_chunk_tokens: int = 10,
) -> list:
    """论文智能切块：摘要独立成块，按最频繁标题级别切分，相邻同级别合并。

    Args:
        content: 已清洗的 markdown 文本
        chunk_token_num: 目标 token 数上限，合并时作为参考
        min_chunk_tokens: 最小 token 阈值

    Returns:
        字符串列表
    """
    if not MARKDOWN_IT_AVAILABLE or not content or not content.strip():
        return split_markdown_to_chunks(content, chunk_token_num)

    md = MarkdownIt("commonmark", {"breaks": True, "html": True})
    md.enable(['table'])

    try:
        tokens = md.parse(content)
        tree = SyntaxTreeNode(tokens)

        # 1. 解析为段落
        nodes = _extract_nodes_with_header_info(tree, [])
        sections = _group_nodes_to_sections(nodes)
        if not sections:
            return [content.strip()]

        # 2. 标记摘要
        sections = _mark_abstract_sections(sections)

        # 3. 找到最频繁的标题级别
        most_level = _get_most_frequent_level(sections)

        # 4. 按最频繁级别切分：遇到 <= last_split_level 的标题时切分，子级别吸收
        chunks = []
        current_parts = []
        current_tokens = 0
        last_split_level = 999  # 确保第一个标题必定触

        for level, content, is_abs in sections:
            sec_tokens = num_tokens_from_string(content)

            # 摘要直接独立成块
            if is_abs:
                if current_parts:
                    chunks.append('\n\n'.join(current_parts).strip())
                    current_parts = []
                    current_tokens = 0
                chunks.append(content)
                continue

            # 遇到 <= most_level 且不深于上一个切分标题 → 新 chunk
            if level <= most_level and level <= last_split_level:
                if current_parts and current_tokens >= min_chunk_tokens:
                    chunks.append('\n\n'.join(current_parts).strip())
                    current_parts = [content]
                    current_tokens = sec_tokens
                else:
                    current_parts.append(content)
                    current_tokens += sec_tokens
                last_split_level = level
            else:
                # 子级别内容或同级连续段落 → 吸收到当前 chunk
                current_parts.append(content)
                current_tokens += sec_tokens

        if current_parts:
            chunks.append('\n\n'.join(current_parts).strip())

        # 后处理：合并过小块
        merged = []
        i = 0
        while i < len(chunks):
            cur = chunks[i]
            cur_tokens = num_tokens_from_string(cur)
            if cur_tokens < min_chunk_tokens and i + 1 < len(chunks):
                merged.append(cur + '\n\n' + chunks[i + 1])
                i += 2
            else:
                merged.append(cur)
                i += 1
        chunks = merged

        return [c for c in chunks if c.strip()]

    except Exception:
        return split_markdown_to_chunks_smart(content, chunk_token_num, min_chunk_tokens)


# ===================== 法文切块 (Laws) =====================

LAW_PATTERNS = {
    'part':  re.compile(r'第[一二三四五六七八九十百]+编'),
    'chapter': re.compile(r'第[一二三四五六七八九十百]+章'),
    'section': re.compile(r'第[一二三四五六七八九十百]+节'),
    'article': re.compile(r'第[一二三四五六七八九十百]+条'),
}


def _classify_legal_level(title: str) -> str:
    """根据标题文本识别法文层级类型"""
    for lvl_name, pattern in reversed(LAW_PATTERNS.items()):
        if pattern.search(title):
            return lvl_name
    return 'other'


def _resolve_group_level(sections: list, chunk_token_num: int) -> str:
    """根据 token 数选择合适的聚合级别（节→章→编）"""
    for group_level in ('section', 'chapter', 'part'):
        merged_size = 0
        last_group = None
        for lvl, _, title in sections:
            group = title if lvl == group_level else last_group
            if group != last_group:
                if merged_size > chunk_token_num * 2:
                    return previous_level if previous_level else group_level
                merged_size = 0
                last_group = group
                previous_level = group_level

    return 'part'


def split_markdown_to_chunks_laws(
    content: str,
    chunk_token_num: int = 1024,
    min_chunk_tokens: int = 10,
) -> list:
    """法文智能切块：识别编/章/节/条层级，以条为单位不可拆分，逐级聚合。

    Args:
        content: 已清洗的 markdown 文本
        chunk_token_num: 目标 token 数上限
        min_chunk_tokens: 最小 token 阈值

    Returns:
        字符串列表
    """
    if not MARKDOWN_IT_AVAILABLE or not content or not content.strip():
        return split_markdown_to_chunks(content, chunk_token_num)

    md = MarkdownIt("commonmark", {"breaks": True, "html": True})
    md.enable(['table'])

    try:
        tokens = md.parse(content)
        tree = SyntaxTreeNode(tokens)
        nodes = _extract_nodes_with_header_info(tree, [])

        # 1. 按标题分节，识别法文层级
        sections = []
        current_parts = []
        current_level = 0
        current_title = ''

        for n in nodes:
            if n['type'] == 'heading':
                if current_parts:
                    sections.append((current_level, '\n\n'.join(current_parts).strip(), current_title))
                current_level = n['level']
                current_title = n.get('title', '')
                current_parts = [n['content']]
            else:
                current_parts.append(n['content'])

        if current_parts:
            sections.append((current_level, '\n\n'.join(current_parts).strip(), current_title))

        if not sections:
            return [content.strip()]

        # 2. 确定条级别 = 法文标题的最低层级（或最常见层级）
        legal_levels = [_classify_legal_level(t) for _, _, t in sections]
        article_level = None
        for (lvl, _, t), lt in zip(sections, legal_levels):
            if lt == 'article':
                article_level = lvl
                break

        # 3. 以条为原子单位，按上级标题聚合
        group_level_names = ['part', 'chapter', 'section']
        group_level = None
        if article_level:
            for gl in group_level_names:
                for (lvl, _, t), lt in zip(sections, legal_levels):
                    if lt == gl and lvl < article_level:
                        group_level = gl
                        break
                if group_level:
                    break
        if not group_level:
            group_level = 'section'

        # 4. 按 group_level 切分，条不拆分
        chunks = []
        current_parts = []
        current_tokens = 0
        last_group_title = None

        for (lvl, text, title), lt in zip(sections, legal_levels):
            sec_tokens = num_tokens_from_string(text)

            if lt == group_level:
                if last_group_title is not None and title != last_group_title:
                    # 遇到新的 group_level 标题 → 切分
                    if current_parts:
                        chunks.append('\n\n'.join(current_parts).strip())
                    current_parts = [text]
                    current_tokens = sec_tokens
                else:
                    current_parts.append(text)
                    current_tokens += sec_tokens
                last_group_title = title
            elif lt in group_level_names[:group_level_names.index(group_level)]:
                # 上级标题（如编在章之上）→ 也切分
                if current_parts:
                    chunks.append('\n\n'.join(current_parts).strip())
                current_parts = [text]
                current_tokens = sec_tokens
                last_group_title = title
            else:
                # 条或更下级 → 不切分
                current_parts.append(text)
                current_tokens += sec_tokens

        if current_parts:
            chunks.append('\n\n'.join(current_parts).strip())

        # 5. 合并过小块
        merged = []
        i = 0
        while i < len(chunks):
            cur = chunks[i]
            cur_tokens = num_tokens_from_string(cur)
            if cur_tokens < min_chunk_tokens and i + 1 < len(chunks):
                if cur_tokens + num_tokens_from_string(chunks[i + 1]) <= chunk_token_num * 1.5:
                    merged.append(cur + '\n\n' + chunks[i + 1])
                    i += 2
                    continue
            merged.append(cur)
            i += 1

        return [c for c in merged if c.strip()]

    except Exception:
        return split_markdown_to_chunks_smart(content, chunk_token_num, min_chunk_tokens)


# ===================== 大模型切块 (LLM) =====================

SENTENCE_SPLIT_PATTERN = re.compile(r'(?<=[。！？!?\n])(?=[^\s])')


def _split_into_sentences(text: str) -> list:
    """按句号、感叹号、问号、换行等分隔符将文本预切为句子级原子段"""
    parts = SENTENCE_SPLIT_PATTERN.split(text)
    return [p.strip() for p in parts if p.strip()]


def _build_segment_list(sentences: list, start_idx: int, max_tokens: int) -> tuple:
    """从 start_idx 开始累积句子，直到接近 max_tokens 上限。

    返回:
        (segments_text, end_idx) — segments_text 为带编号的段落列表文本，
        end_idx 为已消费的最后一个句子索引（不含）。
    """
    lines = []
    current_tokens = 0
    end = start_idx

    for i in range(start_idx, len(sentences)):
        line = f"[{i + 1}] {sentences[i]}"
        line_tokens = num_tokens_from_string(line)
        if current_tokens + line_tokens > max_tokens and lines:
            break
        lines.append(line)
        current_tokens += line_tokens
        end = i + 1

    return '\n'.join(lines), end


def _parse_merge_response(response: str, start_idx: int, sentence_count: int) -> list:
    """解析 LLM 返回的合并决策，如 "1-3\n4\n5-7"。

    返回:
        [(abs_start, abs_end), ...] 每组为 chunk 包含的句子索引区间（闭区间）
    """
    chunks = []
    for line in response.strip().split('\n'):
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        m = re.match(r'(\d+)\s*[-–]\s*(\d+)', line)
        if m:
            a, b = int(m.group(1)), int(m.group(2))
        else:
            m = re.match(r'(\d+)', line)
            if not m:
                continue
            a = b = int(m.group(1))

        abs_a = start_idx + a - 1
        abs_b = start_idx + b - 1
        abs_a = max(0, min(abs_a, sentence_count - 1))
        abs_b = max(abs_a, min(abs_b, sentence_count - 1))
        chunks.append((abs_a, abs_b))

    return chunks


async def split_markdown_to_chunks_llm(
    content: str,
    model: str,
    api_key: str,
    base_url: str,
    chunk_token_num: int = 1024,
    min_chunk_tokens: int = 10,
) -> list:
    """大模型语义切块：预分段为句子 → 分批送 LLM 判断合并边界 → 生成 chunks。

    Args:
        content: 已清洗的 markdown 文本
        model/api_key/base_url: LLM 配置
        chunk_token_num: 目标 token 数上限
        min_chunk_tokens: 最小 token 阈值

    Returns:
        字符串列表
    """
    from tools.llm_chat import chat_with_llm
    from corekg_chunk.prompt.prompt import PROMPTS

    if not content or not content.strip():
        return []

    # 1. 按句分割
    sentences = _split_into_sentences(content)
    if len(sentences) <= 1:
        return [content.strip()]

    # 2. 分批送 LLM 判断合并
    batch_max_tokens = int(chunk_token_num * 8)  # 每批约 8 个 chunk 的内容
    raw_chunks = []
    i = 0
    last_context = ''

    while i < len(sentences):
        segments_text, end = _build_segment_list(sentences, i, batch_max_tokens)

        if last_context:
            prompt = PROMPTS["chunk_merge_continue"].format(
                context=last_context,
                segments=segments_text,
            )
        else:
            prompt = PROMPTS["chunk_merge"].format(segments=segments_text)

        try:
            response = await chat_with_llm(
                prompt=prompt,
                model=model,
                api_key=api_key,
                base_url=base_url,
            )
        except Exception:
            # LLM 调用失败，回退：每句独立成块
            raw_chunks.extend(sentences[i:end])
            i = end
            continue

        chunk_ranges = _parse_merge_response(response, i, len(sentences))

        for abs_a, abs_b in chunk_ranges:
            merged = '\n'.join(sentences[abs_a:abs_b + 1])
            raw_chunks.append(merged)

        # 记录最后一个 chunk 的上下文，传给下一批
        if chunk_ranges:
            last_a, last_b = chunk_ranges[-1]
            last_context = sentences[last_a][:200]
        else:
            last_context = ''

        i = end

    # 3. 过小 chunk 合并
    merged_chunks = []
    i = 0
    while i < len(raw_chunks):
        cur = raw_chunks[i]
        cur_tokens = num_tokens_from_string(cur)
        if cur_tokens < min_chunk_tokens and i + 1 < len(raw_chunks):
            merged_chunks.append(cur + '\n' + raw_chunks[i + 1])
            i += 2
        else:
            merged_chunks.append(cur)
            i += 1

    return [c for c in merged_chunks if c.strip()]


# ===================== 全文单块 (Resume) =====================

def split_markdown_to_chunks_resume(content: str) -> list:
    """全文作为一个完整 chunk，清除 yg_pos 标记后返回。

    适用于简历等短文档场景，不进行任何切分。
    """
    if not content or not content.strip():
        return []
    return [YG_POS_PATTERN.sub('', content).strip()]


# ===================== 按页切块 (Slide) =====================

def split_markdown_to_chunks_slide(content: str):
    """按 yg_pos 中的 page_idx 切分，每页一个 chunk。

    返回:
        (chunks, page_indices) — chunks 为每个页面的清洗后文本，
        page_indices 为对应的页码列表
    """
    if not content or not content.strip():
        return [], []

    matches = list(YG_POS_PATTERN.finditer(content))
    if not matches:
        return [content.strip()], [0]

    markers = [(m.start(), int(m.group(1))) for m in matches]

    chunks = []
    pages = []
    current_page = markers[0][1]
    page_start = 0

    for pos, page_idx in markers:
        if page_idx != current_page:
            chunk_text = content[page_start:pos].strip()
            if chunk_text:
                chunks.append(chunk_text)
                pages.append(current_page)
            page_start = pos
            current_page = page_idx

    chunk_text = content[page_start:].strip()
    if chunk_text:
        chunks.append(chunk_text)
        pages.append(current_page)

    cleaned = [YG_POS_PATTERN.sub('', c).strip() for c in chunks]

    # 过滤空 chunk 并同步 pages
    result_chunks, result_pages = [], []
    for c, p in zip(cleaned, pages):
        if c:
            result_chunks.append(c)
            result_pages.append(p)

    return result_chunks, result_pages


def compute_page_bboxes(yg_mapping: list) -> dict:
    """将每页的所有 yg_pos 合并为一个页面级包围盒

    参数:
        yg_mapping: extract_yg_pos_mapping 的返回列表 [(pos, '<!--yg_pos...-->'), ...]

    返回:
        {page_idx: yg_location_str} 字典
    """
    page_bboxes: dict[int, list] = {}

    for _, marker in yg_mapping:
        m = YG_POS_PATTERN.match(marker)
        if not m:
            continue
        page_idx = int(m.group(1))
        x0, y0, x1, y1 = int(m.group(2)), int(m.group(3)), int(m.group(4)), int(m.group(5))

        if page_idx not in page_bboxes:
            page_bboxes[page_idx] = [x0, y0, x1, y1]
        else:
            pb = page_bboxes[page_idx]
            pb[0] = min(pb[0], x0)
            pb[1] = min(pb[1], y0)
            pb[2] = max(pb[2], x1)
            pb[3] = max(pb[3], y1)

    result = {}
    for page_idx, (x0, y0, x1, y1) in page_bboxes.items():
        result[page_idx] = f"<!--yg_pos{page_idx},{x0},{y0},{x1},{y1}yg_pos-->"

    return result


# ===================== 统一适配接口 =====================

def get_knowchunks(content: str,
                         strategy: str = "smart",
                         chunk_token_num: int = 256,
                         min_chunk_tokens: int = 10,
                         split_level: int = 2,
                         overlap_ratio: float = 0.0,
                         regex_pattern: str = "",
                         delimiter: str = "\n!?。；！？",
                         enable_heading_in_content: bool = False):
    """
    统一的 knowchunk 分块适配接口。

    Args:
        content: 要分块的 markdown 文本（slide 模式需包含 yg_pos 标记）
        strategy: 分块策略 (smart/basic/advanced/title/strict_regex/slide/resume)
        ...
    """
    # ---- 1. 按策略分块 ----
    if strategy == "resume":
        raw_chunks = split_markdown_to_chunks_resume(content)
        chunks = raw_chunks
        metas = [{} for _ in raw_chunks]
        return chunks, metas

    elif strategy == "paper":
        raw_chunks = split_markdown_to_chunks_paper(
            content, chunk_token_num, min_chunk_tokens
        )
        metas = [{} for _ in raw_chunks]
        chunks = raw_chunks
        return chunks, metas

    elif strategy == "laws":
        raw_chunks = split_markdown_to_chunks_laws(
            content, chunk_token_num, min_chunk_tokens
        )
        metas = [{} for _ in raw_chunks]
        chunks = raw_chunks
        return chunks, metas

    elif strategy == "slide":
        raw_chunks, _slide_pages = split_markdown_to_chunks_slide(content)
        chunks = raw_chunks
        metas = [{} for _ in raw_chunks]
        return chunks, metas

    elif strategy == "basic":
        raw_chunks = split_markdown_to_chunks(content, chunk_token_num, delimiter)
        metas = [{} for _ in raw_chunks]

    elif strategy == "strict_regex":
        raw_chunks = split_markdown_to_chunks_strict_regex(
            content, chunk_token_num, min_chunk_tokens, regex_pattern
        )

    elif strategy == "title":
        raw_chunks = split_markdown_to_chunks_title(
            content, chunk_token_num, min_chunk_tokens, split_level,
            enable_heading_in_content
        )

    elif strategy == "advanced":
        raw_chunks = split_markdown_to_chunks_advanced(
            content, chunk_token_num, min_chunk_tokens, overlap_ratio
        )

    elif strategy == "smart":
        raw_chunks = split_markdown_to_chunks_smart(
            content, chunk_token_num, min_chunk_tokens, enable_heading_in_content
        )

    else:
        raw_chunks = split_markdown_to_chunks_smart(content, chunk_token_num, min_chunk_tokens)

    # ---- 2. 统一格式化为 (chunks, metas) ----
    chunks = []
    metas = []
    for chunk in raw_chunks:
        if isinstance(chunk, dict):
            chunks.append(chunk.get('content', ''))
            metas.append(chunk.get('heading_metadata', chunk.get('metadata', {})))
        else:
            chunks.append(chunk)
            metas.append({})

    # ---- 3. 应用重叠 ----
    if overlap_ratio > 0 and len(chunks) > 1:
        for i in range(1, len(chunks)):
            prev_len = len(chunks[i - 1])
            if prev_len == 0:
                continue
            overlap_chars = max(1, int(prev_len * overlap_ratio))
            overlap_text = chunks[i - 1][-overlap_chars:]
            if overlap_text.strip():
                chunks[i] = overlap_text.rstrip() + "\n" + chunks[i]

    return chunks, metas
