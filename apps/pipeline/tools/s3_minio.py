import os
import mimetypes
import boto3
import requests
import yaml
from botocore.client import Config
from tools.temp_file import TempFileManager
import certifi

class MinIoS3:
    """S3上传工具类 测试环境配置"""

    def __init__(self):
        # 读取 YAML 配置文件
        with open(os.getenv('COREKG_CONFIGPATH', './config/analyser_config.yaml'), "r") as f:
            config = yaml.safe_load(f)

        self.s3_config = config["s3"]

        # 初始化 S3 客户端
        self.s3_client = boto3.client(
            "s3",
            endpoint_url=self.s3_config["endpoint_url"],
            aws_access_key_id=self.s3_config["access_key_id"],
            aws_secret_access_key=self.s3_config["secret_access_key"],
            region_name=self.s3_config["region"],
            config=Config(
                s3={"addressing_style": "path"},  # 强制路径式访问
                signature_version="s3v4",  # 确保签名兼容
                connect_timeout=30,  # 可选：调整超时
                retries={"max_attempts": 3}  # 可选：重试策略
            ),
            verify=False  # 如果证书不受信任，关闭 SSL 验证
        )

    
    def download_file(self, download_url, save_dir, filename=None):
        """下载文件到指定目录"""
        # 确保保存目录存在
        os.makedirs(save_dir, exist_ok=True)

        # 获取文件名
        if not filename:
            filename = download_url.split('/')[-1].split('?')[0]  # 去除URL参数

        # 完整的保存路径
        save_path = os.path.join(save_dir, filename)

        # 发送请求下载文件
        response = requests.get(download_url, stream=True,verify=certifi.where())
        response.raise_for_status()  # 检查请求是否成功

        # 写入文件
        with open(save_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                if chunk:  # 过滤掉保持连接的新块
                    f.write(chunk)
        msg = f"{save_path} 下载成功"
        return save_path, msg

    def upload_file(self, local_file_path: str, s3_key: str, bucket: str=None):
        # 设置文件的 Content-Type
        content_type = self.get_content_type(local_file_path)

        try:
            # 上传文件到 S3
            self.s3_client.upload_file(
                Filename=local_file_path,
                Bucket=bucket,
                Key=s3_key,
                ExtraArgs={'ContentType': content_type}
            )

            # 生成你想要的 URL 格式
            file_url = f"""{self.s3_config["endpoint_url"]}/{bucket}/{s3_key}"""
            log_msg = f"上传成功: {local_file_path} → {s3_key}"
            return file_url, log_msg
        except Exception as e:
            log_msg = f"上传失败: {local_file_path} → {s3_key}, \n{str(e)}"
            raise Exception(log_msg)

    def upload_directory(self, local_dir: str, s3_base_path: str, bucket: str = None):
        """上传整个目录到 S3，保持相同目录结构"""
        # 检查路径是否存在
        if not os.path.isdir(local_dir):
            log_msg = f"该路径不存在: {local_dir}"
            raise ValueError(log_msg)

        # 检查目录中是否有文件
        has_files = any(
            os.path.isfile(os.path.join(root, file))
            for root, _, files in os.walk(local_dir)
            for file in files
        )

        if not has_files:
            log_msg = f"目录中没有文件: {local_dir}"
            raise ValueError(log_msg)

        uploaded_urls = []
        log_msgs = []
        bucket = bucket

        # 遍历本地目录
        for root, _, files in os.walk(local_dir):
            for file in files:
                # 本地文件完整路径（如 `/data/project/docs/readme.md`）
                local_file_path = os.path.join(root, file)

                # 计算相对于 local_dir 的相对路径（如 `docs/readme.md`）
                relative_path = os.path.relpath(local_file_path, local_dir)

                # 构建 S3 Key（如 `backups/project/docs/readme.md`）
                s3_key = os.path.join(s3_base_path, relative_path).replace("\\", "/")
                s3_key= s3_key.replace("//", '/')

                # 设置文件的 Content-Type
                content_type = self.get_content_type(file)

                try:
                    # 上传文件到 S3
                    self.s3_client.upload_file(
                        Filename=local_file_path,
                        Bucket=bucket,
                        Key=s3_key,
                        ExtraArgs={'ContentType': content_type}
                    )

                    # 生成你想要的 URL 格式
                    file_url = f"""{self.s3_config["endpoint_url"]}/{bucket}/{s3_key}"""
                    uploaded_urls.append(file_url)
                    log_msgs.append(f"上传成功: {local_file_path} → s3://{bucket}/{s3_key}")
                except Exception as e:
                    log_msgs.append(f"上传失败: {local_file_path} → s3://{bucket}/{s3_key}, \n{str(e)}")
                    raise e

        # 将所有日志消息合并为一个字符串，用换行符分隔
        combined_log = "\n".join(log_msgs)
        return uploaded_urls, combined_log

    def get_endpoint(self):
        return self.s3_config["endpoint_url"]


    def get_content_type(self,file_path):
        """通过文件扩展名获取Content-Type"""
        # 添加md文件的MIME类型映射
        mimetypes.add_type('text/markdown', '.md')
        content_type, encoding = mimetypes.guess_type(file_path)
        return content_type


def download_file(url, save_path):
    """
    下载文件到指定位置

    参数:
        url (str): 文件的URL地址
        save_path (str): 文件保存的完整路径（包括文件名）

    返回:
        bool: 下载成功返回True，失败返回False
    """
    try:
        # 发送HTTP GET请求
        response = requests.get(url, stream=True)
        response.raise_for_status()  # 检查请求是否成功

        # 确保保存目录存在
        os.makedirs(os.path.dirname(save_path), exist_ok=True)

        # 以二进制写入模式打开文件
        with open(save_path, 'wb') as file:
            for chunk in response.iter_content(chunk_size=8192):
                if chunk:  # 过滤掉保持连接的块
                    file.write(chunk)

        print(f"文件已成功下载到: {save_path}")
        return True
    except requests.exceptions.RequestException as e:
        print(f"下载文件时出错: {e}")
        return False
    except IOError as e:
        print(f"写入文件时出错: {e}")
        return False




