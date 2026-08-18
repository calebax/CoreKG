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

package singleagent

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/workflow/api/model/app/bot_common"
	"github.com/insmtx/corekg/apps/workflow/api/model/app/developer_api"
	intelligence "github.com/insmtx/corekg/apps/workflow/api/model/app/intelligence/common"
	"github.com/insmtx/corekg/apps/workflow/application/base/ctxutil"
	"github.com/insmtx/corekg/apps/workflow/bizpkg/config"
	singleagent "github.com/insmtx/corekg/apps/workflow/crossdomain/agent/model"
	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/entity"
	searchEntity "github.com/insmtx/corekg/apps/workflow/domain/search/entity"
	"github.com/insmtx/corekg/apps/workflow/pkg/errorx"
	"github.com/insmtx/corekg/apps/workflow/pkg/lang/ptr"
	"github.com/insmtx/corekg/apps/workflow/types/errno"
)

type FullSingleAgentCreateRequest struct {
	SpaceID            int64
	Name               string
	Description        string
	IconURI            string
	Prompt             string
	PluginInfos        []*bot_common.PluginInfo
	Prologue           string
	SuggestedQuestions []string
	Knowledge          *bot_common.Knowledge
	Workflow           []*bot_common.WorkflowInfo
}

func (s *SingleAgentApplicationService) CreateSingleAgentDraft(ctx context.Context, req *developer_api.DraftBotCreateRequest) (*developer_api.DraftBotCreateResponse, error) {
	modelList, err := config.ModelConf().GetOnlineModelListWithLimit(ctx, 1)
	if err != nil {
		return nil, err
	}

	if len(modelList) == 0 {
		return nil, errorx.New(errno.ErrAgentNoModelInUseCode)
	}

	do, err := s.draftBotCreateRequestToSingleAgent(ctx, req)
	if err != nil {
		return nil, err
	}

	userID := ctxutil.MustGetUIDFromCtx(ctx)
	agentID, err := s.DomainSVC.CreateSingleAgentDraft(ctx, userID, do)
	if err != nil {
		return nil, err
	}

	err = s.appContext.EventBus.PublishProject(ctx, &searchEntity.ProjectDomainEvent{
		OpType: searchEntity.Created,
		Project: &searchEntity.ProjectDocument{
			Status:  intelligence.IntelligenceStatus_Using,
			Type:    intelligence.IntelligenceType_Bot,
			ID:      agentID,
			SpaceID: &req.SpaceID,
			OwnerID: &userID,
			Name:    &do.Name,
		},
	})
	if err != nil {
		return nil, err
	}

	return &developer_api.DraftBotCreateResponse{Data: &developer_api.DraftBotCreateData{
		BotID: agentID,
	}}, nil
}

func (s *SingleAgentApplicationService) CreateFullSingleAgent(ctx context.Context, req *FullSingleAgentCreateRequest) (int64, error) {
	sa, err := s.fullAgentCreateRequestToSingleAgent(ctx, req)
	if err != nil {
		return 0, err
	}

	userID := ctxutil.MustGetUIDFromCtx(ctx)
	agentID, err := s.DomainSVC.CreateSingleAgentDraft(ctx, userID, sa)
	if err != nil {
		return 0, err
	}

	err = s.appContext.EventBus.PublishProject(ctx, &searchEntity.ProjectDomainEvent{
		OpType: searchEntity.Created,
		Project: &searchEntity.ProjectDocument{
			Status:  intelligence.IntelligenceStatus_Using,
			Type:    intelligence.IntelligenceType_Bot,
			ID:      agentID,
			SpaceID: &req.SpaceID,
			OwnerID: &userID,
			Name:    &sa.Name,
		},
	})
	if err != nil {
		return 0, err
	}

	return agentID, nil
}

func (s *SingleAgentApplicationService) draftBotCreateRequestToSingleAgent(ctx context.Context, req *developer_api.DraftBotCreateRequest) (*entity.SingleAgent, error) {
	sa, err := s.newDefaultSingleAgent(ctx)
	if err != nil {
		return nil, err
	}

	sa.SpaceID = req.SpaceID
	sa.Name = req.GetName()
	sa.Desc = req.GetDescription()
	sa.IconURI = req.GetIconURI()

	return sa, nil
}

func (s *SingleAgentApplicationService) fullAgentCreateRequestToSingleAgent(ctx context.Context, req *FullSingleAgentCreateRequest) (*entity.SingleAgent, error) {
	if req == nil {
		return nil, errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "request is nil"))
	}

	sa, err := s.newDefaultSingleAgent(ctx)
	if err != nil {
		return nil, err
	}

	sa.SpaceID = req.SpaceID
	sa.Name = req.Name
	sa.Desc = req.Description
	sa.IconURI = req.IconURI

	if req.Prompt != "" {
		sa.Prompt = &bot_common.PromptInfo{
			Prompt: ptr.Of(req.Prompt),
		}
	}

	if req.Prologue != "" || len(req.SuggestedQuestions) > 0 {
		if sa.OnboardingInfo == nil {
			sa.OnboardingInfo = &bot_common.OnboardingInfo{}
		}
		if req.Prologue != "" {
			sa.OnboardingInfo.Prologue = ptr.Of(req.Prologue)
		}
		if len(req.SuggestedQuestions) > 0 {
			sa.OnboardingInfo.SuggestedQuestions = req.SuggestedQuestions
		}
	}

	if len(req.PluginInfos) > 0 {
		for _, pluginInfo := range req.PluginInfos {
			if pluginInfo != nil && pluginInfo.PluginFrom == nil {
				pluginInfo.PluginFrom = bot_common.PluginFromPtr(bot_common.PluginFrom_Default)
			}
		}
		sa.Plugin = req.PluginInfos
	}

	if req.Knowledge != nil {
		sa.Knowledge = mergeKnowledge(sa.Knowledge, req.Knowledge)
	}

	if len(req.Workflow) > 0 {
		sa.Workflow = req.Workflow
	}

	return sa, nil
}

func mergeKnowledge(base, override *bot_common.Knowledge) *bot_common.Knowledge {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base
	if override.KnowledgeInfo != nil {
		merged.KnowledgeInfo = override.KnowledgeInfo
	}
	if override.TopK != nil {
		merged.TopK = override.TopK
	}
	if override.MinScore != nil {
		merged.MinScore = override.MinScore
	}
	if override.Auto != nil {
		merged.Auto = override.Auto
	}
	if override.SearchStrategy != nil {
		merged.SearchStrategy = override.SearchStrategy
	}
	if override.ShowSource != nil {
		merged.ShowSource = override.ShowSource
	}
	if override.NoRecallReplyMode != nil {
		merged.NoRecallReplyMode = override.NoRecallReplyMode
	}
	if override.NoRecallReplyCustomizePrompt != nil {
		merged.NoRecallReplyCustomizePrompt = override.NoRecallReplyCustomizePrompt
	}
	if override.ShowSourceMode != nil {
		merged.ShowSourceMode = override.ShowSourceMode
	}
	if override.RecallStrategy != nil {
		merged.RecallStrategy = override.RecallStrategy
	}

	return &merged
}

func (s *SingleAgentApplicationService) newDefaultSingleAgent(ctx context.Context) (*entity.SingleAgent, error) {
	mi, err := s.defaultModelInfo(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	return &entity.SingleAgent{
		SingleAgent: &singleagent.SingleAgent{
			OnboardingInfo: &bot_common.OnboardingInfo{},
			ModelInfo:      mi,
			Prompt:         &bot_common.PromptInfo{},
			Plugin:         []*bot_common.PluginInfo{},
			Knowledge: &bot_common.Knowledge{
				TopK:           ptr.Of(int64(1)),
				MinScore:       ptr.Of(0.01),
				SearchStrategy: ptr.Of(bot_common.SearchStrategy_SemanticSearch),
				RecallStrategy: &bot_common.RecallStrategy{
					UseNl2sql:  ptr.Of(true),
					UseRerank:  ptr.Of(true),
					UseRewrite: ptr.Of(true),
				},
			},
			Workflow:     []*bot_common.WorkflowInfo{},
			SuggestReply: &bot_common.SuggestReplyInfo{},
			JumpConfig:   &bot_common.JumpConfig{},
			Database:     []*bot_common.Database{},

			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

func (s *SingleAgentApplicationService) defaultModelInfo(ctx context.Context) (*bot_common.ModelInfo, error) {
	modelList, err := config.ModelConf().GetOnlineModelListWithLimit(ctx, 1)
	if err != nil {
		return nil, err
	}

	if len(modelList) == 0 {
		return nil, errorx.New(errno.ErrAgentResourceNotFound, errorx.KV("type", "model"), errorx.KV("id", "default"))
	}

	dm := modelList[0]

	return &bot_common.ModelInfo{
		ModelId:          ptr.Of(dm.ID),
		Temperature:      dm.GetDefaultTemperature(),
		MaxTokens:        dm.GetDefaultMaxTokens(),
		TopP:             dm.GetDefaultTopP(),
		FrequencyPenalty: dm.GetDefaultFrequencyPenalty(),
		PresencePenalty:  dm.GetDefaultPresencePenalty(),
		TopK:             dm.GetDefaultTopK(),
		ModelStyle:       bot_common.ModelStylePtr(bot_common.ModelStyle_Balance),
		ShortMemoryPolicy: &bot_common.ShortMemoryPolicy{
			ContextMode:  bot_common.ContextModePtr(bot_common.ContextMode_FunctionCall_2),
			HistoryRound: ptr.Of[int32](3),
		},
	}, nil
}
