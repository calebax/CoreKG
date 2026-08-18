package minio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
)

// CreateBucket 创建桶，创建公开读私有写的配置
func CreateBucket(ctx context.Context) error {
	var cfg config.StorageConfig
	err := settings.GetYaml("core", "cos-ke", &cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "get storage config error: %v", err)
		return err
	}
	s3cfg, err := s3config.LoadDefaultConfig(ctx,
		s3config.WithRegion(cfg.S3.Region),
		s3config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, "")),
	)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(s3cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3.EndPoint) // 直接指定 endpoint URL
		o.UsePathStyle = cfg.S3.UsePathStyle         // 使用路径风格的URL
	})
	// 创建 bucket
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(cfg.S3.Bucket),
	})
	if err != nil {
		// 如果 bucket 已存在，AWS 返回 BucketAlreadyExists 或 BucketAlreadyOwnedByYou 错误
		var bucketExists *types.BucketAlreadyExists
		var bucketOwned *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bucketExists) || errors.As(err, &bucketOwned) {
			// 桶已经存在了
			// return nil
		} else {
			logs.ErrorContextf(ctx, "Create bucket error: %v", err)
			return err
		}

	}
	err = CreateKeepFile(ctx, client, cfg.S3.Bucket)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateKeepFile error: %v", err)
		return err
	}
	_, err = client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Policy: aws.String(fmt.Sprintf(`
	{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::%s/*"
            ]
        }
    ]
}
	`, cfg.S3.Bucket)),
	})
	return err
}

// CreateCozeBucket 创建coze桶，创建公开读私有写的配置
func CreateCozeBucket(ctx context.Context) error {
	var cfg config.StorageConfig
	err := settings.GetYaml("core", "cos-ke", &cfg)
	if err != nil {
		logs.ErrorContext(ctx, "get storage config error: %v", err)
		return err
	}
	cfg.S3.Bucket = "opencoze"
	s3cfg, err := s3config.LoadDefaultConfig(ctx,
		s3config.WithRegion(cfg.S3.Region),
		s3config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, "")),
	)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(s3cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3.EndPoint) // 直接指定 endpoint URL
		o.UsePathStyle = cfg.S3.UsePathStyle         // 使用路径风格的URL
	})
	// 创建 bucket
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("opencoze"),
	})
	if err != nil {
		// 如果 bucket 已存在，AWS 返回 BucketAlreadyExists 或 BucketAlreadyOwnedByYou 错误
		var bucketExists *types.BucketAlreadyExists
		var bucketOwned *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bucketExists) || errors.As(err, &bucketOwned) {
			// 桶已经存在了
			// return nil
		} else {
			logs.ErrorContextf(ctx, "Create bucket error: %v", err)
			return err
		}

	}
	err = CreateKeepFile(ctx, client, "opencoze")
	if err != nil {
		logs.ErrorContextf(ctx, "CreateKeepFile error: %v", err)
		return err
	}
	_, err = client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String("opencoze"),
		Policy: aws.String(`
	{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::opencoze/*"
            ]
        }
    ]
}
	`),
	})

	return UploadCozeFile(ctx, cfg)
}

func CreateKeepFile(ctx context.Context, client *s3.Client, bucket string) error {
	content := []byte("this is a test")

	uploader := manager.NewUploader(client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("keep/keep.md"),
		Body:        bytes.NewBuffer(content),
		ContentType: aws.String(mime.TypeByExtension(".md")),
	})
	if err != nil {
		return err
	}
	return nil
}

// UploadCozeFile 上传coze文件到opencoze桶
func UploadCozeFile(ctx context.Context, cfg config.StorageConfig) error {
	logs.InfoContextf(ctx, "UploadCozeFile start")
	st, err := storage.NewStorageWithCfg(cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "file storage init error: %v")
		return err
	}
	_, err = st.UploadDirectory("./scripts/minio", "")
	if err != nil {
		logs.ErrorContextf(ctx, "UploadCozeFile UploadDirectory error: %v")
		return err
	}
	logs.InfoContextf(ctx, "UploadCozeFile success")
	return nil
}
