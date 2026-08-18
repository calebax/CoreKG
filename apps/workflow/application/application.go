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

package application

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/workflow/application/app"
	"github.com/insmtx/corekg/apps/workflow/application/base/appinfra"
	"github.com/insmtx/corekg/apps/workflow/application/connector"
	"github.com/insmtx/corekg/apps/workflow/application/conversation"
	"github.com/insmtx/corekg/apps/workflow/application/knowledge"
	"github.com/insmtx/corekg/apps/workflow/application/memory"
	"github.com/insmtx/corekg/apps/workflow/application/modelmgr"
	"github.com/insmtx/corekg/apps/workflow/application/openauth"
	"github.com/insmtx/corekg/apps/workflow/application/permission"
	"github.com/insmtx/corekg/apps/workflow/application/plugin"
	"github.com/insmtx/corekg/apps/workflow/application/prompt"
	"github.com/insmtx/corekg/apps/workflow/application/search"
	"github.com/insmtx/corekg/apps/workflow/application/shortcutcmd"
	"github.com/insmtx/corekg/apps/workflow/application/singleagent"
	"github.com/insmtx/corekg/apps/workflow/application/template"
	"github.com/insmtx/corekg/apps/workflow/application/upload"
	"github.com/insmtx/corekg/apps/workflow/application/user"
	"github.com/insmtx/corekg/apps/workflow/application/workflow"
	"github.com/insmtx/corekg/apps/workflow/conf"
	crossagent "github.com/insmtx/corekg/apps/workflow/crossdomain/agent"
	singleagentImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/agent/impl"
	crossagentrun "github.com/insmtx/corekg/apps/workflow/crossdomain/agentrun"
	agentrunImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/agentrun/impl"
	crossapp "github.com/insmtx/corekg/apps/workflow/crossdomain/app"
	appImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/app/impl"
	crossconnector "github.com/insmtx/corekg/apps/workflow/crossdomain/connector"
	connectorImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/connector/impl"
	crossconversation "github.com/insmtx/corekg/apps/workflow/crossdomain/conversation"
	conversationImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/conversation/impl"
	crossdatabase "github.com/insmtx/corekg/apps/workflow/crossdomain/database"
	databaseImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/database/impl"
	crossdatacopy "github.com/insmtx/corekg/apps/workflow/crossdomain/datacopy"
	dataCopyImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/datacopy/impl"
	crossknowledge "github.com/insmtx/corekg/apps/workflow/crossdomain/knowledge"
	knowledgeImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/knowledge/impl"
	crossmessage "github.com/insmtx/corekg/apps/workflow/crossdomain/message"
	messageImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/message/impl"
	crosspermission "github.com/insmtx/corekg/apps/workflow/crossdomain/permission"
	permissionImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/permission/impl"
	crossplugin "github.com/insmtx/corekg/apps/workflow/crossdomain/plugin"
	pluginImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/plugin/impl"
	crosssearch "github.com/insmtx/corekg/apps/workflow/crossdomain/search"
	searchImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/search/impl"
	crossupload "github.com/insmtx/corekg/apps/workflow/crossdomain/upload"
	uploadImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/upload/impl"
	crossuser "github.com/insmtx/corekg/apps/workflow/crossdomain/user"
	crossuserImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/user/impl"
	crossvariables "github.com/insmtx/corekg/apps/workflow/crossdomain/variables"
	variablesImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/variables/impl"
	crossworkflow "github.com/insmtx/corekg/apps/workflow/crossdomain/workflow"
	workflowImpl "github.com/insmtx/corekg/apps/workflow/crossdomain/workflow/impl"
	"github.com/insmtx/corekg/apps/workflow/infra/checkpoint"
	"github.com/insmtx/corekg/apps/workflow/infra/document/progressbar"
	progressBarImpl "github.com/insmtx/corekg/apps/workflow/infra/document/progressbar/impl/progressbar"
	"github.com/insmtx/corekg/apps/workflow/infra/eventbus"
	implEventbus "github.com/insmtx/corekg/apps/workflow/infra/eventbus/impl"
	"github.com/insmtx/corekg/apps/workflow/infra/sqlparser"
	sqlparserImpl "github.com/insmtx/corekg/apps/workflow/infra/sqlparser/impl/sqlparser"
	"github.com/insmtx/corekg/apps/workflow/pkg/ctxcache"
)

type eventbusImpl struct {
	resourceEventBus search.ResourceEventBus
	projectEventBus  search.ProjectEventBus
}

type basicServices struct {
	infra         *appinfra.AppDependencies
	eventbus      *eventbusImpl
	modelMgrSVC   *modelmgr.ModelmgrApplicationService
	connectorSVC  *connector.ConnectorApplicationService
	userSVC       *user.UserApplicationService
	promptSVC     *prompt.PromptApplicationService
	templateSVC   *template.ApplicationService
	openAuthSVC   *openauth.OpenAuthApplicationService
	uploadSVC     *upload.UploadService
	permissionSVC *permission.PermissionApplicationService
}

type primaryServices struct {
	basicServices *basicServices
	infra         *appinfra.AppDependencies

	pluginSVC    *plugin.PluginApplicationService
	memorySVC    *memory.MemoryApplicationServices
	knowledgeSVC *knowledge.KnowledgeApplicationService
	workflowSVC  *workflow.ApplicationService
	shortcutSVC  *shortcutcmd.ShortcutCmdApplicationService
}

type complexServices struct {
	primaryServices *primaryServices
	singleAgentSVC  *singleagent.SingleAgentApplicationService
	appSVC          *app.APPApplicationService
	searchSVC       *search.SearchApplicationService
	conversationSVC *conversation.ConversationApplicationService
}

func Init(ctx context.Context) (err error) {
	ctx = ctxcache.Init(ctx)
	infra, err := appinfra.Init(ctx, conf.GetAppConfig())
	if err != nil {
		return err
	}

	progressbar.New = progressBarImpl.NewProgressBar
	sqlparser.New = sqlparserImpl.NewSQLParser

	eventbus := initEventBus(infra)

	basicServices, err := initBasicServices(ctx, infra, eventbus)
	if err != nil {
		return fmt.Errorf("Init - initBasicServices failed, err: %v", err)
	}

	primaryServices, err := initPrimaryServices(ctx, basicServices)
	if err != nil {
		return fmt.Errorf("Init - initPrimaryServices failed, err: %v", err)
	}

	complexServices, err := initComplexServices(ctx, primaryServices)
	if err != nil {
		return fmt.Errorf("Init - initVitalServices failed, err: %v", err)
	}

	crosspermission.SetDefaultSVC(permissionImpl.InitDomainService(basicServices.permissionSVC.DomainSVC))
	crossconnector.SetDefaultSVC(connectorImpl.InitDomainService(basicServices.connectorSVC.DomainSVC))
	crossdatabase.SetDefaultSVC(databaseImpl.InitDomainService(primaryServices.memorySVC.DatabaseDomainSVC))

	crossknowledge.SetDefaultSVC(knowledgeImpl.InitDomainService(primaryServices.knowledgeSVC.DomainSVC))

	crossplugin.SetDefaultSVC(pluginImpl.InitDomainService(primaryServices.pluginSVC.DomainSVC, infra.OSS))
	crossvariables.SetDefaultSVC(variablesImpl.InitDomainService(primaryServices.memorySVC.VariablesDomainSVC))
	crossworkflow.SetDefaultSVC(workflowImpl.InitDomainService(primaryServices.workflowSVC.DomainSVC))
	crossconversation.SetDefaultSVC(conversationImpl.InitDomainService(complexServices.conversationSVC.ConversationDomainSVC))
	crossmessage.SetDefaultSVC(messageImpl.InitDomainService(complexServices.conversationSVC.MessageDomainSVC))
	crossagentrun.SetDefaultSVC(agentrunImpl.InitDomainService(complexServices.conversationSVC.AgentRunDomainSVC))
	crossagent.SetDefaultSVC(singleagentImpl.InitDomainService(complexServices.singleAgentSVC.DomainSVC))
	crossuser.SetDefaultSVC(crossuserImpl.InitDomainService(basicServices.userSVC.DomainSVC))
	crossdatacopy.SetDefaultSVC(dataCopyImpl.InitDomainService(basicServices.infra))
	crosssearch.SetDefaultSVC(searchImpl.InitDomainService(complexServices.searchSVC.DomainSVC))
	crossupload.SetDefaultSVC(uploadImpl.InitDomainService(basicServices.uploadSVC.UploadSVC))

	crossapp.SetDefaultSVC(appImpl.InitDomainService(complexServices.appSVC.DomainSVC))
	return nil
}

func initEventBus(infra *appinfra.AppDependencies) *eventbusImpl {
	e := &eventbusImpl{}
	eventbus.SetDefaultSVC(implEventbus.NewConsumerService())
	e.resourceEventBus = search.NewResourceEventBus(infra.ResourceEventProducer)
	e.projectEventBus = search.NewProjectEventBus(infra.AppEventProducer)

	return e
}

// initBasicServices init basic services that only depends on infra.
func initBasicServices(ctx context.Context, infra *appinfra.AppDependencies, e *eventbusImpl) (*basicServices, error) {
	uploadSVC := upload.InitService(&upload.UploadComponents{Cache: infra.CacheCli, Oss: infra.OSS, DB: infra.DB, Idgen: infra.IDGenSVC})
	openAuthSVC := openauth.InitService(infra.DB, infra.IDGenSVC)
	promptSVC := prompt.InitService(infra.DB, infra.IDGenSVC, e.resourceEventBus)
	modelMgrSVC := modelmgr.InitService(infra.OSS)
	connectorSVC := connector.InitService(infra.OSS)
	userSVC := user.InitService(ctx, infra.DB, infra.OSS, infra.IDGenSVC)
	templateSVC := template.InitService(ctx, &template.ServiceComponents{
		DB:      infra.DB,
		IDGen:   infra.IDGenSVC,
		Storage: infra.OSS,
	})
	permissionSVC := permission.InitService(&permission.ServiceComponents{})

	return &basicServices{
		infra:         infra,
		eventbus:      e,
		modelMgrSVC:   modelMgrSVC,
		connectorSVC:  connectorSVC,
		userSVC:       userSVC,
		promptSVC:     promptSVC,
		templateSVC:   templateSVC,
		openAuthSVC:   openAuthSVC,
		uploadSVC:     uploadSVC,
		permissionSVC: permissionSVC,
	}, nil
}

// initPrimaryServices init primary services that depends on basic services.
func initPrimaryServices(ctx context.Context, basicServices *basicServices) (*primaryServices, error) {
	pluginSVC, err := plugin.InitService(ctx, basicServices.toPluginServiceComponents())
	if err != nil {
		return nil, err
	}

	memorySVC := memory.InitService(basicServices.toMemoryServiceComponents())

	knowledgeSVC, err := knowledge.InitService(ctx,
		basicServices.toKnowledgeServiceComponents(memorySVC),
		basicServices.eventbus.resourceEventBus)
	if err != nil {
		return nil, err
	}

	workflowDomainSVC, err := workflow.InitService(ctx,
		basicServices.toWorkflowServiceComponents(pluginSVC, memorySVC, knowledgeSVC))
	if err != nil {
		return nil, err
	}

	shortcutSVC := shortcutcmd.InitService(basicServices.infra.DB, basicServices.infra.IDGenSVC)

	return &primaryServices{
		basicServices: basicServices,
		pluginSVC:     pluginSVC,
		memorySVC:     memorySVC,
		knowledgeSVC:  knowledgeSVC,
		workflowSVC:   workflowDomainSVC,
		shortcutSVC:   shortcutSVC,
		infra:         basicServices.infra,
	}, nil
}

// initComplexServices init complex services that depends on primary services.
func initComplexServices(ctx context.Context, p *primaryServices) (*complexServices, error) {
	singleAgentSVC, err := singleagent.InitService(p.toSingleAgentServiceComponents())
	if err != nil {
		return nil, err
	}

	appSVC, err := app.InitService(p.toAPPServiceComponents())
	if err != nil {
		return nil, err
	}

	searchSVC, err := search.InitService(ctx, p.toSearchServiceComponents(singleAgentSVC, appSVC))
	if err != nil {
		return nil, err
	}

	conversationSVC := conversation.InitService(p.toConversationComponents(singleAgentSVC))

	return &complexServices{
		primaryServices: p,
		singleAgentSVC:  singleAgentSVC,
		appSVC:          appSVC,
		searchSVC:       searchSVC,
		conversationSVC: conversationSVC,
	}, nil
}

func (b *basicServices) toPluginServiceComponents() *plugin.ServiceComponents {
	return &plugin.ServiceComponents{
		IDGen:    b.infra.IDGenSVC,
		DB:       b.infra.DB,
		EventBus: b.eventbus.resourceEventBus,
		OSS:      b.infra.OSS,
		UserSVC:  b.userSVC.DomainSVC,
	}
}

func (b *basicServices) toKnowledgeServiceComponents(memoryService *memory.MemoryApplicationServices) *knowledge.ServiceComponents {
	return &knowledge.ServiceComponents{
		DB:                  b.infra.DB,
		IDGen:               b.infra.IDGenSVC,
		RDB:                 memoryService.RDBDomainSVC,
		Producer:            b.infra.KnowledgeEventProducer,
		SearchStoreManagers: b.infra.SearchStoreManagers,
		ParseManager:        b.infra.ParserManager,
		Storage:             b.infra.OSS,
		Rewriter:            b.infra.Rewriter,
		Reranker:            b.infra.Reranker,
		NL2Sql:              b.infra.NL2SQL,
		CacheCli:            b.infra.CacheCli,
	}
}

func (b *basicServices) toMemoryServiceComponents() *memory.ServiceComponents {
	return &memory.ServiceComponents{
		IDGen:                  b.infra.IDGenSVC,
		DB:                     b.infra.DB,
		EventBus:               b.eventbus.resourceEventBus,
		TosClient:              b.infra.OSS,
		ResourceDomainNotifier: b.eventbus.resourceEventBus,
		CacheCli:               b.infra.CacheCli,
	}
}

func (b *basicServices) toWorkflowServiceComponents(pluginSVC *plugin.PluginApplicationService, memorySVC *memory.MemoryApplicationServices, knowledgeSVC *knowledge.KnowledgeApplicationService) *workflow.ServiceComponents {
	return &workflow.ServiceComponents{
		IDGen:                    b.infra.IDGenSVC,
		DB:                       b.infra.DB,
		Cache:                    b.infra.CacheCli,
		Tos:                      b.infra.OSS,
		ImageX:                   b.infra.ImageXClient,
		DatabaseDomainSVC:        memorySVC.DatabaseDomainSVC,
		VariablesDomainSVC:       memorySVC.VariablesDomainSVC,
		PluginDomainSVC:          pluginSVC.DomainSVC,
		KnowledgeDomainSVC:       knowledgeSVC.DomainSVC,
		DomainNotifier:           b.eventbus.resourceEventBus,
		CPStore:                  checkpoint.NewRedisStore(b.infra.CacheCli),
		CodeRunner:               b.infra.CodeRunner,
		WorkflowBuildInChatModel: b.infra.WorkflowBuildInChatModel,
	}
}

func (p *primaryServices) toSingleAgentServiceComponents() *singleagent.ServiceComponents {
	return &singleagent.ServiceComponents{
		IDGen:                p.basicServices.infra.IDGenSVC,
		DB:                   p.basicServices.infra.DB,
		Cache:                p.basicServices.infra.CacheCli,
		TosClient:            p.basicServices.infra.OSS,
		ImageX:               p.basicServices.infra.ImageXClient,
		UserDomainSVC:        p.basicServices.userSVC.DomainSVC,
		EventBus:             p.basicServices.eventbus.projectEventBus,
		DatabaseDomainSVC:    p.memorySVC.DatabaseDomainSVC,
		ConnectorDomainSVC:   p.basicServices.connectorSVC.DomainSVC,
		KnowledgeDomainSVC:   p.knowledgeSVC.DomainSVC,
		PluginDomainSVC:      p.pluginSVC.DomainSVC,
		WorkflowDomainSVC:    p.workflowSVC.DomainSVC,
		VariablesDomainSVC:   p.memorySVC.VariablesDomainSVC,
		ShortcutCMDDomainSVC: p.shortcutSVC.ShortCutDomainSVC,
		CPStore:              checkpoint.NewRedisStore(p.infra.CacheCli),
	}
}

func (p *primaryServices) toSearchServiceComponents(singleAgentSVC *singleagent.SingleAgentApplicationService, appSVC *app.APPApplicationService) *search.ServiceComponents {
	infra := p.basicServices.infra

	return &search.ServiceComponents{
		DB:                   infra.DB,
		Cache:                infra.CacheCli,
		TOS:                  infra.OSS,
		ESClient:             infra.ESClient,
		ProjectEventBus:      p.basicServices.eventbus.projectEventBus,
		ResourceEventBus:     p.basicServices.eventbus.resourceEventBus,
		SingleAgentDomainSVC: singleAgentSVC.DomainSVC,
		APPDomainSVC:         appSVC.DomainSVC,
		KnowledgeDomainSVC:   p.knowledgeSVC.DomainSVC,
		PluginDomainSVC:      p.pluginSVC.DomainSVC,
		WorkflowDomainSVC:    p.workflowSVC.DomainSVC,
		UserDomainSVC:        p.basicServices.userSVC.DomainSVC,
		ConnectorDomainSVC:   p.basicServices.connectorSVC.DomainSVC,
		PromptDomainSVC:      p.basicServices.promptSVC.DomainSVC,
		DatabaseDomainSVC:    p.memorySVC.DatabaseDomainSVC,
	}
}

func (p *primaryServices) toAPPServiceComponents() *app.ServiceComponents {
	infra := p.basicServices.infra
	basic := p.basicServices
	return &app.ServiceComponents{
		IDGen:           infra.IDGenSVC,
		DB:              infra.DB,
		OSS:             infra.OSS,
		CacheCli:        infra.CacheCli,
		ProjectEventBus: basic.eventbus.projectEventBus,
		UserSVC:         basic.userSVC.DomainSVC,
		ConnectorSVC:    basic.connectorSVC.DomainSVC,
		VariablesSVC:    p.memorySVC.VariablesDomainSVC,
	}
}

func (p *primaryServices) toConversationComponents(singleAgentSVC *singleagent.SingleAgentApplicationService) *conversation.ServiceComponents {
	infra := p.basicServices.infra

	return &conversation.ServiceComponents{
		DB:                   infra.DB,
		IDGen:                infra.IDGenSVC,
		TosClient:            infra.OSS,
		ImageX:               infra.ImageXClient,
		SingleAgentDomainSVC: singleAgentSVC.DomainSVC,
	}
}
