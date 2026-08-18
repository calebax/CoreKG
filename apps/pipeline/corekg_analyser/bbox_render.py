"""
bbox 坐标可视化渲染模块

读取 content_list.json 中已转换的 bbox 像素坐标（300 DPI），
绘制到原始 PDF 页面上，生成带颜色标注的可视化 PDF。

颜色映射:
    蓝色 - text    红色 - image   绿色 - table
    橙色 - chart   紫色 - equation   灰色 - 其他
"""

import json
from pathlib import Path
from loguru import logger
from PIL import ImageDraw, ImageFont

TYPE_COLORS = {
    'text':     '#4D88FF',
    'image':    '#FF4D4D',
    'table':    '#1CB34B',
    'chart':    '#FF8C00',
    'equation': '#9B4DFF',
    'code':     '#B3B300',
    'list':     '#00CCCC',
    'seal':     '#FF00FF',
}
DEFAULT_COLOR = '#888888'


def _get_font(size: int = 12):
    """获取可用字体"""
    try:
        return ImageFont.truetype('arial.ttf', size)
    except Exception:
        try:
            return ImageFont.truetype('DejaVuSans.ttf', size)
        except Exception:
            return ImageFont.load_default()


def render_bbox_to_pdf(
    content_list_json_path: str,
    pdf_path: str,
    output_pdf_path: str = None,
    render_dpi: int = 300,
):
    """
    读取 content_list.json 中的像素坐标 bbox，绘制到 PDF 页面上。

    bbox 坐标应已在上游（work_pdf.py）完成 0-1000 → 像素转换。

    参数:
        content_list_json_path: content_list.json 文件路径
        pdf_path: 原始 PDF 文件路径
        output_pdf_path: 输出 PDF 路径，默认文件名加 _bbox 后缀
        render_dpi: 渲染分辨率，需与 bbox 坐标 DPI 一致，默认 300

    返回:
        输出文件路径
    """
    from pdf2image import convert_from_path

    if output_pdf_path is None:
        in_path = Path(pdf_path)
        output_pdf_path = str(in_path.parent / f"{in_path.stem}_bbox.pdf")

    with open(content_list_json_path, 'r', encoding='utf-8') as f:
        content_list = json.load(f)

    logger.info(f"正在渲染 PDF 页面，DPI={render_dpi} ...")
    images = convert_from_path(pdf_path, dpi=render_dpi)

    page_items = {}
    for item in content_list:
        if not item or not isinstance(item, dict):
            continue
        bbox = item.get('bbox', [])
        if not bbox or len(bbox) < 4:
            continue
        page_idx = item.get('page_idx', 0)
        if page_idx < 0 or page_idx >= len(images):
            continue
        page_items.setdefault(page_idx, []).append(item)

    font = _get_font(10)
    rendered = 0

    for page_idx, img in enumerate(images):
        draw = ImageDraw.Draw(img)

        for item in page_items.get(page_idx, []):
            bbox = item.get('bbox', [])[:4]
            item_type = item.get('type', '')
            color = TYPE_COLORS.get(item_type, DEFAULT_COLOR)

            x0, y0, x1, y1 = bbox

            draw.rectangle((x0, y0, x1, y1), outline=color, width=2)
            label = item_type or '?'
            tx, ty = x0 + 3, y0 + 3
            text_bbox = draw.textbbox((tx, ty), label, font=font)
            draw.rectangle(text_bbox, fill=color)
            draw.text((tx, ty), label, fill='white', font=font)
            rendered += 1

    if rendered == 0:
        logger.warning("未找到可渲染的 bbox")

    images[0].save(
        output_pdf_path,
        'PDF',
        save_all=True,
        append_images=images[1:],
        resolution=render_dpi,
    )
    logger.info(
        f"bbox 渲染完成: {rendered} 个框, "
        f"{len(images)} 页, 输出: {Path(output_pdf_path).name}"
    )
    return output_pdf_path
