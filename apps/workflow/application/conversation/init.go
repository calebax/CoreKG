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

package conversation

import (
	"gorm.io/gorm"

	"github.com/insmtx/corekg/apps/workflow/application/singleagent"
	"github.com/insmtx/corekg/apps/workflow/domain/conversation/agentrun/repository"
	agentrun "github.com/insmtx/corekg/apps/workflow/domain/conversation/agentrun/service"
	convRepo "github.com/insmtx/corekg/apps/workflow/domain/conversation/conversation/repository"
	conversation "github.com/insmtx/corekg/apps/workflow/domain/conversation/conversation/service"
	msgRepo "github.com/insmtx/corekg/apps/workflow/domain/conversation/message/repository"
	message "github.com/insmtx/corekg/apps/workflow/domain/conversation/message/service"
	shortcutRepo "github.com/insmtx/corekg/apps/workflow/domain/shortcutcmd/repository"
	"github.com/insmtx/corekg/apps/workflow/domain/shortcutcmd/service"
	uploadService "github.com/insmtx/corekg/apps/workflow/domain/upload/service"
	"github.com/insmtx/corekg/apps/workflow/infra/idgen"
	"github.com/insmtx/corekg/apps/workflow/infra/imagex"
	"github.com/insmtx/corekg/apps/workflow/infra/storage"
)

type ServiceComponents struct {
	IDGen     idgen.IDGenerator
	DB        *gorm.DB
	TosClient storage.Storage
	ImageX    imagex.ImageX

	SingleAgentDomainSVC singleagent.SingleAgent
}

func InitService(s *ServiceComponents) *ConversationApplicationService {
	mDomainComponents := &message.Components{
		MessageRepo: msgRepo.NewMessageRepo(s.DB, s.IDGen),
	}
	messageDomainSVC := message.NewService(mDomainComponents)

	cDomainComponents := &conversation.Components{
		ConversationRepo: convRepo.NewConversationRepo(s.DB, s.IDGen),
	}

	conversationDomainSVC := conversation.NewService(cDomainComponents)

	arDomainComponents := &agentrun.Components{
		RunRecordRepo: repository.NewRunRecordRepo(s.DB, s.IDGen),
		ImagexSVC:     s.ImageX,
	}

	agentRunDomainSVC := agentrun.NewService(arDomainComponents)
	components := &service.Components{
		ShortCutCmdRepo: shortcutRepo.NewShortCutCmdRepo(s.DB, s.IDGen),
	}
	shortcutCmdDomainSVC := service.NewShortcutCommandService(components)

	ConversationSVC.AgentRunDomainSVC = agentRunDomainSVC
	ConversationSVC.MessageDomainSVC = messageDomainSVC
	ConversationSVC.ConversationDomainSVC = conversationDomainSVC
	ConversationSVC.appContext = s
	ConversationSVC.ShortcutDomainSVC = shortcutCmdDomainSVC

	ConversationOpenAPISVC.ShortcutDomainSVC = shortcutCmdDomainSVC
	uploadSVC := uploadService.NewUploadSVC(s.DB, s.IDGen, s.TosClient)
	ConversationOpenAPISVC.UploaodDomainSVC = uploadSVC
	OpenapiMessageSVC.UploaodDomainSVC = uploadSVC

	return ConversationSVC
}
