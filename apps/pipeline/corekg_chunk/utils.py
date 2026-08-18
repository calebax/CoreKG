import os
import re
import base64
import requests
from loguru import logger
from hashlib import md5
from typing import Any
import tiktoken
import numpy as np
from io import BytesIO
from PIL import Image

def image_url_base64(image_url: str, width: int = None, height: int = None):
    """
    下载图片并转换为 Base64，支持指定宽度和高度。
    修复了 unknown file extension 错误。
    """
    try:
        # 1. 下载图片字节
        resp = requests.get(image_url, timeout=5)
        resp.raise_for_status()
        
        # 2. 打开图片对象
        img = Image.open(BytesIO(resp.content))
        
        # 记录原始格式，用于后续决定保存格式和 MIME 类型
        original_format = img.format 
        if not original_format:
            # 尝试根据文件扩展名推断（如果URL包含扩展名），否则默认为 PNG
            # 但通常 PIL 能识别 content，这里主要防止 format 为 None
            original_format = 'PNG' 

        # 3. 根据参数调整大小
        if width is not None or height is not None:
            original_w, original_h = img.size
            
            if width is not None and height is not None:
                # 情况A: 同时指定了宽和高 -> 强制拉伸 (或者 crop + resize)
                img = img.resize((width, height), Image.Resampling.LANCZOS)
            elif width is not None:
                # 情况B: 只指定宽 -> 按比例缩放高
                scale = width / original_w
                new_h = int(original_h * scale)
                img = img.resize((width, new_h), Image.Resampling.LANCZOS)
            else:
                # 情况C: 只指定高 -> 按比例缩放宽
                scale = height / original_h
                new_w = int(original_w * scale)
                img = img.resize((new_w, height), Image.Resampling.LANCZOS)
        
        # 4. 转换回 BytesIO 以便编码
        buffered = BytesIO()

        # 如果原图是 RGBA/LA/P (带透明度)，必须保存为 PNG 或 WebP，不能存为 JPEG
        # 如果原图是 RGB，且用户没指定格式，我们可以根据原格式保存，或者统一转为 JPEG (体积更小)
        # 为了兼容性，建议：
        # - 有透明度 -> PNG
        # - 无透明度 -> 保持原格式 (如果是 JPEG 则存 JPEG，其他转 JPEG 以减小体积)
        
        save_format = original_format
        
        if img.mode in ('RGBA', 'LA', 'P'):
            save_format = 'PNG'
        else:
            # 如果没有透明度，优先保持原格式，但如果原格式不是常见的 (JPEG, PNG, WEBP)，则转为 JPEG
            if save_format not in ['JPEG', 'PNG', 'WEBP']:
                save_format = 'JPEG'
        
        # 针对 JPEG 的特殊处理：如果是 RGBA 转 JPEG 会报错，所以必须先转 mode='RGB'
        if save_format == 'JPEG' and img.mode in ('RGBA', 'LA', 'P'):
             # 这种情况上面已经通过 save_format='PNG' 处理了，这里主要是防御性编程
             pass 
        
        # 保存
        # 注意：如果原图是 JPEG 但有透明度（极少见），或者需要压缩质量
        img.save(buffered, format=save_format, quality=85)
        
        img_bytes = buffered.getvalue()
        
        # 5. 确定 MIME Type
        mime_map = {
            'JPEG': 'image/jpeg',
            'JPG': 'image/jpeg',
            'PNG': 'image/png',
            'WEBP': 'image/webp',
            'GIF': 'image/gif'
        }
        mime_type = mime_map.get(save_format, 'image/png') # 默认 fallback
        
        # 6. 转 base64
        img_base64 = base64.b64encode(img_bytes).decode("utf-8")
        
        return img_base64, mime_type
        
    except Exception as e:
        raise RuntimeError(f"图片处理失败: {str(e)}")
    
class TiktokenFunc:
    def __init__(self, model_name="gpt-4o-mini"):
        self.model_name = model_name
        self.tokenizer = tiktoken.encoding_for_model(self.model_name)

    def encode(self, content):
        return self.tokenizer.encode(content)

    def decode(self, tokens):
        return self.tokenizer.decode(tokens)
    
# 从url读取文本，尝试多解码方式防止乱码
def get_content(file_url):
    """从 URL 读取文本内容，自动尝试多种编码防止乱码"""

    encodings = ['utf-8-sig', 'utf-8', 'gbk', 'gb2312', 'big5', 'latin1']

    try:
        response = requests.get(file_url, timeout=10)
        
        if response.status_code != 200:
            return None

        raw_data = response.content

        # 逐个尝试编码解码
        for encoding in encodings:
            try:
                content = raw_data.decode(encoding, errors='strict')
                return content
            except UnicodeDecodeError:
                continue

        # 所有尝试失败，强制用 latin1 + replace 保证可读
        return raw_data.decode('latin1', errors='replace')

    except requests.exceptions.RequestException as e:
        raise

# 清洗掉全部的非图片url
def url_clean(content):
    pattern = r'(?<!!\[\])[a-zA-Z]ttps?://[^\s)]+(?![^\(]*\))'
    content = re.sub(pattern, '', content, flags=re.DOTALL)
    return content

# 清洗掉全部email
def email_clean(content):
    pattern = r'\b[\w\.-]+@[\w\.-]+\.\w{2,}\b'
    content = re.sub(pattern, '', content)
    return content

# 删除文章中全部的无意义空字符，返回str        
def others_clean(content):
    content = _clean_spacess(content)
    content = _clean_t(content)
    return content

# others_clean内部方法(连续空格清洗)
def _clean_spacess(content):
    pattern = r' {2,}'
    content = re.sub(pattern, ' ', content)
    return content

# others_clean内部方法(连续制表符清洗)
def _clean_t(content):
    pattern = r'\t+'
    content = re.sub(pattern, '\t', content)
    return content

# 文本的初步清洗
def simple_clean(content: str,
                 remove_URL: bool,
                 remove_email: bool,
                 remove_empty_line: bool,
                 ):
    logger.info("**********************开始文本清洗**********************")
    
    if remove_URL:
        content = url_clean(content)
        logger.info('url已清洗')
        
    if remove_email:
        content = email_clean(content)
        logger.info('email已清洗')

    if remove_empty_line:
        content = others_clean(content)
        logger.info('连续空格，空字符已清洗')

    return content

def safe_unicode_decode(content):
    # Regular expression to find all Unicode escape sequences of the form \uXXXX
    unicode_escape_pattern = re.compile(r"\\u([0-9a-fA-F]{4})")

    # Function to replace the Unicode escape with the actual character
    def replace_unicode_escape(match):
        # Convert the matched hexadecimal value into the actual Unicode character
        return chr(int(match.group(1), 16))

    # Perform the substitution
    decoded_content = unicode_escape_pattern.sub(
        replace_unicode_escape, content.decode("utf-8")
    )

    return decoded_content

def compute_mdhash_id(content: str, prefix: str = "") -> str:
    """
    Compute a unique ID for a given content string.

    The ID is a combination of the given prefix and the MD5 hash of the content string.
    """
    return prefix + md5(content.encode()).hexdigest()

def clean_str(text: str) -> str:
    return text.strip().replace('"', "").replace("'","").replace('|','').replace('\n','').replace('=','').replace('>','').replace('<','').replace('&gt;','')

def clean_text(text: str) -> str:
    return text.strip().replace("\x00", " ").replace('\n','').replace('\\n','')

def pack_user_ass_to_openai_messages(*args: str):
    roles = ["user", "assistant"]
    return [
        {"role": roles[i % 2], "content": content} for i, content in enumerate(args)
    ]


