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

	"github.com/insmtx/corekg/apps/workflow/api/model/admin/config"
	"github.com/insmtx/corekg/apps/workflow/infra/document/searchstore"
	"github.com/insmtx/corekg/apps/workflow/infra/document/searchstore/impl/elasticsearch"
	embImpl "github.com/insmtx/corekg/apps/workflow/infra/embedding/impl"
	"github.com/insmtx/corekg/apps/workflow/infra/es/impl/es"
)

type Manager = searchstore.Manager

func New(ctx context.Context, conf *config.KnowledgeConfig, esClient es.Client) ([]Manager, error) {
	esSearchstoreManager := elasticsearch.NewManager(&elasticsearch.ManagerConfig{Client: esClient})

	emb, err := embImpl.GetEmbedding(ctx, conf.EmbeddingConfig)
	if err != nil {
		return nil, fmt.Errorf("init embedding for es vector store failed, err=%w", err)
	}

	esVectorManager := elasticsearch.NewManager(&elasticsearch.ManagerConfig{
		Client:     esClient,
		VectorMode: true,
		Embedding:  emb,
	})

	return []searchstore.Manager{esSearchstoreManager, esVectorManager}, nil
}
