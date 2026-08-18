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

// QueryAgentVersionListResponse Agent版本历史列表
type QueryAgentVersionListResponse struct {
	apiobj.QueryResponse
	Data []*chattype.ChatAgentVersion
}

// QueryAgentVersionList 查询Agent版本历史列表
func QueryAgentVersionList(ctx context.Context, opt apiobj.PageQuery, agent_id uint, agentVersionHistoryList *QueryAgentVersionListResponse) error {
	query := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameAgentVersion).
		Where("agent_id = ?", agent_id).
		Where("deleted_at is null")
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "agent_id":
			query = query.Where(chattype.TableNameAgentVersion+".`agent_id` = ?", filter.Value[0])
		case "version":
			query = query.Where(chattype.TableNameAgentVersion+".`id` = ?", filter.Value[0])
		default:
			logs.ErrorContextf(ctx, "[chat][QueryAgentVersionHistoryList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&agentVersionHistoryList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "query agent version history count error: %v", err)
		return err
	}
	if agentVersionHistoryList.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	err := query.Find(&agentVersionHistoryList.Data).Error
	if err != nil {
		logs.ErrorContextf(ctx, "query agent version history list error: %v", err)
		return err
	}
	return nil
}

// ChooseAgentVersion 选择Agent版本
func ChooseAgentVersion(ctx context.Context, agent_id uint, version uint) error {
	// 更新机器人版本
	if err := dbutil.Chat().WithContext(ctx).
		Model(&chattype.ChatAgent{}).Where("id = ?", agent_id).Update("version", version).Error; err != nil {
		logs.ErrorContextf(ctx, "update agent version error: %v", err)
		return err
	}
	return nil
}
