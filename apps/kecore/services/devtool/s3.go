package devtool

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

// s3FsInternal 用于通过 unsafe 访问 S3Fs 的私有字段
// 注意：字段顺序和类型必须与真实结构体完全一致
type s3FsInternal struct {
	opt     config.StorageOption
	s3fsCfg config.S3StorageConfig
	client  *s3.Client
	ctx     context.Context
}

// GetS3Internal 从 fs.Forest 中解析 internal 结构体
func GetS3Internal() (*s3FsInternal, error) {
	if fs.Forest == nil {
		return nil, errors.New("storage not initialized")
	}

	// 取出接口内部数据
	iface := (*[2]uintptr)(unsafe.Pointer(&fs.Forest))
	if iface[0] == 0 || iface[1] == 0 {
		return nil, errors.New("invalid storage interface")
	}

	internal := (*s3FsInternal)(unsafe.Pointer(iface[1]))
	if internal == nil {
		return nil, errors.New("internal storage is nil")
	}
	if internal.client == nil {
		return nil, errors.New("s3 client is nil")
	}
	if internal.s3fsCfg.Bucket == "" {
		return nil, errors.New("bucket name is empty")
	}

	return internal, nil
}

// CountFiles 统计 S3 指定 prefix 下（支持所有子目录）的文件数量
func CountFiles(ctx context.Context, prefix, fileExt string) (int, error) {
	internal, err := GetS3Internal()
	if err != nil {
		return 0, err
	}

	// 自动补 /，防止漏掉子目录
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	bucket := internal.s3fsCfg.Bucket

	paginator := s3.NewListObjectsV2Paginator(internal.client, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix), // 会自动列出所有子目录
		MaxKeys: aws.Int32(1000),
	})

	count := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "s3 list objects error: %v", err)
			return 0, err
		}

		for _, obj := range page.Contents {
			// S3 不区分目录，prefix 直接匹配所有层级
			if strings.ToLower(filepath.Ext(aws.ToString(obj.Key))) == fileExt {
				count++
			}
		}
	}

	return count, nil
}

// ProcessFilesWithModification 基于 prefix 和 fileExt 查找文件，下载到本地，修改后重新上传，最后删除本地文件
// 逐条处理以节省内存
// prefix: S3 路径前缀
// fileExt: 文件扩展名（如 ".txt", ".json"），会自动转换为小写匹配
// localPath: 本地临时目录路径，用于存放下载的文件
// modifier: 文件修改函数，接收本地文件路径，返回修改后的文件路径（如果相同则返回原路径）
// args: 修改器参数
// 返回处理成功的文件数量和错误
func ProcessFilesWithModification(ctx context.Context, prefix, fileExt, localPath string, modifier FileModifier, args ModifierArgs) (int, error) {
	internal, err := GetS3Internal()
	if err != nil {
		return 0, err
	}

	// 确保本地目录存在
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return 0, fmt.Errorf("create local directory failed: %w", err)
	}

	// 自动补 /，防止漏掉子目录
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}

	bucket := internal.s3fsCfg.Bucket

	// 创建下载器和上传器，使用流式处理以节省内存
	downloader := manager.NewDownloader(internal.client)
	uploader := manager.NewUploader(internal.client)

	paginator := s3.NewListObjectsV2Paginator(internal.client, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(8000),
	})

	successCount := 0
	processedKeys := make([]string, 0)

	// 逐页处理
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "s3 list objects error: %v", err)
			// 清理已下载的文件
			cleanupLocalFiles(ctx, localPath, processedKeys)
			return successCount, err
		}

		// 逐条处理文件
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)

			// 跳过目录（S3 中目录以 / 结尾）
			if strings.HasSuffix(key, "/") {
				continue
			}

			// 检查文件扩展名
			if strings.ToLower(filepath.Ext(key)) != fileExt {
				continue
			}

			// 处理单个文件
			if err := processSingleFile(ctx, internal, downloader, uploader, bucket, key, localPath, modifier, args); err != nil {
				logs.ErrorContextf(ctx, "process file %s error: %v", key, err)
				// 继续处理下一个文件，不中断整个流程
				continue
			}

			successCount++
			processedKeys = append(processedKeys, key)
		}
	}

	return successCount, nil
}

// processSingleFile 处理单个文件：下载、修改、上传、删除本地文件
func processSingleFile(ctx context.Context, internal *s3FsInternal, downloader *manager.Downloader, uploader *manager.Uploader, bucket, key, localPath string, modifier FileModifier, args ModifierArgs) error {
	logs.InfoContextf(ctx, "[processSingleFile] start, remote file: %s", key)
	// 生成本地文件路径，使用 key 的 MD5 确保唯一性，避免文件名冲突
	fileName := filepath.Base(key)
	// 使用 key 的 MD5 作为文件名的一部分，确保唯一性
	hash := md5.Sum([]byte(key))
	hashStr := fmt.Sprintf("%x", hash)[:8]
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)
	localFilePath := filepath.Join(localPath, fmt.Sprintf("%s_%s%s", nameWithoutExt, hashStr, ext))

	// 1. 下载文件到本地
	file, err := os.Create(localFilePath)
	if err != nil {
		return fmt.Errorf("create local file failed: %w", err)
	}
	defer file.Close()

	_, err = downloader.Download(ctx, file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		os.Remove(localFilePath) // 清理失败的文件
		return fmt.Errorf("download file failed: %w", err)
	}
	file.Close() // 关闭文件以便后续修改

	// 2. 读取文件内容
	content, err := os.ReadFile(localFilePath)
	if err != nil {
		os.Remove(localFilePath) // 清理本地文件
		return fmt.Errorf("read file failed: %w", err)
	}

	// 3. 修改文件
	newContent, modifiedPath, modified, err := modifier(ctx, localFilePath, content, args)
	if err != nil {
		os.Remove(localFilePath) // 清理本地文件
		return fmt.Errorf("modify file failed: %w", err)
	}

	// 如果文件内容没有被修改，跳过上传
	if !modified {
		logs.InfoContextf(ctx, "[processSingleFile] skipping upload file, remote file: %s", key)
		os.Remove(localFilePath) // 清理本地文件
		return nil
	}

	// 4. 将修改后的内容写回文件
	if err := os.WriteFile(localFilePath, newContent, 0644); err != nil {
		os.Remove(localFilePath) // 清理本地文件
		return fmt.Errorf("write modified file failed: %w", err)
	}

	// 如果修改函数返回了不同的路径，使用新路径；否则使用原路径
	uploadFilePath := modifiedPath
	if uploadFilePath == "" {
		uploadFilePath = localFilePath
	}

	// 5. 上传修改后的文件
	uploadFile, err := os.Open(uploadFilePath)
	if err != nil {
		// 清理本地文件
		if uploadFilePath != localFilePath {
			os.Remove(localFilePath)
		}
		os.Remove(uploadFilePath)
		return fmt.Errorf("open modified file failed: %w", err)
	}
	defer uploadFile.Close()

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   uploadFile,
	})
	if err != nil {
		// 清理本地文件
		if uploadFilePath != localFilePath {
			os.Remove(localFilePath)
		}
		os.Remove(uploadFilePath)
		return fmt.Errorf("upload file failed: %w", err)
	}
	uploadFile.Close()

	logs.InfoContextf(ctx, "[processSingleFile] success, remote file: %s", key)

	// 6. 删除本地文件
	if uploadFilePath != localFilePath {
		if err := os.Remove(localFilePath); err != nil {
			logs.WarnContextf(ctx, "remove original local file %s failed: %v", localFilePath, err)
		}
	}
	if err := os.Remove(uploadFilePath); err != nil {
		logs.WarnContextf(ctx, "remove modified local file %s failed: %v", uploadFilePath, err)
	}

	return nil
}

// cleanupLocalFiles 清理本地临时文件
func cleanupLocalFiles(ctx context.Context, localPath string, keys []string) {
	for _, key := range keys {
		fileName := filepath.Base(key)
		localFilePath := filepath.Join(localPath, fileName)
		if err := os.Remove(localFilePath); err != nil && !os.IsNotExist(err) {
			logs.WarnContextf(ctx, "cleanup local file %s failed: %v", localFilePath, err)
		}
	}
}

// ProcessFilesWithModificationStream 基于 prefix 和 fileExt 查找文件，使用流式处理下载、修改、上传
// 这是更节省内存的版本，使用 io.Pipe 进行流式处理
// prefix: S3 路径前缀
// fileExt: 文件扩展名（如 ".txt", ".json"），会自动转换为小写匹配
// modifier: 流式修改函数，接收原始内容流，返回修改后的内容流
// 返回处理成功的文件数量和错误
func ProcessFilesWithModificationStream(ctx context.Context, prefix, fileExt string, modifier func(reader io.Reader) (io.Reader, error)) (int, error) {
	internal, err := GetS3Internal()
	if err != nil {
		return 0, err
	}

	// 自动补 /，防止漏掉子目录
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// 统一转换为小写进行匹配
	fileExt = strings.ToLower(fileExt)
	if !strings.HasPrefix(fileExt, ".") {
		fileExt = "." + fileExt
	}

	bucket := internal.s3fsCfg.Bucket

	// 创建上传器
	uploader := manager.NewUploader(internal.client)

	paginator := s3.NewListObjectsV2Paginator(internal.client, &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1000),
	})

	successCount := 0

	// 逐页处理
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "s3 list objects error: %v", err)
			return successCount, err
		}

		// 逐条处理文件
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)

			// 跳过目录
			if strings.HasSuffix(key, "/") {
				continue
			}

			// 检查文件扩展名
			if strings.ToLower(filepath.Ext(key)) != fileExt {
				continue
			}

			// 流式处理单个文件
			if err := processSingleFileStream(ctx, internal, uploader, bucket, key, modifier); err != nil {
				logs.ErrorContextf(ctx, "process file %s error: %v", key, err)
				// 继续处理下一个文件
				continue
			}

			successCount++
		}
	}

	return successCount, nil
}

// processSingleFileStream 流式处理单个文件：下载、修改、上传（不保存到本地）
func processSingleFileStream(ctx context.Context, internal *s3FsInternal, uploader *manager.Uploader, bucket, key string, modifier func(reader io.Reader) (io.Reader, error)) error {
	// 1. 下载文件（获取流）
	getObjOutput, err := internal.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("download file failed: %w", err)
	}
	defer getObjOutput.Body.Close()

	// 2. 修改文件内容（流式处理）
	modifiedReader, err := modifier(getObjOutput.Body)
	if err != nil {
		return fmt.Errorf("modify file failed: %w", err)
	}

	// 3. 上传修改后的文件
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   modifiedReader,
	})
	if err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}

	return nil
}
