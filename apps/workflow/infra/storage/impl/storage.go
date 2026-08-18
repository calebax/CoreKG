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

package impl

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/insmtx/corekg/apps/workflow/infra/imagex"
	"github.com/insmtx/corekg/apps/workflow/infra/storage"
	"github.com/insmtx/corekg/apps/workflow/infra/storage/impl/minio"
	"github.com/insmtx/corekg/apps/workflow/infra/storage/impl/s3"
	"github.com/insmtx/corekg/apps/workflow/infra/storage/impl/tos"
)

type Storage = storage.Storage

func New(ctx context.Context) (Storage, error) {
	appCfg := conf.GetAppConfig()
	storageConf := appCfg.Workflow.Storage
	useSSL := storageConf.UploadHTTPScheme == "https"
	switch storageConf.Type {
	case "minio":
		return minio.New(
			ctx,
			storageConf.MinIO.Endpoint,
			storageConf.MinIO.AK,
			storageConf.MinIO.SK,
			storageConf.Bucket,
			storageConf.MinIO.Region,
			useSSL,
		)
	case "tos":
		return tos.New(
			ctx,
			storageConf.TOS.AccessKey,
			storageConf.TOS.SecretKey,
			storageConf.Bucket,
			storageConf.TOS.Endpoint,
			storageConf.TOS.Region,
		)
	case "s3":
		return s3.New(
			ctx,
			storageConf.S3.AccessKey,
			storageConf.S3.SecretKey,
			storageConf.Bucket,
			storageConf.S3.Endpoint,
			storageConf.S3.Region,
		)
	}

	return nil, fmt.Errorf("unknown storage type: %s", storageConf.Type)
}

func NewImagex(ctx context.Context) (imagex.ImageX, error) {
	appCfg := conf.GetAppConfig()
	storageConf := appCfg.Workflow.Storage
	useSSL := storageConf.UploadHTTPScheme == "https"
	switch storageConf.Type {
	case "minio":
		return minio.NewStorageImagex(
			ctx,
			storageConf.MinIO.Endpoint,
			storageConf.MinIO.AK,
			storageConf.MinIO.SK,
			storageConf.Bucket,
			storageConf.MinIO.Region,
			useSSL,
		)
	case "tos":
		return tos.NewStorageImagex(
			ctx,
			storageConf.TOS.AccessKey,
			storageConf.TOS.SecretKey,
			storageConf.Bucket,
			storageConf.TOS.Endpoint,
			storageConf.TOS.Region,
		)
	case "s3":
		return s3.NewStorageImagex(
			ctx,
			storageConf.S3.AccessKey,
			storageConf.S3.SecretKey,
			storageConf.Bucket,
			storageConf.S3.Endpoint,
			storageConf.S3.Region,
		)
	}
	return nil, fmt.Errorf("unknown storage type: %s", storageConf.Type)
}
