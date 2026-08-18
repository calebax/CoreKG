package svcdashboard

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/account/models/account"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
)

const (
	StatKeyUserTotal                    = "user_total"
	StatKeyNewUserTotal                 = "new_user_total"
	StatKeyQATotal                      = "qa_total"
	StatKeyForestTotal                  = "forest_total"
	StatKeyAgentTotal                   = "agent_total"
	StatKeyMultimodalUploadFileTotal    = "multimodal_upload_file_total"
	StatKeyMultimodalReadyResourceTotal = "multimodal_ready_resource_total"
	StatKeyFileParseSuccessTotal        = "file_parse_success_total"
	StatKeyFileParseFailTotal           = "file_parse_fail_total"
	StatKeyFileParseAvgCost             = "file_parse_avg_cost"
	StatKeyFileIndexSuccessTotal        = "file_index_success_total"
	StatKeyFileIndexFailTotal           = "file_index_fail_total"
	StatKeyFileIndexAvgCost             = "file_index_avg_cost"
	StatKeyFileSummarySuccessTotal      = "file_summary_success_total"
	StatKeyFileSummaryFailTotal         = "file_summary_fail_total"
	StatKeyFileSummaryAvgCost           = "file_summary_avg_cost"
	StatKeyGraphTotal                   = "graph_total"
	StatKeyGraphSuccessTotal            = "graph_success_total"
	StatKeyGraphFailTotal               = "graph_fail_total"
)

const (
	forestFileStatusFieldParse     = "parse_status"
	forestFileStatusFieldDesc      = "desc_status"
	forestFileStatusFieldKnowledge = "knowledge_status"
)

var forestFileStatusFieldMap = map[string]struct{}{
	forestFileStatusFieldParse:     {},
	forestFileStatusFieldDesc:      {},
	forestFileStatusFieldKnowledge: {},
}

func countUser(ctx context.Context, query *StatQuery) (MetricValue, error) {
	total, err := account.NewUserDao().CountByCond(ctx, &account.UserCond{})
	if err != nil {
		return nil, fmt.Errorf("countUser failed: %v", err)
	}
	return IntMetric{
		Value: total,
	}, nil
}

func countNewUser(ctx context.Context, query *StatQuery) (MetricValue, error) {
	total, err := account.NewUserDao().CountByCond(ctx, &account.UserCond{
		BaseCond: account.BaseCond{
			BeginTime: query.BeginAt,
			EndTime:   query.EndAt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("countNewUser failed: %v", err)
	}
	return IntMetric{
		Value: total,
	}, nil
}

func countForest(ctx context.Context, query *StatQuery) (MetricValue, error) {
	total, err := forest.NewForestDao().CountByCond(ctx, &forest.ForestCond{
		BaseCond: forest.BaseCond{
			CompanyID: query.CompanyID,
			BeginTime: query.BeginAt,
			EndTime:   query.EndAt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("countForest failed: %v", err)
	}
	return IntMetric{
		Value: total,
	}, nil
}

func countAgent(ctx context.Context, query *StatQuery) (MetricValue, error) {
	total, err := chat.NewChatAgentDao().CountByCond(ctx, &chat.ChatAgentCond{
		BaseCond: chat.BaseCond{
			CompanyID: query.CompanyID,
			BeginTime: query.BeginAt,
			EndTime:   query.EndAt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("countAgent failed: %v", err)
	}
	return IntMetric{
		Value: total,
	}, nil
}

func countGraph(ctx context.Context, query *StatQuery) (MetricValue, error) {
	total, err := forest.NewKeForestGraphDao().CountByCond(ctx, &forest.KeForestGraphCond{
		BaseCond: forest.BaseCond{
			CompanyID: query.CompanyID,
			BeginTime: query.BeginAt,
			EndTime:   query.EndAt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("countGraph failed: %v", err)
	}
	return IntMetric{
		Value: total,
	}, nil
}

func countQA(ctx context.Context, query *StatQuery) (MetricValue, error) {
	total, err := essearch.StatChatHistory(ctx, &essearch.StatChatHistoryReq{
		CreatedAtStart: query.BeginAt,
		CreatedAtEnd:   query.EndAt,
	})
	if err != nil {
		return nil, fmt.Errorf("countQA failed: %v", err)
	}
	return IntMetric{
		Value: total,
	}, nil
}

func countMultimodalUploadFile(ctx context.Context, query *StatQuery) (MetricValue, error) {
	db := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKnownowForestFile).Model(&foresttype.KnownowForestFile{}).
		Joins(fmt.Sprintf("join %s on %s.id = %s.forest_id", foresttype.TableNameKnownowForest, foresttype.TableNameKnownowForest, foresttype.TableNameKnownowForestFile)).
		Where(fmt.Sprintf("%s.forest_type = ?", foresttype.TableNameKnownowForest), foresttype.ForestTypeFile)
	total, err := forest.NewForestFileDao().WithTx(db).CountByCond(ctx, &forest.ForestFileCond{
		BaseCond: forest.BaseCond{
			CompanyID: query.CompanyID,
			BeginTime: query.BeginAt,
			EndTime:   query.EndAt,
		},
		IsDir: types.False,
	})
	if err != nil {
		return nil, fmt.Errorf("countMultimodalUploadFile failed: %v", err)
	}

	logs.InfoContextf(ctx, "countMultimodalUploadFile total: %d", total)

	return IntMetric{
		Value: total,
	}, nil
}

func buildCountMultimodalFileFunc(statusFieldMap map[string]foresttype.KnownowForestTaskStatus) StatFunc {
	return func(ctx context.Context, query *StatQuery) (MetricValue, error) {
		db := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKnownowForestFile).Model(&foresttype.KnownowForestFile{}).
			Joins(fmt.Sprintf("join %s on %s.id = %s.forest_id", foresttype.TableNameKnownowForest, foresttype.TableNameKnownowForest, foresttype.TableNameKnownowForestFile)).
			Where(fmt.Sprintf("%s.forest_type = ?", foresttype.TableNameKnownowForest), foresttype.ForestTypeFile)
		baseCond := forest.BaseCond{
			CompanyID: query.CompanyID,
			BeginTime: query.BeginAt,
			EndTime:   query.EndAt,
		}
		cond := &forest.ForestFileCond{
			BaseCond: baseCond,
			Status:   foresttype.FileStatusNormal,
		}
		for statusField, status := range statusFieldMap {
			switch statusField {
			case forestFileStatusFieldParse:
				cond.ParseStatus = status
			case forestFileStatusFieldDesc:
				cond.DescStatus = status
			case forestFileStatusFieldKnowledge:
				cond.KnowledgeStatus = status
			default:
				return nil, fmt.Errorf("buildCountMultimodalFileFunc invalid statusField: %s", statusField)
			}
		}

		total, err := forest.NewForestFileDao().WithTx(db).CountByCond(ctx, cond)
		if err != nil {
			return nil, fmt.Errorf("buildCountMultimodalFileFunc failed: %v", err)
		}
		return IntMetric{
			Value: total,
		}, nil

	}
}

func buildAvgTaskCostFunc(taskType string, taskStatus task.TaskStatus) StatFunc {
	return func(ctx context.Context, query *StatQuery) (MetricValue, error) {
		avgCost, err := task.NewTaskDao().AvgCostTimeByCond(ctx, &task.TaskCond{
			BaseCond: task.BaseCond{
				BeginTime: query.BeginAt,
				EndTime:   query.EndAt,
			},
			TaskStatus: taskStatus,
			TaskType:   taskType,
		})
		if err != nil {
			return nil, fmt.Errorf("buildAvgTaskCostFunc failed: %v", err)
		}
		return FloatMetric{
			Value: avgCost,
		}, nil
	}
}

func buildCountGraphStatusFunc(graphStatus foresttype.GraphStatus) StatFunc {
	return func(ctx context.Context, query *StatQuery) (MetricValue, error) {
		total, err := forest.NewForestDao().CountByCond(ctx, &forest.ForestCond{
			BaseCond: forest.BaseCond{
				CompanyID: query.CompanyID,
				BeginTime: query.BeginAt,
				EndTime:   query.EndAt,
			},
			GraphStatus: graphStatus,
		})
		if err != nil {
			return nil, fmt.Errorf("buildCountGraphStatusFunc failed: %v", err)
		}
		return IntMetric{
			Value: total,
		}, nil
	}
}
