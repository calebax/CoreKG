import json
import csv
from corekg_analyser.video.video_process import video_process
from loguru import logger

def json_to_markdown(json_data):
    """递归将 JSON 数据转换为 Markdown 层级格式"""
    if isinstance(json_data, dict):
        md_content = "## JSON Data\n\n"
        for key, value in json_data.items():
            md_content += f"### {key}\n\n"
            md_content += json_to_markdown(value) + "\n"
    elif isinstance(json_data, list):
        md_content = ""
        for item in json_data:
            md_content += json_to_markdown(item) + "\n"
    else:
        md_content = f"{json_data}\n"
    return md_content

def csv_to_markdown(rows):
    """将 CSV 二维数组转换为 Markdown 表格"""
    md_content = ""
    if rows:
        # 添加表头
        headers = rows[0]
        md_content += "| " + " | ".join(headers) + " |\n"
        md_content += "| " + " | ".join(["---"] * len(headers)) + " |\n"
        # 添加数据行
        for row in rows[1:]:
            md_content += "| " + " | ".join(row) + " |\n"
    return md_content
            

def others_process(pdf_file_name, final_md_file, public_path, file_ext, *args, **kwargs):
    """非 PDF 文件的格式分发处理，根据 file_ext 路由到对应的转换器"""
    if file_ext in ('.txt', '.md'):
        # 打开日志文件并读取内容
        with open(pdf_file_name, 'r', encoding='utf-8') as txt_file:
            txt_content = txt_file.read()
        # 将日志内容转换为Markdown格式
        md_content = f"{txt_content}"
        # 将Markdown内容写入新的文件
        with open(final_md_file, 'w', encoding='utf-8') as md_file:
            md_file.write(md_content)
        logger.info(f"{final_md_file} 已写入完成！")

    elif file_ext == '.log':
        # 打开日志文件并读取内容
        with open(pdf_file_name, 'r', encoding='utf-8') as log_file:
            log_content = log_file.read()
        # 将日志内容转换为Markdown格式
        md_content = f"{log_content}"
        # 将Markdown内容写入新的文件
        with open(final_md_file, 'w', encoding='utf-8') as md_file:
            md_file.write(md_content)
        logger.info(f"{final_md_file} 已写入完成！")

    elif file_ext == '.json':
        with open(pdf_file_name, 'r', encoding='utf-8') as json_file:
            json_data = json.load(json_file)
        # 将JSON数据转换为Markdown格式
        md_content = json_to_markdown(json_data)
        # 将Markdown内容写入新的文件
        with open(final_md_file, 'w', encoding='utf-8') as md_file:
            md_file.write(md_content)
        logger.info(f"{final_md_file} 已写入完成！")

    elif file_ext == '.csv':
        # 打开CSV文件并读取内容
        with open(pdf_file_name, 'r', encoding='utf-8') as csv_file:
            reader = csv.reader(csv_file)
            rows = list(reader)
        # 将CSV数据转换为Markdown表格格式
        md_content = csv_to_markdown(rows)
        # 将Markdown内容写入新的文件
        with open(final_md_file, 'w', encoding='utf-8') as md_file:
            md_file.write(md_content)
        logger.info(f"{final_md_file} 已写入完成！")
        
    else:
        video_process(video_path=pdf_file_name, final_md_file=final_md_file, image_prefix=public_path, *args, **kwargs)
        