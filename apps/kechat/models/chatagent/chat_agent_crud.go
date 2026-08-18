package chatagent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/internal/dto/dtoagent"
	"github.com/insmtx/corekg/apps/kechat/models/coze"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/settings"

	"github.com/insmtx/corekg/apps/kechat/models/chatmodel"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/types"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"gorm.io/gorm"
)

// GetAgentDetail 获取指定 agent 的详细信息，包括版本信息
func GetAgentDetail(ctx context.Context, id uint) (*AgentWithVersion, error) {
	var agent AgentWithVersion
	// 初始化查询
	query := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameAgent).
		Select(chattype.TableNameAgent+".*, "+
			"chat_agent_version.description, "+
			"chat_agent_version.greeting_message, "+
			"chat_agent_version.agent_type, "+
			"chat_agent_version.params, "+
			"chat_agent_version.chat_model_ids, "+
			"chat_agent_version.forest_option").
		Joins("left JOIN chat_agent_version ON "+chattype.TableNameAgent+".version = chat_agent_version.id").
		Where(chattype.TableNameAgent+".deleted_at IS NULL").
		Where(chattype.TableNameAgent+".id = ?", id)

	// 查询结果
	err := query.First(&agent).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentDetail err: %v", err)
		return nil, err
	}
	return &agent, nil
}

// GetChatAgentByID 获取机器人详情
func GetChatAgentByID(ctx context.Context, id uint) (*chattype.ChatAgent, error) {
	var agent chattype.ChatAgent
	if err := dbutil.Chat().WithContext(ctx).Where("id = ?", id).Where("deleted_at IS NULL").First(&agent).Error; err != nil {
		logs.ErrorContextf(ctx, "GetChatAgentByID err: %v", err)
		return nil, err
	}
	return &agent, nil
}

// UpdateChatAgent 更新机器人信息
func UpdateChatAgent(ctx context.Context, id uint, agent UpdateChatAgentItem) error {
	db := dbutil.Chat().WithContext(ctx)
	tx := db.Begin()
	if tx.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [UpdateChatAgent] begin transaction failed: %s", tx.Error)
		return tx.Error
	}

	// 查询 Agent
	var existingAgent chattype.ChatAgent
	if err := tx.Table(chattype.TableNameAgent).WithContext(ctx).Where("id = ?", id).First(&existingAgent).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [UpdateChatAgent] 查询 Agent 失败: %s", err)
		tx.Rollback()
		return err
	}

	// 构造更新结构体和 Select 字段
	selectFields := make([]string, 0)
	updateStruct := chattype.ChatAgent{}

	if agent.AvatarURL != "" {
		updateStruct.AvatarURL = agent.AvatarURL
		selectFields = append(selectFields, "avatar_url")
	}
	if agent.ShowName != "" {
		updateStruct.ShowName = agent.ShowName
		selectFields = append(selectFields, "show_name")
	}

	// 执行更新（只更新非空字段）
	if len(selectFields) > 0 {
		if err := tx.Table(chattype.TableNameAgent).
			Where("id = ?", id).
			Select(selectFields).
			Updates(&updateStruct).Error; err != nil {
			logs.ErrorContextf(ctx, "[chat] [UpdateChatAgent] 更新 Agent 失败: %s", err)
			tx.Rollback()
			return err
		}
	}

	// 如果传入了 Description，更新 AgentVersion 表
	if agent.Description != "" {
		if err := tx.Table(chattype.TableNameAgentVersion).
			Where("id = ?", existingAgent.Version).
			Select("description").
			Updates(&chattype.ChatAgentVersion{
				Description: agent.Description,
			}).Error; err != nil {
			logs.ErrorContextf(ctx, "[chat] [UpdateChatAgent] 更新 AgentVersion 失败: %s", err)
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// DeleteChatAgent 删除机器人
func DeleteChatAgent(ctx context.Context, id uint) error {
	db := dbutil.Chat().WithContext(ctx)
	tx := db.Begin()
	if tx.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [DeleteChatAgent] begin transaction failed: %s", tx.Error)
		return tx.Error
	}
	// 删除Agent
	if err := tx.Where("id = ?", id).Delete(&chattype.ChatAgent{}).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [DeleteChatAgent] delete BaseAgent failed: %s", err)
		tx.Rollback()
		return err
	}

	if err := tx.Where("agent_id = ?", id).Delete(&chattype.ChatAgentVersion{}).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [DeleteChatAgent] delete PromptAgent failed: %s", err)
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [DeleteChatAgent] commit transaction failed: %s", err)
		return err
	}

	return nil
}

// GetChatAgentVersionByID 获取机器人版本信息
func GetChatAgentVersionByID(ctx context.Context, agentID, versionID uint) (*chattype.ChatAgentVersion, error) {
	var agentVersion chattype.ChatAgentVersion
	query := dbutil.Chat().WithContext(ctx).Where("id = ? AND agent_id = ?", versionID, agentID).First(&agentVersion).Error
	if query != nil {
		logs.ErrorContextf(ctx, "GetChatAgentVersionByID err: %v", query)
		return nil, fmt.Errorf("failed to query agent version information: %v", query)
	}

	return &agentVersion, nil
}

// GetAgentDetailByName 获取指定 agent 的详细信息，包括版本信息
func GetAgentDetailByName(ctx context.Context, name string) (*AgentWithVersion, error) {
	var agent AgentWithVersion
	// 初始化查询
	query := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameAgent).
		Select(chattype.TableNameAgent+".*, "+
			"chat_agent_version.description, "+
			"chat_agent_version.greeting_message, "+
			"chat_agent_version.agent_type, "+
			"chat_agent_version.chat_model_ids, "+
			"chat_agent_version.prompt_template, "+
			"chat_agent_version.temperature, "+
			"chat_agent_version.params, "+
			"chat_agent_version.forest_option").
		Joins("left JOIN chat_agent_version ON "+chattype.TableNameAgent+".version = chat_agent_version.id").
		Where(chattype.TableNameAgent+".deleted_at IS NULL").
		Where(chattype.TableNameAgent+".name = ?", name)

	// 查询结果
	err := query.First(&agent).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentDetailByName err: %v", err)
		return nil, err
	}
	return &agent, nil
}

func QueryChatAgentList(ctx context.Context, opt apiobj.PageQuery, chatAgentList *QueryChatAgentListResponse) error {
	query := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameAgent + " a ").
		Unscoped().
		Select("a.*, " +
			"v.id AS version_id, " +
			"v.description, " +
			"v.greeting_message, " +
			"v.agent_type, " +
			"v.params").
		Joins("LEFT JOIN chat_agent_version v ON a.version = v.id").
		Where("a.deleted_at IS NULL")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "show_name":
			query = query.Where(chattype.TableNameAgent+".`show_name` = ?", filter.Value[0])
		case "agent_type":
			query = query.Where(chattype.TableNameAgent+".`agent_type` in (?)", filter.Value)
		case "uin":
			query = query.Where(chattype.TableNameAgent+".`uin` = ?", opt.Uin)
		case "created_type":
			query = query.Where(chattype.TableNameAgent+".`created_type` = ?", filter.Value[0])
		default:
			logs.WarnContextf(ctx, "[chat][QueryChatAgentList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// ======== 核心权限控制逻辑 ========
	// 构建有权限查看的 agent_id 子查询
	var authIDs []uint
	if err := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("resource_type = ? AND deleted_at IS NULL", foresttype.ResourceTypeAgent).
		Where("("+
			// 1. 公开权限 (action = 'view' 且 scope_type = 'public')
			"(action = ? AND scope_type = ?) OR "+
			// 2. 个人管理权限 (action = 'manage' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 3. 个人查看权限 (action = 'view' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 4. 公司权限 (action = 'view' 且 scope_type = 'company' 且 scope_id = 当前公司)
			"(action = ? AND scope_type = ? AND scope_id = ?)"+
			")",
			foresttype.ActionView, foresttype.ScopeTypePublic, // 公开
			foresttype.ActionManage, foresttype.ScopeTypeUser, opt.Uin, // 个人管理
			foresttype.ActionView, foresttype.ScopeTypeUser, opt.Uin, // 个人查看
			foresttype.ActionView, foresttype.ScopeTypeCompany, opt.CompanyID). // 公司权限
		Pluck("distinct resource_id", &authIDs).
		Error; err != nil {
		return err
	}

	otherAuth := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameAgent+" a ").
		Select("a.id").
		Where("a.deleted_at IS NULL").
		Where("a.publish_status = ? AND a.uin = ?", chattype.StatusDraft, opt.Uin)

	query = query.Where("("+
		//有权限查看
		"a.id IN (?) OR "+
		//草稿
		"a.id IN (?)"+
		")",
		authIDs,
		otherAuth)

	if err := query.Count(&chatAgentList.Total).Error; err != nil {
		return err
	}
	if chatAgentList.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	if err := query.Find(&chatAgentList.Data).Error; err != nil {
		return err
	}

	collects, err := GetAgentCollectByUin(ctx, opt.Uin)
	if err != nil {
		return err
	}
	collectMap := make(map[uint]bool)
	for _, collect := range collects {
		collectMap[collect.AgentAppID] = true
	}
	// 查询管理权限
	manageAgentIDs := make(map[uint]bool)

	manageScopesIDs := perm.GetManageList(ctx, opt.Uin, foresttype.ResourceTypeAgent)

	for _, scope := range manageScopesIDs {
		manageAgentIDs[scope] = true
	}

	for i, agent := range chatAgentList.Data {
		chatAgentList.Data[i].IsCollected = collectMap[agent.ID]
		chatAgentList.Data[i].IsAdmin = manageAgentIDs[agent.ID]
	}
	return nil
}

// CreateAgentApp 创建Agent
func CreateAgentApp(ctx *gin.Context, uin, companyID uint, agentApp AgentApp) (*chattype.ChatAgent, error) {
	db := dbutil.Chat().WithContext(ctx)
	tx := db.Begin()
	if tx.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] begin transaction failed: %s", tx.Error)
		return nil, tx.Error
	}

	Agent := &chattype.ChatAgent{
		Uin:            uin,
		CompanyID:      companyID,
		AvatarURL:      agentApp.AvatarURL,
		Name:           random.Alphanum(7),
		ShowName:       agentApp.ShowName,
		Path:           "/lesson-plan",
		PublishStatus:  chattype.StatusDraft,
		AgentType:      agentApp.AgentType,
		CozeWorkflowID: agentApp.WorkflowID,
		CozeSpaceID:    agentApp.SpaceID,
	}

	// Create Agent
	if err := tx.Create(Agent).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] create BaseAgent failed: %s", err)
		tx.Rollback()
		return nil, err
	}

	AgentVersion := &chattype.ChatAgentVersion{
		AgentID:     Agent.ID,
		Description: agentApp.Description,
		AgentType:   agentApp.AgentType,
	}
	if AgentVersion.AgentType == chattype.AgentTypeRolePlay {
		conf, err := GetAgentDetailByName(ctx, GetAgentI18nName(ctx, "", global.ChatAgentESChat))
		if err != nil {
			logs.ErrorContextf(ctx, "get ForestPromptConfig config error: %v", err)
			return nil, err
		}
		AgentVersion.ForestOption = chattype.ForestChatOption{
			// ForestIDs:      opt.ForestIDs,
			PromptTemplate: conf.PromptTemplate,
		}
	}

	// Create AgentVersion
	if err := tx.Create(AgentVersion).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] create BaseAgent failed: %s", err)
		tx.Rollback()
		return nil, err
	}
	// 更新 Agent
	result := tx.Where("id = ?", Agent.ID).Updates(&chattype.ChatAgent{
		Version: AgentVersion.ID,
	})
	if result.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] update Agent failed: %s", result.Error)
		tx.Rollback()
		return nil, result.Error
	}

	if agentApp.AgentType == chattype.AgentTypeWorkflow {
		if agentApp.WorkflowID == "" {
			spaceID, workflowID, err := CreateCozeWorkflow(ctx, agentApp.ShowName, agentApp.Description)
			if err != nil {
				logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] create CozeWorkflow failed: %s", err)
				tx.Rollback()
				return nil, err
			}
			result = tx.Where("id = ?", Agent.ID).Updates(&chattype.ChatAgent{
				CozeSpaceID:    spaceID,
				CozeWorkflowID: workflowID,
			})
			if result.Error != nil {
				logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] update Agent failed: %s", result.Error)
				tx.Rollback()
				return nil, result.Error
			}
			if err := chattype.CreateCozeMapping(ctx, &chattype.ChatCozeMapping{
				Uin:      uin,
				Type:     chattype.ChatTypeWorkflow,
				CoreKGID: Agent.ID,
				CozeID:   workflowID,
			}); err != nil {
				logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] create CozeMapping failed: %s", err)
				tx.Rollback()
				return nil, err
			}
			Agent.CozeSpaceID = spaceID
			Agent.CozeWorkflowID = workflowID
		} else {
			if err := chattype.CreateCozeMapping(ctx, &chattype.ChatCozeMapping{
				Uin:      uin,
				Type:     chattype.ChatTypeWorkflow,
				CoreKGID: Agent.ID,
				CozeID:   agentApp.WorkflowID,
			}); err != nil {
				logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] create CozeMapping failed: %s", err)
				tx.Rollback()
				return nil, err
			}
			Agent.CozeSpaceID = agentApp.SpaceID
			Agent.CozeWorkflowID = agentApp.WorkflowID
		}
	}

	if err := dbutil.Knownow().Create(&foresttype.KeResourceScope{
		ResourceType: foresttype.ResourceTypeAgent,
		ResourceID:   Agent.ID,
		ScopeType:    foresttype.ScopeTypeUser,
		ScopeID:      uin,
		Action:       foresttype.ActionManage,
	}).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] create User ResourceScope failed: %s", err)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreateAgentApp] commit transaction failed: %s", err)
		return nil, err
	}

	return Agent, nil
}

// CreatePromptAgent 创建指令型机器人
func CreateAgent(ctx context.Context, opt *CreateAgentItem) error {
	db := dbutil.Chat().WithContext(ctx)
	tx := db.Begin()
	if tx.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [CreatePromptAgent] begin transaction failed: %s", tx.Error)
		return tx.Error
	}

	Agent := &chattype.ChatAgent{
		Uin:           opt.Uin,
		CompanyID:     opt.CompanyID,
		AvatarURL:     opt.AvatarURL,
		Name:          random.Alphanum(7),
		ShowName:      opt.ShowName,
		PublicScope:   opt.PublicScope,
		Path:          "/lesson-plan",
		PublishStatus: chattype.StatusPublished,
		//ManagerIDs:    opt.ManagerIDs,
	}

	if err := tx.Create(Agent).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreatePromptAgent] create Agent failed: %s", err)
		tx.Rollback()
		return err
	}
	promptAgent := &chattype.ChatAgentVersion{
		AgentID:         Agent.ID,
		Description:     opt.Description,
		ChatModelIDs:    opt.ChatModelIDs,
		Temperature:     opt.Temperature,
		PromptTemplate:  opt.PromptTemplate,
		AgentType:       opt.AgentType,
		GreetingMessage: opt.GreetingMessage,
	}
	if opt.AgentType == chattype.AgentTypePrompt {
		for _, param := range opt.Params {
			switch param.InputType {
			case chattype.InputTypeSelect:
				if len(param.InputArray) == 0 {
					logs.ErrorContextf(ctx, "When InputType is select ，must use InputArray")
					return fmt.Errorf("input error")
				}
			}
		}
		promptAgent.Params = opt.Params
	}
	if opt.AgentType == chattype.AgentTypeRolePlay {
		conf, err := GetAgentDetailByName(ctx, GetAgentI18nName(ctx, "", global.ChatAgentESChat))
		if err != nil {
			logs.ErrorContextf(ctx, "get ForestPromptConfig config error: %v", err)
			return err
		}
		promptAgent.ForestOption = chattype.ForestChatOption{
			ForestIDs:      opt.ForestIDs,
			PromptTemplate: conf.PromptTemplate,
		}
	}

	if err := tx.Create(promptAgent).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreatePromptAgent] create CommandAgent failed: %s", err)
		tx.Rollback()
		return err
	}
	// 更新 Agent
	result := tx.Where("id = ?", Agent.ID).Updates(&chattype.ChatAgent{
		Version: promptAgent.ID,
	})
	if result.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [CreatePromptAgent] update Agent failed: %s", result.Error)
		tx.Rollback()
		return result.Error
	}

	if err := tx.Commit().Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [CreatePromptAgent] commit transaction failed: %s", err)
		return err
	}

	return nil
}

// UpdatePromptAgent 更新指令型机器人
func UpdateAgent(ctx context.Context, opt *UpdateAgentItem) error {
	db := dbutil.Chat().WithContext(ctx)
	tx := db.Begin()
	if tx.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] begin transaction failed: %s", tx.Error)
		return tx.Error
	}
	// 创建 PromptAgent
	promptAgent := chattype.ChatAgentVersion{
		AgentID:         opt.ID,
		AgentType:       opt.AgentType,
		Description:     opt.Description,
		ChatModelIDs:    opt.ChatModelIDs,
		Temperature:     opt.Temperature,
		PromptTemplate:  opt.PromptTemplate,
		Params:          opt.Params,
		GreetingMessage: opt.GreetingMessage,
	}
	if opt.AgentType == chattype.AgentTypeRolePlay {
		conf, err := GetAgentDetailByName(ctx, GetAgentI18nName(ctx, "", global.ChatAgentESChat))
		if err != nil {
			logs.ErrorContextf(ctx, "get ForestPromptConfig config error: %v", err)
			return err
		}
		promptAgent.ForestOption = chattype.ForestChatOption{
			ForestIDs:      opt.ForestIDs,
			PromptTemplate: conf.PromptTemplate,
		}
	}
	if err := tx.Create(&promptAgent).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] create PromptAgent failed: %s", err)
		tx.Rollback()
		return err
	}

	// 更新 Agent
	result := tx.Where("id = ?", opt.ID).Updates(&chattype.ChatAgent{
		AvatarURL:     opt.AvatarURL,
		ShowName:      opt.ShowName,
		PublicScope:   opt.PublicScope,
		Version:       promptAgent.ID,
		PublishStatus: chattype.StatusPublished,
		//ManagerIDs:    opt.ManagerIDs,
	})
	if result.Error != nil {
		logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] update BaseAgent failed: %s", result.Error)
		tx.Rollback()
		return result.Error
	}
	if result.RowsAffected == 0 {
		logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] BaseAgent not found with ID: %d", opt.ID)
		tx.Rollback()
		return fmt.Errorf("BaseAgent not found with ID: %d", opt.ID)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] commit transaction failed: %s", err)
		return err
	}

	return nil
}

func GetChatAgentWithPermByID(ctx context.Context, id uint) (*AgentWithPerm, error) {

	var agent chattype.ChatAgent

	if err := dbutil.Chat().WithContext(ctx).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&agent).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetChatAgentWithPermByID] failed to find ChatAgent with id %d: %v", id, err)
		return nil, err
	}

	var (
		scopeIDs, managerIDs []uint
		rss                  []*foresttype.KeResourceScope
	)

	if err := dbutil.Knownow().
		Where("deleted_at IS NULL").
		Where("resource_type", foresttype.ResourceTypeAgent).
		Where("resource_id = ?", agent.ID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Find(&rss).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetChatAgentWithPermByID] failed to find resource scope: %v", err)
		return nil, err
	}

	for _, v := range rss {
		switch v.Action {
		case foresttype.ActionManage:
			managerIDs = append(managerIDs, v.ScopeID)
		case foresttype.ActionView:
			scopeIDs = append(scopeIDs, v.ScopeID)
		}
	}

	return &AgentWithPerm{agent, types.NewUintArray(managerIDs), types.NewUintArray(scopeIDs)}, nil
}

// GetAgentTypeByID 获取指令型机器人详细信息
func GetAgentTypeByID(ctx context.Context, versionID uint) (*AgentItemInfo, error) {
	db := dbutil.Chat().WithContext(ctx)

	var promptAgent chattype.ChatAgentVersion
	// 使用 First 方法查询
	if err := db.Where("id = ?", versionID).
		Where("deleted_at is null").
		First(&promptAgent).Error; err != nil {
		// 只要有错误就返回
		logs.ErrorContextf(ctx, "[chat] [GetPromptAgentTypeByID] failed to find PromptAgent with id %d: %v", versionID, err)
		return nil, fmt.Errorf("failed to find PromptAgent with id %d: %v", versionID, err)
	}

	tempChatModelIDs := promptAgent.ChatModelIDs.Slice()
	logs.DebugContextf(ctx, "GetAgentTypeByID model ids:[%v]", tempChatModelIDs)
	// 查询 llm_model 表，仅获取 name 字段
	chatModelNames, err := chatmodel.GetModelNameByIDs(ctx, tempChatModelIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetPromptAgentTypeByID] failed to find model name: %v", err)
		return nil, fmt.Errorf("failed to find model name: %v", err)
	}
	logs.DebugContextf(ctx, "GetAgentTypeByID model names:[%v]", chatModelNames)
	var forestInfo []*foresttype.KnownowForest
	if len(promptAgent.ForestOption.ForestIDs) > 0 {
		forestInfo, err = forest.GetForestByIDS(promptAgent.ForestOption.ForestIDs)
		if err != nil {
			logs.ErrorContextf(ctx, "[chat] [GetForestAgentTypeByID] failed to find model name: %v", err)
			return nil, fmt.Errorf("failed to find model name: %v", err)
		}
	}

	// 返回 PromptAgent 的详细信息
	return &AgentItemInfo{
		Description:     promptAgent.Description,
		ChatModels:      chatModelNames,
		Temperature:     promptAgent.Temperature,
		AgentType:       promptAgent.AgentType,
		PromptTemplate:  promptAgent.PromptTemplate,
		Params:          promptAgent.Params,
		GreetingMessage: promptAgent.GreetingMessage,
		Forests:         forestInfo,
	}, nil
}

// UpdateAgentWithPerm 更新机器人并更新权限
func UpdateAgentWithPerm(ctx context.Context, opt *UpdateAgentItem) error {
	return dbutil.Chat().Transaction(func(tx *gorm.DB) error {
		promptAgent := chattype.ChatAgentVersion{
			AgentID:         opt.ID,
			AgentType:       opt.AgentType,
			Description:     opt.Description,
			ChatModelIDs:    opt.ChatModelIDs,
			Temperature:     opt.Temperature,
			PromptTemplate:  opt.PromptTemplate,
			Params:          opt.Params,
			GreetingMessage: opt.GreetingMessage,
		}
		if opt.AgentType == chattype.AgentTypeRolePlay {
			conf, err := GetAgentDetailByName(ctx, GetAgentI18nName(ctx, "", global.ChatAgentESChat))
			if err != nil {
				logs.ErrorContextf(ctx, "get ForestPromptConfig config error: %v", err)
				return err
			}
			promptAgent.ForestOption = chattype.ForestChatOption{
				ForestIDs:      opt.ForestIDs,
				PromptTemplate: conf.PromptTemplate,
			}
		}
		if err := tx.Create(&promptAgent).Error; err != nil {
			logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] create PromptAgent failed: %s", err)
			return err
		}
		if err := tx.Where("id = ?", opt.ID).Updates(&chattype.ChatAgent{
			AvatarURL:     opt.AvatarURL,
			ShowName:      opt.ShowName,
			PublicScope:   opt.PublicScope,
			Version:       promptAgent.ID,
			PublishStatus: chattype.StatusPublished,
		}).Error; err != nil {
			logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] update BaseAgent failed: %s", err)
			return err
		}

		return perm.UpdateResourceScope(ctx, dbutil.Knownow(), opt.ID, foresttype.ResourceTypeAgent, opt.ScopeIDs.Slice(), opt.ManagerIDs.Slice(), foresttype.PublicScope(opt.PublicScope), opt.CompanyID)
	})
}

// GetALLAgentsByCompanyID get all agents that a company can meet include drafts
func GetALLAgentsByCompanyID(ctx context.Context, companyID uint) (res []*chattype.ChatAgent, err error) {
	err = dbutil.Chat().WithContext(ctx).
		Table(chattype.TableNameAgent+" AS f").
		Where("f.deleted_at IS NULL").
		Where("f.public_scope != ?", chattype.PublicScopePrivate).
		Where("f.company_id = ?", companyID).
		Find(&res).Error
	return
}

func CreateCozeWorkflow(ctx *gin.Context, name, desc string) (spaceID, workflowID string, err error) {
	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
		return "", "", err
	}
	sessionKey := runtime.LoginStatus(ctx).Token
	space, code, err := coze.GetSpaceAPI(ctx, cozeUrl, sessionKey)
	if err != nil {
		logs.ErrorContextf(ctx, "get coze space error, %s", err.Error())
		return "", "", err
	}
	if code != 0 || space == "" {
		logs.ErrorContextf(ctx, "get coze space error, code is not 0")
		return "", "", err
	}

	workflowID, err = coze.CreateCozeWorkflowAPI(ctx, desc, name, space, sessionKey, cozeUrl)
	if err != nil {
		logs.ErrorContextf(ctx, "create coze workflow error, %s", err.Error())
		return "", "", err
	}
	return space, workflowID, nil
}

func UpdateWorkflowVsrsion(ctx context.Context, agentID uint, version string) error {
	return dbutil.Chat().WithContext(ctx).
		Table(chattype.TableNameAgent).
		Where("deleted_at IS NULL").
		Where("id = ?", agentID).
		Updates(map[string]interface{}{
			"workflow_version": version,
		}).Error
}

// UpdateChatAgentInput 更新机器人入参
func UpdateChatAgentInput(ctx context.Context, opt *UpdateAgentItem) error {
	return dbutil.Chat().Transaction(func(tx *gorm.DB) error {
		promptAgent := chattype.ChatAgentVersion{
			AgentID:         opt.ID,
			AgentType:       opt.AgentType,
			Description:     opt.Description,
			ChatModelIDs:    opt.ChatModelIDs,
			Temperature:     opt.Temperature,
			PromptTemplate:  opt.PromptTemplate,
			Params:          opt.Params,
			GreetingMessage: opt.GreetingMessage,
		}

		if err := tx.Create(&promptAgent).Error; err != nil {
			logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] create PromptAgent failed: %s", err)
			return err
		}
		if err := tx.Where("id = ?", opt.ID).Updates(&chattype.ChatAgent{
			AvatarURL:     opt.AvatarURL,
			ShowName:      opt.ShowName,
			PublicScope:   opt.PublicScope,
			Version:       promptAgent.ID,
			PublishStatus: chattype.StatusPublished,
		}).Error; err != nil {
			logs.ErrorContextf(ctx, "[chat] [UpdatePromptAgent] update BaseAgent failed: %s", err)
			return err
		}
		return nil
	})
}

// GetLatestAgent 获取最近使用的机器人 (最终修正版，GORM V2 兼容)
func GetLatestAgent(ctx context.Context, opt apiobj.PageQuery) (res []dtoagent.AgentItem, apiRes *apiobj.QueryResponse, err error) {
	res = make([]dtoagent.AgentItem, 0)
	apiRes = &apiobj.QueryResponse{
		Total:  0,
		Limit:  opt.Limit,
		Offset: opt.Offset,
	}
	db := dbutil.Chat().WithContext(ctx)

	// --- 步骤 1: 构建活跃会话子查询 (top_sessions) ---
	// topSessionsTable 现在是一个 *gorm.DB 实例
	topSessionsTable := db.Table(chattype.TableNameChatSessions+" AS s").
		Select("s.base_agent_id, MAX(s.updated_at) AS last_active_time").
		Where("s.resource_type = ?", chattype.ResourceTypeAgent).
		Where("s.deleted_at IS NULL").
		Where("s.uin = ?", opt.Uin).
		Group("s.base_agent_id")

	//fix permission
	var authIDs []uint
	if err = dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKeResourceScope).
		Where("resource_type = ? AND deleted_at IS NULL", foresttype.ResourceTypeAgent).
		Where("("+
			// 1. 公开权限 (action = 'view' 且 scope_type = 'public')
			"(action = ? AND scope_type = ?) OR "+
			// 2. 个人管理权限 (action = 'manage' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 3. 个人查看权限 (action = 'view' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 4. 公司权限 (action = 'view' 且 scope_type = 'company' 且 scope_id = 当前公司)
			"(action = ? AND scope_type = ? AND scope_id = ?)"+
			")",
			foresttype.ActionView, foresttype.ScopeTypePublic, // 公开
			foresttype.ActionManage, foresttype.ScopeTypeUser, opt.Uin, // 个人管理
			foresttype.ActionView, foresttype.ScopeTypeUser, opt.Uin, // 个人查看
			foresttype.ActionView, foresttype.ScopeTypeCompany, opt.CompanyID). // 公司权限
		Pluck("distinct resource_id", &authIDs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetLatestAgent] get authIDs failed: %s", err)
		return
	}

	// --- 步骤 2: 构建主查询的基础部分 ---
	mainQuery := db.WithContext(ctx).Table(chattype.TableNameAgent+" AS a").
		Where("a.deleted_at IS NULL").
		Where("a.id IN (?)",
			authIDs)

	// 应用 Name 过滤器
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "show_name":
			mainQuery = mainQuery.Where("a.`show_name` like ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[chat][GetLatestAgent] invalid filter field: %s", filter.Field)
			return res, apiRes, fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// --- 步骤 3: 获取 Total Count (总数) ---

	// GORM 会将其识别为子查询并正确生成 SQL：INNER JOIN (SELECT ...) AS top_sessions
	countQuery := mainQuery.Session(&gorm.Session{}).
		Joins("INNER JOIN (?) AS top_sessions ON a.id = top_sessions.base_agent_id", topSessionsTable)

	if err := countQuery.Count(&apiRes.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetLatestAgent] get total count failed: %s", err)
		return res, apiRes, err
	}
	if apiRes.Total == 0 {
		return res, apiRes, nil
	}

	// --- 步骤 4: 构建最终数据查询 (Find) ---

	selectFields := []string{
		"a.id",
		"a.name",
		"a.show_name",
		"a.avatar_url",
		"a.agent_type",
	}

	// 💡 关键修正：Joins 传入 *gorm.DB 实例
	dataQuery := mainQuery.Select(selectFields).
		Joins("INNER JOIN (?) AS top_sessions ON a.id = top_sessions.base_agent_id", topSessionsTable)

	// 处理 OrderBy 和 Limit/Offset
	if len(opt.OrderBy) > 0 {
		dataQuery = dataQuery.Order(strings.Join(opt.OrderBy, ","))
	} else {
		// 依赖子查询中的 last_active_time 排序
		dataQuery = dataQuery.Order("top_sessions.last_active_time DESC")
	}

	// 应用分页
	dataQuery = dataQuery.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		dataQuery = dataQuery.Limit(opt.Limit)
	}

	if err := dataQuery.Find(&res).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetLatestAgent] get result failed: %s", err)
		return res, apiRes, err
	}

	return
}
