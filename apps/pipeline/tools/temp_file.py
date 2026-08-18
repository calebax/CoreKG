import os
import shutil
from pathlib import Path
import logging


class TempFileManager:
    """临时文件与目录管理器，负责创建和清理临时目录"""

    def __init__(self, base_temp_dir: str = "/tmp/temp_data/download", task_id: int = 00):
        """初始化临时目录路径"""
        self.base_temp_dir = f"{base_temp_dir}/{task_id}"
        self.logger = logging.getLogger(__name__+f" ({task_id})")
    def create_temp_dir(self) -> Path:
        """创建基础临时目录，不存在则自动创建"""
        if not os.path.exists(self.base_temp_dir):
            self.logger.info(f"Creating temp directory: {self.base_temp_dir}")
            os.makedirs(self.base_temp_dir, exist_ok=True)
        return Path(self.base_temp_dir)

    def get_temp_path(self, file_id: str) -> str:
        """根据文件 ID 生成完整的临时文件路径"""
        temp_dir = self.create_temp_dir()
        return str(temp_dir / file_id)
    @staticmethod
    def remove_file(file_path: str):
        """删除本地源文件，避免误上传到对象存储"""
        try:
            if os.path.isfile(file_path):
                os.remove(file_path)
                return f"文件已删除: {file_path}"
            return None
        except Exception as e:
            return f"文件删除失败 {file_path}: {str(e)}"

    def clear(self):
        """
        删除目录
        """
        shutil.rmtree(self.base_temp_dir)
