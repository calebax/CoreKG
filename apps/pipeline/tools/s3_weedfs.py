import os
import mimetypes
import boto3
import requests
import yaml
from botocore.client import Config


class WeedFsS3:
    """S3上传工具类 测试环境配置"""

    def __init__(self):
        # 读取 YAML 配置文件
        with open(os.getenv('COREKG_CONFIGPATH', './config/analyser_config.yaml'), "r") as f:
            config = yaml.safe_load(f)

        self.s3_config = config["WEEDFS"]
        self.bucket = self.s3_config["BUCKET"]
        if self.s3_config['public_endpoint_url']:
            self.public_endpoint_url = self.s3_config['public_endpoint_url']
        else:
            self.public_endpoint_url = self.s3_config['ENDPOINT_URL']

        # 初始化 S3 客户端
        self.s3_client = boto3.client(
            "s3",
            endpoint_url=self.s3_config["ENDPOINT_URL"],
            aws_access_key_id=self.s3_config["ACCESS_KEY_ID"],
            aws_secret_access_key=self.s3_config["SECRET_ACCESS_KEY"],
            region_name=self.s3_config["REGION"],
            config=Config(
                s3={"addressing_style": "path"},  # 强制路径式访问
                signature_version="s3v4",  # 确保签名兼容
                connect_timeout=30,  # 可选：调整超时
                retries={"max_attempts": 3}  # 可选：重试策略
            ),
            verify=False  # 如果证书不受信任，关闭 SSL 验证
        )


    def download_file(self,download_url, save_dir, filename=None):
        """
        下载文件到指定目录

        :param url: 文件下载URL
        :param save_dir: 保存目录
        :param filename: 保存文件名(可选)，如果不指定则使用URL中的文件名
        :return: 保存的文件路径
        """
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
        msg=f"{save_path} download success !!!"
        return save_path, msg

    def upload_file(self, local_file_path: str, s3_key: str, bucket: str=None):
        # 设置文件的 Content-Type
        content_type = self.get_content_type(local_file_path)
        if not bucket:
            bucket = self.s3_config["BUCKET"]
        file_url = f"""{self.s3_config["ENDPOINT_URL"]}/{bucket}/{s3_key}"""

        try:
            # 上传文件到 S3
            self.s3_client.upload_file(
                Filename=local_file_path,
                Bucket=bucket,
                Key=s3_key,
                ExtraArgs={'ContentType': content_type}
            )

            # 生成你想要的 URL 格式
            log_msg = f"上传成功: {local_file_path} → {file_url}"
            return file_url, log_msg
        except Exception as e:
            log_msg = f"上传失败: {local_file_path} → {file_url}, \n{str(e)}"
            raise Exception(log_msg)

    def upload_directory(self, local_dir: str, s3_base_path: str, bucket: str = None):
        """上传整个目录（包括子目录）到 S3，并保持相同的目录结构
        Args:
            local_dir: 本地目录路径（如 `/data/project`）
            s3_base_path: S3 目标路径前缀（如 `backups/project`）取file_id 作为桶路径
            bucket: 存储桶名称（可选，默认使用初始化时的 bucket）
        Returns:
            所有上传文件的 S3 访问 URL 列表和日志消息
        """

        if not bucket:
            bucket = self.s3_config["BUCKET"]
        # Check if path exists
        if not os.path.isdir(local_dir):
            log_msg = f"该路径不存在: {local_dir}"
            raise ValueError(log_msg)

        # Check if directory contains any files
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
                    file_url = f"""{self.s3_config["ENDPOINT_URL"]}/{bucket}/{s3_key}"""
                    uploaded_urls.append(file_url)
                    log_msgs.append(f"上传成功: {local_file_path} → s3://{bucket}/{s3_key}")
                except Exception as e:
                    log_msgs.append(f"上传失败: {local_file_path} → s3://{bucket}/{s3_key}, \n{str(e)}")
                    raise e

        # 将所有日志消息合并为一个字符串，用换行符分隔
        combined_log = "\n".join(log_msgs)
        return uploaded_urls, combined_log

    def get_endpoint(self):
        return self.s3_config["ENDPOINT_URL"]


    def get_content_type(self,file_path):
        """通过文件扩展名获取Content-Type"""
        # 添加md文件的MIME类型映射
        mimetypes.add_type('text/markdown', '.md')
        content_type, encoding = mimetypes.guess_type(file_path)
        return content_type