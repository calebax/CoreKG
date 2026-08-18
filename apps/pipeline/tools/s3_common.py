from tools.s3_minio import MinIoS3
from tools.s3_weedfs import WeedFsS3


class S3():
    def __new__(self, s3_type: str='minio',):
        """
        :params: s3_type = "minio/weedfs", 通过参数确认使用哪个s3，默认是minio
        """
        if s3_type.lower() == 'minio':
            return MinIoS3()
        else:
            return WeedFsS3()
