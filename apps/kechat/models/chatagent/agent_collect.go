package chatagent

import (
	"context"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

type QueryAgentCollectListResponse struct {
	Total int64
	Data  []*AgentCollectInfoItem
}

type AgentCollectInfoItem struct {
	chattype.ChatAgent
	// 是否收藏
	IsCollected bool `json:"is_collected"`
}

// QueryAgentCollectList 查询收藏的 Agent 列表
func QueryAgentCollectList(ctx context.Context, opt apiobj.PageQuery, collectList *QueryAgentCollectListResponse) error {
	// 初始化查询，关联 AgentCollect 和 Agent 表
	query := dbutil.Chat().WithContext(ctx).
		Table(chattype.TableNameAgentCollect+" AS collect").
		Joins("JOIN "+chattype.TableNameAgent+" AS agent ON agent.id = collect.agent_app_id").
		Where("collect.deleted_at IS NULL").
		Where("agent.deleted_at IS NULL").
		Where("collect.uin = ?", opt.Uin)

	// 添加过滤器（例如按 Agent 名称搜索）
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("agent.name LIKE ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[chat][QueryAgentCollectList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// 查询总记录数
	if err := query.Count(&collectList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "QueryAgentCollectList err: %v", err)
		return err
	}
	if collectList.Total == 0 {
		return nil
	}

	// 排序（默认按收藏时间倒序）
	query = query.Order("collect.created_at DESC")

	// 自定义排序（如果 opt.OrderBy 有值）
	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页逻辑
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	// 查询收藏的 Agent 列表
	err := query.
		Select("agent.*, true AS is_collected").
		Find(&collectList.Data).
		Error
	if err != nil {
		logs.ErrorContextf(ctx, "QueryAgentCollectList err: %v", err)
		return err
	}

	return nil
}

// GetAgentCollectByUin 根据用户uin获取收藏记录
func GetAgentCollectByUin(ctx context.Context, uin uint) ([]*chattype.ChatAgentCollect, error) {
	var agent_collects []*chattype.ChatAgentCollect
	err := dbutil.Chat().WithContext(ctx).Where("uin = ?", uin).Find(&agent_collects).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentCollectByUin err: %v", err)
		return nil, err
	}
	return agent_collects, nil
}

// IsExistAgentCollect 检查收藏记录是否存在
func IsExistAgentCollect(ctx context.Context, uin, agent_app_id uint) (bool, error) {
	var count int64
	err := dbutil.Chat().WithContext(ctx).Model(&chattype.ChatAgentCollect{}).
		Where("uin = ? AND agent_app_id = ?", uin, agent_app_id).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "IsExistAgentCollect err: %v", err)
		return false, err
	}
	return count > 0, nil
}

// DeleteAgentCollect 删除收藏记录
func DeleteAgentCollect(ctx context.Context, uin, agent_app_id uint) error {
	var agent_collect chattype.ChatAgentCollect
	err := dbutil.Chat().WithContext(ctx).Unscoped().
		Where("uin = ? AND agent_app_id = ?", uin, agent_app_id).
		Delete(&agent_collect).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteAgentCollect err: %v", err)
		return err
	}
	return nil
}

// CreateAgentCollect 创建收藏记录
func CreateAgentCollect(ctx context.Context, uin, agent_app_id uint) error {
	agent_collect := &chattype.ChatAgentCollect{
		Uin:        uin,
		AgentAppID: agent_app_id,
	}
	err := dbutil.Chat().WithContext(ctx).Create(agent_collect).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateAgentCollect err: %v", err)
		return err
	}
	return nil
}
