import json
import shutil
from typing import List, Dict, Any, Optional
from pathlib import Path
from loguru import logger


class ContentListToMarkdown:
    """将 content_list.json 转换为 Markdown 格式"""
    
    # 内容类型常量
    CONTENT_TYPE_TEXT = 'text'
    CONTENT_TYPE_IMAGE = 'image'
    CONTENT_TYPE_TABLE = 'table'
    CONTENT_TYPE_CHART = 'chart'
    CONTENT_TYPE_EQUATION = 'equation'
    CONTENT_TYPE_CODE = 'code'
    CONTENT_TYPE_LIST = 'list'
    CONTENT_TYPE_SEAL = 'seal'
    
    # 页面辅助类型：屏蔽以下内容不在md展示
    PAGE_TYPES = {
        # 'header', 'footer', 'page_number', 'aside_text', 'page_footnote'
    }
    
    # LaTeX 分隔符
    INLINE_MATH_DELIM = '$'
    DISPLAY_MATH_DELIM = '$$'
    
    def __init__(self, image_dir: str = 'images'):
        """初始化转换器"""
        self.image_dir = image_dir
        self.markdown_lines = []
    
    def convert(self, content_list: List[Dict[str, Any]]) -> str:
        """
        将 content_list 转换为 Markdown
        
        Args:
            content_list: content_list.json 的内容（列表格式）
            
        Returns:
            完整的 Markdown 文本
        """
        self.markdown_lines = []
        
        for item in content_list:
            if not item:
                continue
            
            try:
                markdown = self._convert_item(item)
                
                if markdown and markdown.strip():
                    self.markdown_lines.append(markdown)
            except Exception as e:
                logger.warning(f"⚠️ 转换失败: {e}, 项目: {item.get('type')}")
                continue
        
        return '\n\n'.join(self.markdown_lines)
    
    def _convert_item(self, item: Dict[str, Any]) -> str:
        """转换单个内容项"""
        item_type = item.get('type')
        
        # 文本/标题
        if item_type == self.CONTENT_TYPE_TEXT:
            return self._convert_text(item)
        
        # 图片
        elif item_type == self.CONTENT_TYPE_IMAGE:
            return self._convert_image(item)
        
        # 表格
        elif item_type == self.CONTENT_TYPE_TABLE:
            return self._convert_table(item)
        
        # 图表
        elif item_type == self.CONTENT_TYPE_CHART:
            return self._convert_chart(item)
        
        # 方程式
        elif item_type == self.CONTENT_TYPE_EQUATION:
            return self._convert_equation(item)
        
        # 代码
        elif item_type == self.CONTENT_TYPE_CODE:
            return self._convert_code(item)
        
        # 列表
        elif item_type == self.CONTENT_TYPE_LIST:
            return self._convert_list(item)
        
        # 印章
        elif item_type == self.CONTENT_TYPE_SEAL:
            return self._convert_seal(item)
        
        # 页面辅助块
        elif item_type in self.PAGE_TYPES:
            return self._convert_page_element(item, item_type)
        
        return ''
    
    def _get_first_list_item(self, lst: Any, default: str = '') -> str:
        """
        安全地获取列表的第一个元素
        
        Args:
            lst: 列表或其他类型
            default: 默认值
            
        Returns:
            第一个元素或默认值
        """
        if isinstance(lst, list) and len(lst) > 0:
            item = lst[0]
            return str(item).strip() if item else default
        return default
    
    def _get_list_as_strings(self, lst: Any) -> List[str]:
        """
        将列表转换为字符串列表，处理各种情况
        
        Args:
            lst: 列表或其他类型
            
        Returns:
            字符串列表
        """
        if not lst:
            return []
        
        if not isinstance(lst, list):
            return [str(lst).strip()]
        
        result = []
        for item in lst:
            if item and str(item).strip():
                result.append(str(item).strip())
        
        return result
    
    def _convert_text(self, item: Dict[str, Any]) -> str:
        """转换文本项"""
        text = item.get('text', '').strip()
        if not text:
            return ''
        
        # 检查是否有标题级别
        text_level = item.get('text_level', 0)
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        
        # 生成位置标记
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        if text_level > 0:
            # 标题
            return f"{'#' * text_level} {text}{pos_marker}"
        else:
            # 普通文本
            return f"{text}{pos_marker}"
    
    def _convert_image(self, item: Dict[str, Any]) -> str:
        """转换图片项"""
        parts = []
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # 图片本体
        img_path = item.get('img_path', '')
        if img_path:
            img_url = self._build_image_url(img_path)
            # 使用第一个标题，如果没有则使用空字符串
            caption = self._get_first_list_item(item.get('image_caption'), '')
            parts.append(f"![{caption}]({img_url}){pos_marker}")
        
        # 图片标题
        captions = self._get_list_as_strings(item.get('image_caption', []))
        for cap in captions:
            if cap:
                parts.append(cap)
        
        # 图片脚注
        footnotes = self._get_list_as_strings(item.get('image_footnote', []))
        for note in footnotes:
            if note:
                parts.append(f"> {note}")
        
        return '\n'.join(parts)
    
    def _convert_table(self, item: Dict[str, Any]) -> str:
        """转换表格项"""
        parts = []
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # 表格标题
        captions = self._get_list_as_strings(item.get('table_caption', []))
        for cap in captions:
            if cap:
                parts.append(f"**{cap}**")
        
        # 表格 HTML
        table_html = item.get('table_body', '')
        
        # 兼容不同的字段名
        if not table_html:
            # 尝试使用 'table_body' 作为 key（可能是直接的 HTML）
            for key in item.keys():
                if 'body' in key.lower() and isinstance(item[key], str):
                    table_html = item[key]
                    break
        
        if table_html:
            # 如果有 HTML，直接使用
            parts.append(f"{table_html}{pos_marker}")
        else:
            # 否则尝试使用图片
            img_path = item.get('img_path', '')
            if img_path:
                img_url = self._build_image_url(img_path)
                parts.append(f"![Table]({img_url}){pos_marker}")
        
        # 表格脚注
        footnotes = self._get_list_as_strings(item.get('table_footnote', []))
        for note in footnotes:
            if note:
                parts.append(f"> {note}")
        
        return '\n'.join(parts)
    
    def _convert_chart(self, item: Dict[str, Any]) -> str:
        """转换图表项"""
        parts = []
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # 图表标题
        captions = self._get_list_as_strings(item.get('chart_caption', []))
        for cap in captions:
            if cap:
                parts.append(f"**{cap}**")
        
        # 图表图片
        img_path = item.get('img_path', '')
        if img_path:
            img_url = self._build_image_url(img_path)
            parts.append(f"![Chart]({img_url}){pos_marker}")
        
        # 图表内容（Markdown 表格文本）
        content = item.get('content', '')
        if content and str(content).strip():
            parts.append(str(content).strip())
        
        # 图表脚注
        footnotes = self._get_list_as_strings(item.get('chart_footnote', []))
        for note in footnotes:
            if note:
                parts.append(f"> {note}")
        
        return '\n'.join(parts)
    
    def _convert_equation(self, item: Dict[str, Any]) -> str:
        """转换方程式项"""
        parts = []
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # LaTeX 公式
        text = item.get('text', '')
        if text:
            text = str(text).strip()
            # 如果已经有 LaTeX 分隔符，直接使用
            if text.startswith(self.DISPLAY_MATH_DELIM):
                parts.append(f"{text}{pos_marker}")
            else:
                parts.append(f"{self.DISPLAY_MATH_DELIM}\n{text}\n{self.DISPLAY_MATH_DELIM}{pos_marker}")
        else:
            # 使用图片
            img_path = item.get('img_path', '')
            if img_path:
                img_url = self._build_image_url(img_path)
                parts.append(f"![Equation]({img_url}){pos_marker}")
        
        return '\n'.join(parts)
    
    def _convert_code(self, item: Dict[str, Any]) -> str:
        """转换代码项"""
        parts = []
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # 代码标题
        sub_type = item.get('sub_type', 'code')
        captions = self._get_list_as_strings(item.get('code_caption', []))
        
        for cap in captions:
            if cap:
                if sub_type == 'algorithm':
                    parts.append(f"**算法：{cap}**")
                else:
                    parts.append(f"**{cap}**")
        
        # 代码体
        code_body = item.get('code_body', '')
        if code_body:
            code_body = str(code_body).strip()
            language = 'python'  # 默认语言
            if sub_type == 'algorithm':
                language = 'text'
            
            parts.append(f"```{language}\n{code_body}\n```{pos_marker}")
        
        # 代码脚注
        footnotes = self._get_list_as_strings(item.get('code_footnote', []))
        for note in footnotes:
            if note:
                parts.append(f"> {note}")
        
        return '\n'.join(parts)
    
    def _convert_list(self, item: Dict[str, Any]) -> str:
        """转换列表项"""
        list_items = item.get('list_items', [])
        sub_type = item.get('sub_type', '')
        
        if not list_items:
            return ''
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        parts = []
        
        for list_item in list_items:
            if not list_item:
                continue
            
            list_item_str = str(list_item).strip()
            if list_item_str:
                if sub_type == 'ref_text':
                    # 参考文献列表
                    parts.append(f"- {list_item_str}")
                else:
                    # 普通列表
                    parts.append(f"- {list_item_str}")
        
        # 在整个列表结束后添加位置标记
        result = '\n'.join(parts)
        if result:
            result += pos_marker
        
        return result
    
    def _convert_seal(self, item: Dict[str, Any]) -> str:
        """转换印章项"""
        parts = []
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # 印章图片
        img_path = item.get('img_path', '')
        if img_path:
            img_url = self._build_image_url(img_path)
            parts.append(f"![Seal]({img_url}){pos_marker}")
        
        # 印章文本
        text = item.get('text', '')
        if text:
            text = str(text).strip()
            if text:
                parts.append(f"*{text}*")
        
        return '\n'.join(parts)
    
    def _convert_page_element(self, item: Dict[str, Any], element_type: str) -> str:
        """转换页面元素"""
        text = item.get('text', '')
        if not text:
            return ''
        
        text = str(text).strip()
        if not text:
            return ''
        
        # 获取位置信息
        page_idx = item.get('page_idx', 0)
        bbox = item.get('bbox', [0, 0, 0, 0])
        scaled_bbox = bbox
        pos_marker = f" \n<!--yg_pos{page_idx+1},{scaled_bbox[0]},{scaled_bbox[1]},{scaled_bbox[2]},{scaled_bbox[3]}yg_pos-->"
        
        # 根据类型添加前缀（可选）
        if element_type == 'header':
            return f"**[Header]** {text}{pos_marker}"
        elif element_type == 'footer':
            return f"**[Footer]** {text}{pos_marker}"
        elif element_type == 'page_number':
            return f"*Page: {text}*{pos_marker}"
        elif element_type == 'aside_text':
            return f"> {text}{pos_marker}"
        elif element_type == 'page_footnote':
            return f"^[{text}]{pos_marker}"
        
        return f"{text}{pos_marker}"
    
    def _build_image_url(self, img_path: str) -> str:
        """构建图片 URL"""
        if not img_path:
            return ''
        
        # 规范化路径
        img_path = str(img_path).strip()
        
        # 如果路径已经包含目录或是绝对路径，直接使用
        if '/' in img_path or '\\' in img_path or img_path.startswith('.'):
            return img_path
        
        # 否则添加图片目录前缀
        return f"{self.image_dir}/{img_path}"


def convert_content_list_to_markdown(
    content_list_json_path: str,
    output_md_path: Optional[str] = None,
    image_dir: str = 'images',
    target_images_dir: str = 'images',
) -> str:
    """
    从 content_list.json 文件转换为 Markdown

    Args:
        content_list_json_path: content_list.json 文件路径
        output_md_path: 输出 Markdown 文件路径（可选）
        image_dir: 图片目录
        target_images_dir: 目标图片目录
    Returns:
        生成的 Markdown 内容
    """
    # 读取 JSON 文件
    try:
        with open(content_list_json_path, 'r', encoding='utf-8') as f:
            content_list = json.load(f)
    except FileNotFoundError:
        logger.error(f"❌ 文件未找到: {content_list_json_path}")
        return ''
    except json.JSONDecodeError as e:
        logger.error(f"❌ JSON 解析错误: {e}")
        return ''
    
    # 确保 content_list 是列表
    if not isinstance(content_list, list):
        logger.warning(f"⚠️ 警告: content_list 不是列表，而是 {type(content_list)}")
        if isinstance(content_list, dict):
            # 如果是字典，尝试提取列表
            for key, value in content_list.items():
                if isinstance(value, list):
                    content_list = value
                    break
    
    # 转换
    converter = ContentListToMarkdown(image_dir=image_dir)
    markdown_content = converter.convert(content_list)
    
    # 保存（如果指定了输出路径）
    if output_md_path:
        output_path = Path(output_md_path)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        try:
            with open(output_md_path, 'w', encoding='utf-8') as f:
                f.write(markdown_content)
            
            logger.info(f"✅ Markdown 文件已保存到: {output_md_path}")
        except Exception as e:
            logger.error(f"❌ 保存文件失败: {e}")
    
        # 获取 images 目录的绝对路径
        source_images_dir = Path(image_dir).resolve()

        if source_images_dir.exists():
            try:                
                # 递归复制整个 images 目录到 md 同级目录下
                shutil.copytree(source_images_dir, target_images_dir)
                logger.info(f"✅ 图片目录已成功复制到: {target_images_dir}")
            except Exception as e:
                logger.error(f"❌ 复制图片目录失败: {e}")
        else:
            logger.warning(f"⚠️ 警告: 原图片目录不存在，跳过复制: {source_images_dir}")
    
    return markdown_content
