/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package minio

import (
	"context"
	"strings"
	"time"

	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/workflow/infra/imagex"
	"github.com/insmtx/corekg/apps/workflow/pkg/ctxcache"
	"github.com/insmtx/corekg/apps/workflow/types/consts"
)

func NewStorageImagex(ctx context.Context, endpoint, accessKeyID, secretAccessKey, bucketName, region string, useSSL bool) (imagex.ImageX, error) {
	m, err := getMinioClient(ctx, endpoint, accessKeyID, secretAccessKey, bucketName, region, useSSL)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *minioClient) GetUploadHost(ctx context.Context) string {
	currentHost, ok := ctxcache.Get[string](ctx, consts.HostKeyInCtx)
	if !ok {
		return ""
	}
	return currentHost + consts.ApplyUploadActionURI
}

func (m *minioClient) GetServerID() string {
	return ""
}

func (m *minioClient) GetUploadAuth(ctx context.Context, opt ...imagex.UploadAuthOpt) (*imagex.SecurityToken, error) {
	return m.GetUploadAuthWithExpire(ctx, time.Hour, opt...)
}

func (m *minioClient) GetUploadAuthWithExpire(ctx context.Context, expire time.Duration, opt ...imagex.UploadAuthOpt) (*imagex.SecurityToken, error) {
	storageCfg := conf.GetAppConfig().Workflow.Storage
	scheme := strings.ToLower(storageCfg.UploadHTTPScheme)
	if scheme == "" {
		scheme = "http"
	}

	stsEndpoint := strings.TrimSpace(storageCfg.MinIO.APIHost)
	if stsEndpoint == "" {
		stsEndpoint = m.endpoint
	}
	if !strings.Contains(stsEndpoint, "://") {
		stsEndpoint = scheme + "://" + stsEndpoint
	}

	expireSeconds := int(expire.Seconds())
	if expireSeconds < 0 {
		expireSeconds = 0
	}

	stsCreds, err := credentials.NewSTSAssumeRole(stsEndpoint, credentials.STSAssumeRoleOptions{
		AccessKey:       m.accessKeyID,
		SecretKey:       m.secretAccessKey,
		DurationSeconds: expireSeconds,
		Location:        m.region,
	})
	if err != nil {
		return nil, err
	}

	value, err := stsCreds.GetWithContext(nil)
	if err != nil {

		return nil, err
	}

	loc := time.FixedZone("CST", 8*60*60)
	now := time.Now().In(loc)
	expTime := value.Expiration
	if expTime.IsZero() {
		expTime = now.Add(expire)
	}
	expTime = expTime.In(loc)
	return &imagex.SecurityToken{
		AccessKeyID:     value.AccessKeyID,
		SecretAccessKey: value.SecretAccessKey,
		SessionToken:    value.SessionToken,
		ExpiredTime:     expTime.Format("2006-01-02 15:04:05"),
		CurrentTime:     now.Format("2006-01-02 15:04:05"),
		HostScheme:      scheme,
	}, nil
}

func (m *minioClient) GetResourceURL(ctx context.Context, uri string, opts ...imagex.GetResourceOpt) (*imagex.ResourceURL, error) {
	url, err := m.GetObjectUrl(ctx, uri)
	if err != nil {
		return nil, err
	}
	return &imagex.ResourceURL{
		URL: url,
	}, nil
}

func (m *minioClient) Upload(ctx context.Context, data []byte, opts ...imagex.UploadAuthOpt) (*imagex.UploadResult, error) {
	return nil, nil
}
