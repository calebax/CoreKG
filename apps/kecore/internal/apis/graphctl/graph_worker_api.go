package graphctl

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/coretask"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// ParseGraph 开始解析图谱
// @Tags 知识森林知识图谱
// @Summary 开始解析图谱
// @Description 开始解析图谱
// @Router /forest.ParseGraph [post]
// @Param user body ParseGraphRequest true "入参"
// @Success 200 {object} ParseGraphResponse "返回值"
func ParseGraph(ctx *gin.Context, req *ParseGraphRequest, resp *ParseGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ParseGraph.Validity failed: %s", resp.Message)
		return
	}
	// 查看图谱状态
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetGraph err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图谱信息失败
		return
	}
	// 加锁
	if graphInfo.Status != foresttype.GraphStatusUpdatable &&
		graphInfo.Status != foresttype.GraphStatusSuccess &&
		graphInfo.Status != foresttype.GraphStatusDraft {
		logs.ErrorContextf(ctx, "ParseGraph failed graph status is not draft")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_graph_not_draft" // 请勿重复操作图谱
		return
	}
	rdsKey := "graph:parse:graph_%d"
	if !redispool.GetLock(fmt.Sprintf(rdsKey, req.Request.GraphID), time.Minute) {
		logs.ErrorContextf(ctx, "ParseGraph failed graph is parsing")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_graph_parsing" // 图谱正在解析中，请稍后重试
		return
	}
	defer redispool.UnLock(fmt.Sprintf(rdsKey, req.Request.GraphID))

	// 获取上一个版本
	previousVersion, err := graph.GetPreviousVersion(ctx, graphInfo.ID, graphInfo.VersionID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.ErrorContextf(ctx, "ParseGraph failed GetPreviousVersion err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "内部错误"
		return
	}
	// 复制旧版本tag
	if previousVersion != nil {
		// 加完锁来创建所有tag
		pretags, err := graph.GetGraphTags(ctx, previousVersion.ID, previousVersion.VersionID)
		if err != nil {
			logs.ErrorContextf(ctx, "ParseGraph failed GetGraphTags previousVersion err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_get_tag_failed" // 获取图谱标签失败
			return
		}
		newtags := []*foresttype.GraphTag{}
		for _, v := range pretags {
			newtags = append(newtags, &foresttype.GraphTag{
				Uin:            v.Uin,
				CompanyID:      v.CompanyID,
				Description:    v.Description,
				GraphID:        v.GraphID,
				GraphVersionID: graphInfo.VersionID,
				TagName:        v.TagName,
				TagType:        v.TagType,
				Properties:     v.Properties,
			})
		}

		err = dbutil.Knownow().WithContext(ctx).CreateInBatches(newtags, 100).Error
		if err != nil {
			logs.ErrorContextf(ctx, "ParseGraph failed CreateInBatches err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "创建类型失败"
			return
		}
	}

	// 加完锁来创建所有tag
	tags, err := graph.GetGraphTags(ctx, graphInfo.ID, graphInfo.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "ParseGraph failed GetGraphTags err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_tag_failed" // 获取图谱标签失败
		return
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "")
	if err != nil {
		logs.ErrorContextf(ctx, "NewNebulaCLI error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_tag_failed" // 获取图谱标签失败
		return
	}
	defer cli.Release()
	err = cli.CheckSpaceExists(ctx, graphInfo.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "CheckSpaceExists error: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_tag_failed"
		return
	}
	db := dbutil.Knownow()
	for _, v := range tags {
		err = cli.CreateGraphTag(db, v)
		if err != nil {
			logs.ErrorContextf(ctx, "ParseGraph CreateGraphTag error: %v", err)
			continue
		}
	}
	// worker生成任务
	if err = coretask.GenerateForestGraphTask(ctx, graphInfo, true); err != nil {
		logs.ErrorContextf(ctx, "ParseGraph failed GenerateForestGraphTask err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_task_failed" // 生成任务失败
		return
	}
	forestInfo, err := forest.GetForestByID(ctx, graphInfo.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "ParseGraph failed GetForestByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_forest_failed" // 生成任务失败
		return
	}
	// 更新图谱状态
	graphInfo.Status = foresttype.GraphStatusPending
	if err := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = graph.UpdateGraph(ctx, graphInfo, tx); err != nil {
			logs.ErrorContextf(ctx, "ParseGraph failed UpdateGraphStatus err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_update_graph_failed" // 更新图谱状态失败
			return err
		}
		forestInfo.GraphStatus = foresttype.GraphStatusPending
		return tx.Save(forestInfo).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "ParseGraph UpdateGraph failed: %v", err)
		return
	}
	// 复制上一版本的手动创建节点到新版本
	tMap, err := graph.GetTagNameMapByGraphID(ctx, graphInfo.ID, graphInfo.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "ParseGraph GetTagNameMapByGraphID err: %v", err)
	} else {
		// 无err
		eMap, err := graph.GetEdgeNameMapByGraphID(ctx, graphInfo.ID, graphInfo.VersionID)
		if err != nil {
			logs.ErrorContextf(ctx, "ParseGraph GetEdgeNameMapByGraphID err: %v", err)
		} else {
			// 无err
			if err := graph.CopyPreviousVersionManualNodes(ctx, graphInfo, cli, tMap, eMap); err != nil {
				logs.ErrorContextf(ctx, "ParseGraph CopyPreviousVersionManualNodes err: %v", err)
			}
		}
	}
}

// RestockGraph 增量更新图谱
// @Tags 知识森林知识图谱
// @Summary 增量更新图谱
// @Description 增量更新图谱
// @Router /forest.RestockGraph [post]
// @Param user body ParseGraphRequest true "入参"
// @Success 200 {object} ParseGraphResponse "返回值"
func RestockGraph(ctx *gin.Context, req *ParseGraphRequest, resp *ParseGraphResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "ParseGraph.Validity failed: %s", resp.Message)
		return
	}
	// 查看图谱状态
	graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetGraph err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_graph_info_failed" // 获取图谱信息失败
		return
	}
	// 加锁
	if graphInfo.Status != foresttype.GraphStatusUpdatable &&
		graphInfo.Status != foresttype.GraphStatusSuccess {
		logs.ErrorContextf(ctx, "ParseGraph failed graph status is not draft")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_graph_not_draft" // 请勿重复操作图谱
		return
	}
	rdsKey := "graph:parse:graph_%d"
	if !redispool.GetLock(fmt.Sprintf(rdsKey, req.Request.GraphID), time.Minute) {
		logs.ErrorContextf(ctx, "ParseGraph failed graph is parsing")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_graph_parsing" // 图谱正在解析中，请稍后重试
		return
	}
	defer redispool.UnLock(fmt.Sprintf(rdsKey, req.Request.GraphID))
	// worker生成任务
	if err = coretask.GenerateForestGraphTask(ctx, graphInfo, false); err != nil {
		logs.ErrorContextf(ctx, "ParseGraph failed GenerateForestGraphTask err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_create_task_failed" // 生成任务失败
		return
	}
	forestInfo, err := forest.GetForestByID(ctx, graphInfo.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "ParseGraph failed GetForestByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kesearch_query_forest_failed" // 生成任务失败
		return
	}
	// 更新图谱状态
	graphInfo.Status = foresttype.GraphStatusRunning
	if err := dbutil.Knownow().Transaction(func(tx *gorm.DB) error {
		if err = graph.UpdateGraph(ctx, graphInfo, tx); err != nil {
			logs.ErrorContextf(ctx, "ParseGraph failed UpdateGraphStatus err: %v", err)
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "kecore_update_graph_failed" // 更新图谱状态失败
			return err
		}
		forestInfo.GraphStatus = foresttype.GraphStatusRunning
		return tx.Save(forestInfo).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "ParseGraph UpdateGraph failed: %v", err)
		return
	}
}

// GraphTaskCallback 获取知识森林词云图对应知识图谱
// @Tags 知识森林知识图谱
// @Summary 获取知识森林词云图对应知识图谱
// @Description 获取知识森林词云图对应知识图谱
// @Router /forest.GraphTaskCallback [post]
// @Param user body GraphTaskCallbackRequest true "入参"
// @Success 200 {object} GraphTaskCallbackResponse "返回值"
func GraphTaskCallback(ctx *gin.Context, req *GraphTaskCallbackRequest, resp *GraphTaskCallbackResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		logs.ErrorContextf(ctx, "GraphTaskCallback.Validity failed: %s", resp.Message)
		return
	}
	// // 查看图谱状态
	// graphInfo, err := graph.GetGraph(ctx, req.Request.GraphID)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetGraph err: %v", err)
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_get_graph_info_failed" // 获取图谱信息失败
	// 	return
	// }
	err := graph.GetLock(ctx, req.Request.GraphID)
	if err != nil {
		resp.Code = 503
		resp.Message = "kecore_task_running" // 该图谱有任务在运行中，请稍后重试
		logs.WarnContextf(ctx, "GraphTaskCallback.GetLock failed: %v", err)
		return
	}
	defer graph.UnLock(ctx, req.Request.GraphID)
	_, err = forest.GetForestFileByID(req.Request.FileID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.ErrorContextf(ctx, "GraphTaskCallBack GetForestFileByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_parse_failed" // 解析失败
		return
	}
	if err == gorm.ErrRecordNotFound {
		// 啥都不干
		return
	}
	wrapper, err := graph.NewParseAlgoWrapper(ctx, req.Request)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.WarnContextf(ctx, "[GraphTaskCallback] graph not found, id: %d", req.Request.GraphID)
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_graph_not_found" // 图谱不存在
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("kecore_parse_failed: %v", err) // 解析失败
		logs.ErrorContextf(ctx, "GraphTaskCallback.NewParseAlgoWrapper failed: %v", err)
		return
	}
	defer wrapper.Close()
	// 算法调用回调后处理结果
	err = wrapper.ParseAlgoResault()
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_parse_failed" // 解析失败
		logs.ErrorContextf(ctx, "GraphTaskCallback.ParseAlgoResault failed: %v", err)
		return
	}
	// // 临时逻辑，更新图谱状态
	// err = coretask.UpdateGraphStatus(ctx, graphInfo)
	// if err != nil {
	// 	resp.Code = errcode.ErrCode_InternalError
	// 	resp.Message = "kecore_update_graph_failed" // 更新图谱状态失败
	// 	logs.ErrorContextf(ctx, "GraphTaskCallback.UpdateGraphStatus failed: %v", err)
	// 	return
	// }
}
