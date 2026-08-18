SET NAMES utf8mb4;

ALTER TABLE `core_upload_files`
ADD COLUMN `upload_chunk_size` INT DEFAULT 0 COMMENT '分片大小（字节）',
ADD COLUMN `upload_chunk_total` INT DEFAULT 0 COMMENT '分片总数',
MODIFY COLUMN `status` VARCHAR(32) NOT NULL DEFAULT 'normal' COMMENT '文件状态，init：初始化，uploading：上传中，normal：已完成，aborted：已取消，failed：上传失败',
ADD COLUMN `uploaded_chunks` JSON DEFAULT NULL COMMENT '已上传分片列表，例如 [{"partNumber":1,"etag":"xxx"}]',
ADD COLUMN `progress` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '上传进度（%）',
ADD COLUMN `exists` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否命中秒传',
ADD COLUMN `upload_s3_id` VARCHAR(128) DEFAULT '' COMMENT 'S3 MultipartUpload ID',
ADD COLUMN `renew_count` INT NOT NULL DEFAULT 0 COMMENT '预签名 URL 续签次数',
ADD COLUMN `abort_at` DATETIME DEFAULT NULL COMMENT '用户取消上传时间',
ADD COLUMN `completed_at` DATETIME DEFAULT NULL COMMENT '文件上传完成时间',
ADD COLUMN `extra` JSON DEFAULT NULL COMMENT '通用扩展属性，存储自定义元数据，额外业务信息';