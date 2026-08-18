package svcdashboard

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/internal/dto/dtodashboard"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/ygpkg/yg-go/logs"
)

func GetDashboardOverview(ctx *gin.Context, req *dtodashboard.GetDashboardOverviewRequest) (res *dtodashboard.GetDashboardOverviewResponse, err error) {
	res = &dtodashboard.GetDashboardOverviewResponse{}
	mg := NewStatManager()
	multimodalReadyResourceStatusFieldMap := map[string]foresttype.KnownowForestTaskStatus{
		forestFileStatusFieldKnowledge: foresttype.TaskStatusSuccess,
		forestFileStatusFieldParse:     foresttype.TaskStatusSuccess,
		forestFileStatusFieldDesc:      foresttype.TaskStatusSuccess,
	}
	err = mg.BatchRegister(map[string]StatFunc{
		StatKeyUserTotal:                 countUser,
		StatKeyNewUserTotal:              countNewUser,
		StatKeyQATotal:                   countQA,
		StatKeyForestTotal:               countForest,
		StatKeyAgentTotal:                countAgent,
		StatKeyGraphTotal:                countGraph,
		StatKeyMultimodalUploadFileTotal: countMultimodalUploadFile,
		// 拆 chunk 成功相当于资源就绪
		StatKeyMultimodalReadyResourceTotal: buildCountMultimodalFileFunc(multimodalReadyResourceStatusFieldMap),
		StatKeyFileParseSuccessTotal:        buildCountMultimodalFileFunc(map[string]foresttype.KnownowForestTaskStatus{forestFileStatusFieldParse: foresttype.TaskStatusSuccess}),
		StatKeyFileParseFailTotal:           buildCountMultimodalFileFunc(map[string]foresttype.KnownowForestTaskStatus{forestFileStatusFieldParse: foresttype.TaskStatusFail}),
		StatKeyFileParseAvgCost:             buildAvgTaskCostFunc(coretask.PraseTask, task.TaskStatusSuccess),
		StatKeyFileIndexSuccessTotal:        buildCountMultimodalFileFunc(map[string]foresttype.KnownowForestTaskStatus{forestFileStatusFieldKnowledge: foresttype.TaskStatusSuccess}),
		StatKeyFileIndexFailTotal:           buildCountMultimodalFileFunc(map[string]foresttype.KnownowForestTaskStatus{forestFileStatusFieldKnowledge: foresttype.TaskStatusFail}),
		StatKeyFileIndexAvgCost:             buildAvgTaskCostFunc(coretask.KnowledgeTask, task.TaskStatusSuccess),
		StatKeyFileSummarySuccessTotal:      buildCountMultimodalFileFunc(map[string]foresttype.KnownowForestTaskStatus{forestFileStatusFieldDesc: foresttype.TaskStatusSuccess}),
		StatKeyFileSummaryFailTotal:         buildCountMultimodalFileFunc(map[string]foresttype.KnownowForestTaskStatus{forestFileStatusFieldDesc: foresttype.TaskStatusFail}),
		StatKeyFileSummaryAvgCost:           buildAvgTaskCostFunc(coretask.DescriptionTask, task.TaskStatusSuccess),
		StatKeyGraphSuccessTotal:            buildCountGraphStatusFunc(foresttype.GraphStatusSuccess),
		StatKeyGraphFailTotal:               buildCountGraphStatusFunc(foresttype.GraphStatusFailed),
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetDashboardOverview] BatchRegister fail, err: %v", err)
		return nil, err
	}
	statMap, err := mg.Execute(ctx, &StatQuery{
		CompanyID: req.Request.CompanyID,
		BeginAt:   time.Unix(req.Request.BeginAt, 0),
		EndAt:     time.Unix(req.Request.EndAt, 0),
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetDashboardOverview] Execute fail, err: %v", err)
		return nil, err
	}
	basicStatInfo := dtodashboard.BasicStatistics{}
	resourceStatInfo := dtodashboard.MultimodalStatistics{}
	graphStatInfo := dtodashboard.GraphStatistics{}
	for key, metric := range statMap {
		switch key {
		case StatKeyUserTotal:
			basicStatInfo.UserCount = metric.GetIntValue()
		case StatKeyNewUserTotal:
			basicStatInfo.NewUserCount = metric.GetIntValue()
		case StatKeyForestTotal:
			basicStatInfo.ForestCount = metric.GetIntValue()
		case StatKeyQATotal:
			basicStatInfo.QACount = metric.GetIntValue()
		case StatKeyAgentTotal:
			basicStatInfo.AgentCount = metric.GetIntValue()
		case StatKeyGraphTotal:
			basicStatInfo.GraphCount = metric.GetIntValue()
		case StatKeyMultimodalUploadFileTotal:
			resourceStatInfo.UploadCount = metric.GetIntValue()
		case StatKeyMultimodalReadyResourceTotal:
			resourceStatInfo.ReadyCount = metric.GetIntValue()
		case StatKeyFileParseSuccessTotal:
			resourceStatInfo.ParseSuccessCount = metric.GetIntValue()
		case StatKeyFileParseFailTotal:
			resourceStatInfo.ParseFailCount = metric.GetIntValue()
		case StatKeyFileParseAvgCost:
			resourceStatInfo.ParseAvgCost = metric.GetIntValue()
		case StatKeyFileIndexSuccessTotal:
			resourceStatInfo.IndexSuccessCount = metric.GetIntValue()
		case StatKeyFileIndexFailTotal:
			resourceStatInfo.IndexFailCount = metric.GetIntValue()
		case StatKeyFileIndexAvgCost:
			resourceStatInfo.IndexAvgCost = metric.GetIntValue()
		case StatKeyFileSummarySuccessTotal:
			resourceStatInfo.SummarySuccessCount = metric.GetIntValue()
		case StatKeyFileSummaryFailTotal:
			resourceStatInfo.SummaryFailCount = metric.GetIntValue()
		case StatKeyFileSummaryAvgCost:
			resourceStatInfo.SummaryAvgCost = metric.GetIntValue()
		case StatKeyGraphSuccessTotal:
			graphStatInfo.SuccessCount = metric.GetIntValue()
		case StatKeyGraphFailTotal:
			graphStatInfo.FailCount = metric.GetIntValue()
		}
	}
	resourceStatInfo.ParseSuccessRate = utils.Percentage(resourceStatInfo.ParseSuccessCount, resourceStatInfo.UploadCount, 2)
	resourceStatInfo.IndexSuccessRate = utils.Percentage(resourceStatInfo.IndexSuccessCount, resourceStatInfo.UploadCount, 2)
	resourceStatInfo.SummarySuccessRate = utils.Percentage(resourceStatInfo.SummarySuccessCount, resourceStatInfo.UploadCount, 2)
	graphStatInfo.SuccessRate = utils.Percentage(graphStatInfo.SuccessCount, graphStatInfo.SuccessCount+graphStatInfo.FailCount, 2)

	res.Response.BasicStatistics = basicStatInfo
	res.Response.MultimodalStatistics = resourceStatInfo
	res.Response.GraphStatistics = graphStatInfo

	return res, nil
}
