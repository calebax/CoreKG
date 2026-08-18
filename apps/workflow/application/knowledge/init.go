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

package knowledge

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/workflow/application/search"
	"github.com/insmtx/corekg/apps/workflow/conf"
	knowledgeImpl "github.com/insmtx/corekg/apps/workflow/domain/knowledge/service"
	"github.com/insmtx/corekg/apps/workflow/infra/eventbus"
	"github.com/insmtx/corekg/apps/workflow/types/consts"
)

type ServiceComponents = knowledgeImpl.KnowledgeSVCConfig

func InitService(ctx context.Context, c *ServiceComponents, bus search.ResourceEventBus) (*KnowledgeApplicationService, error) {

	var (
		knowledgeDomainSVC    knowledgeImpl.Knowledge
		knowledgeEventHandler eventbus.ConsumerHandler
	)

	knowledgeDomainSVC = knowledgeImpl.NewYyguKnowledgeSVC()

	// knowledgeDomainSVC, knowledgeEventHandler = knowledgeImpl.NewKnowledgeSVC(c)

	if knowledgeEventHandler != nil {
		nameServer := conf.GetAppConfig().Workflow.MQ.NameServer
		if err := eventbus.GetDefaultSVC().RegisterConsumer(nameServer, consts.RMQTopicKnowledge, consts.RMQConsumeGroupKnowledge, knowledgeEventHandler); err != nil {
			return nil, fmt.Errorf("register knowledge consumer failed, err=%w", err)
		}
	}

	KnowledgeSVC.DomainSVC = knowledgeDomainSVC
	KnowledgeSVC.eventBus = bus
	KnowledgeSVC.storage = c.Storage
	return KnowledgeSVC, nil
}
