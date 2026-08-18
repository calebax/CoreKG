import os
import io
import re

from loguru import logger
import mimetypes
import zipfile

import requests
from pathlib import Path

import yaml

# ===================== 加载配置 =====================
with open("./config/analyser_config.yaml", "r", encoding="utf-8") as file:
    config = yaml.safe_load(file)

def process_pdf_task_api(local_file_path, output_target_path):
    # 1. 创建输出目录
    os.makedirs(output_target_path, exist_ok=True)
    
    # 2. 构建请求参数
    if not local_file_path or not os.path.exists(local_file_path):
        raise FileNotFoundError(f"本地文件未找到: {local_file_path}")

    # 从本地路径中提取真实的文件名和后缀
    filename = os.path.basename(local_file_path)
    # 自动猜测文件的 MIME Type，如果猜不到则默认为 'application/octet-stream'（二进制流）
    mime_type = mimetypes.guess_type(local_file_path)[0] or 'application/octet-stream'

    # 准备 multipart/form-data 的字段
    files = {
        # 使用动态获取的文件名和 MIME Type
        'files': (filename, open(local_file_path, 'rb'), mime_type),
        'return_md': (None, 'true'),
        'response_format_zip': (None, 'true'),
        'return_original_file': (None, 'true'),
        'return_images': (None, 'true'),
        'return_content_list': (None, 'true'),
        'return_middle_json': (None, 'true'),
    }

    # 3. 发送请求
    url = config.get("analyser_api_url", None)
    
    headers = {
        "Accept": "*/*",
        "Connection": "keep-alive"
    }

    try:
        logger.info(f"开始上传文件: {local_file_path} 到 {url}")
        response = requests.post(url, headers=headers, files=files, stream=True)
        response.raise_for_status()

        # 4. 处理返回的 ZIP 流
        # 将响应内容读取为字节流
        zip_bytes = io.BytesIO(response.content)
        
        # 5. 解压 ZIP 文件
        with zipfile.ZipFile(zip_bytes, 'r') as zip_ref:
            # 解压所有文件到 output_target_path
            zip_ref.extractall(output_target_path)
            logger.info(f"文件解压成功，路径: {output_target_path}")

        return True

    except Exception as e:
        logger.error(f"任务执行失败: {str(e)}")
        return False
    finally:
        # 关闭文件句柄
        files['files'][1].close()

 # 智能查找目标文件路径
def find_output_files(output_target_path):
    root_dir = Path(output_target_path)

    # 初始化结果
    result = {
        "md_path": None,
        "json_path": None,
        "images_path": None,
        "origin_path": None
    }

    # A. 查找 Markdown 文件 (.md)
    md_files = list(root_dir.rglob("*.md"))
    if md_files:
        result["md_path"] = str(md_files[0])

    # B. 查找 Content List JSON 文件
    json_files = list(root_dir.rglob("*content_list.json"))
    if not json_files:
        json_files = list(root_dir.rglob("*.json"))
    if json_files:
        result["json_path"] = str(json_files[0])

    # C. 查找 Images 文件夹
    image_dirs = [d for d in root_dir.rglob("images") if d.is_dir()]
    if image_dirs:
        images_path = str(image_dirs[0])
    else:
        images_path = f"{output_target_path}/images"
        os.makedirs(images_path, exist_ok=True)
    result["images_path"] = images_path

    # D. 查找 PDF 原始文件
    pdf_files = list(root_dir.rglob("*.pdf"))
    if pdf_files:
        result["origin_path"] = str(pdf_files[0])
    else:
        result["origin_path"] = None
        logger.warning(f"未找到 PDF 文件")

    logger.info(f"解析完成: {result}")
    return result

def add_prefix_and_save(md_file_path, pub_path):
    """
    读取Markdown文件，为本地图片路径添加前缀，并直接写回源文件。
    
    :param md_file_path: Markdown文件的绝对或相对路径 (例如: "docs/note.md")
    :param pub_path: 需要添加的前缀路径 (例如: "https://cdn.example.com/" 或 "/assets/")
    """
    # 1. 读取源文件内容
    try:
        with open(md_file_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        logger.error(f"❌ 读取文件失败: {md_file_path}, 错误: {e}")
        return

    # 2. 定义正则替换逻辑
    pattern = r'!\[(.*?)\]\((.*?)\)'

    def replace_func(match):
        alt_text = match.group(1)
        img_path = match.group(2).strip()
        
        # 跳过已经是网络绝对路径的图片
        if img_path.startswith(('http://', 'https://')):
            return match.group(0)
            
        # 处理以 / 开头的绝对路径，避免产生双斜杠
        if img_path.startswith('/'):
            img_path = img_path[1:]
            
        # 构造带有前缀的新路径并统一使用正斜杠
        new_img_path = f"{pub_path}/{img_path}"
        new_img_path = new_img_path.replace("\\", "/")
        
        return f'![{alt_text}]({new_img_path})'

    # 3. 执行替换
    new_content = re.sub(pattern, replace_func, content)

    # 4. 将处理后的内容写回源文件
    # 只有当内容有变化时才进行写入操作
    if new_content != content:
        try:
            with open(md_file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            logger.info(f"✅ 已成功更新源文件: {md_file_path}")
        except Exception as e:
            logger.error(f"❌ 写入文件失败: {md_file_path}, 错误: {e}")
    else:
        logger.info(f"⏩ 文件无需修改: {md_file_path}")