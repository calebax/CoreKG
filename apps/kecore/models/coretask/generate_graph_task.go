package coretask

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// GenerateForestGraphTask 生成图谱任务
func GenerateForestGraphTask(ctx context.Context, graphInof *foresttype.ForestGraphInfo, full bool) error {
	// 制作payload
	payload := graph.GraphAlgoReq{
		CommonPayload: task.CommonPayload{
			TaskType: GraphFileTask,
			Timeout:  int64(TaskTimeout)},
		GraphID: graphInof.ID,
		// TODO 动态获取es索引
		EsIndex:        "ke_0",
		Mode:           graphInof.ParseMode,
		IsIgnoreStatus: full,
	}
	tMap, err := graph.GetTagIDMapByGraphID(ctx, graphInof.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateForestGraphTask GetTagIDMapByGraphID err: %v", err)
		return err
	}
	eMap, err := graph.GetEdgeIDMapByGraphID(ctx, graphInof.ID, graphInof.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateForestGraphTask GetEdgeIDMapByGraphID err: %v", err)
		return err
	}
	etList, err := graph.ListEdgeTag(ctx, graphInof.ID, graphInof.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateForestGraphTask ListEdgeTag err: %v", err)
		return err
	}
	for _, v := range tMap {
		t := graph.Tag{
			TagName:    v.TagName,
			Properties: v.Properties,
			Comment:    v.Description,
		}
		payload.Tags = append(payload.Tags, t)
	}
	for _, v := range etList {
		if _, ok := eMap[v.EdgeTypeID]; !ok {
			continue
		}
		if _, ok := tMap[v.SrcTagID]; !ok {
			continue
		}
		if _, ok := tMap[v.DstTagID]; !ok {
			continue
		}
		e := graph.Edge{
			EdgeName:   eMap[v.EdgeTypeID].TagName,
			SrcTagName: tMap[v.SrcTagID].TagName,
			DstTagName: tMap[v.DstTagID].TagName,
		}
		payload.Edges = append(payload.Edges, e)

	}

	var taskList []*task.Task
	files, err := forest.ListAllForestFile(graphInof.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateForestGraphTask ListAllForestFile err: %v", err)
		return err
	}
	for _, v := range files {
		if v.IsDir.Value() {
			continue
		}
		if !full {
			// 图谱生成过了无需再生成
			if v.GraphStatus == foresttype.TaskStatusSuccess {
				continue
			}
		}
		payload.FileID = v.ID
		// 生成任务
		tsk := &task.Task{
			Uin:               graphInof.Uin,
			CompanyID:         graphInof.CompanyID,
			TaskType:          GraphFileTask,
			TaskStatus:        task.TaskStatusPending,
			SubjectID:         graphInof.VersionID,
			Comment:           "图谱生成任务",
			Payload:           payload.String(),
			TaskConfigTimeout: TaskTimeout,
			AppGroup:          GraphAppGroup,
			TaskConfigRedo:    TaskRedo,
			Priority:          TaskPriority,
		}
		taskList = append(taskList, tsk)
	}
	if len(taskList) == 0 {
		logs.ErrorContextf(ctx, "GenerateForestGraphTask taskList is empty")
		return fmt.Errorf("GenerateForestGraphTask taskList is empty")
	}
	// 批量插入
	if err := dbutil.Core().WithContext(ctx).CreateInBatches(taskList, 100).Error; err != nil {
		logs.ErrorContextf(ctx, "GenerateForestGraphTask error %v", err)
		return err
	}
	return nil
}

// UpdateGraphStatus 修改图谱状态和知识库状态
func UpdateGraphStatus(ctx context.Context, graphInof *foresttype.ForestGraphInfo) error {
	forestInfo, err := forest.GetForestByID(ctx, graphInof.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateGraphStatus failed GetForestByID err: %v", err)
		return err
	}
	defer func() {
		err = dbutil.Knownow().Save(forestInfo).Error
		if err != nil {
			logs.ErrorContextf(ctx, "GraphTaskCallback.UpdateforestStatus failed: %v", err)
			return
		}
	}()
	// 查看是否有可执行的任务
	candoTask, err := FindCanDoTaskBySubjectID(ctx, graphInof.VersionID, GraphAppGroup)
	if err != nil {
		logs.ErrorContextf(ctx, "GraphTaskCallback.FindCanDoTaskBySubjectID failed: %v", err)
		return err
	}
	if len(candoTask) > 1 {
		err := graph.UpdateGraphStatus(ctx, graphInof.ID, foresttype.GraphStatusRunning)
		if err != nil {
			logs.ErrorContextf(ctx, "GraphTaskCallback.UpdateGraphStatus failed: %v", err)
			return err
		}
		forestInfo.GraphStatus = foresttype.GraphStatusRunning
		return nil
	}

	// 临时逻辑，更新图谱状态
	tsks, err := FindPendingTaskBySubjectID(ctx, graphInof.VersionID, GraphAppGroup)
	if err != nil {
		logs.ErrorContextf(ctx, "GraphTaskCallback.FindPendingTaskBySubjectID failed: %v", err)
		return err
	}
	if len(tsks) == 1 && graphInof.Status != foresttype.GraphStatusFailed {
		// 说明当前是最后一个任务
		err := graph.UpdateGraphStatus(ctx, graphInof.ID, foresttype.GraphStatusSuccess)
		if err != nil {
			logs.ErrorContextf(ctx, "GraphTaskCallback.UpdateGraphStatus failed: %v", err)
			return err
		}
		forestInfo.GraphStatus = foresttype.GraphStatusSuccess
	} else if tsks[0].TaskStatus == task.TaskStatusFail && tsks[0].Redo == tsks[0].TaskConfigRedo {
		// forestInfo.GraphStatus = foresttype.GraphStatusFailed
		forestInfo.GraphStatus = foresttype.GraphStatusSuccess
		// TODO 测试效果
		err := graph.UpdateGraphStatus(ctx, graphInof.ID, foresttype.GraphStatusSuccess)
		if err != nil {
			logs.ErrorContextf(ctx, "GraphTaskCallback.UpdateGraphStatus failed: %v", err)
			return err
		}
	}
	return nil
}

// GraphTaskCallBack 图谱任务回调函数
func GraphTaskCallBack(ctx context.Context, tsk *task.Task) error {
	graphInfo, err := graph.GetGraphWithVersionID(ctx, tsk.SubjectID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetGraph err: %v", err)
		return err
	}
	// 更新文件状态
	payload := graph.GraphAlgoReq{}
	err = json.Unmarshal([]byte(tsk.Payload), &payload)
	if err != nil {
		logs.ErrorContextf(ctx, "GraphTaskCallBack json.Unmarshal err: %v", err)
		return err
	}
	payload.IsIgnoreStatus = false
	tsk.Payload = payload.String()
	file, err := forest.GetForestFileByID(payload.FileID)
	if err != nil && err != gorm.ErrRecordNotFound {
		logs.ErrorContextf(ctx, "GraphTaskCallBack GetForestFileByID err: %v", err)
		return err
	}
	if err == gorm.ErrRecordNotFound {
		// 更新payload状态
		tsk.TaskStatus = task.TaskStatusSuccess
		return UpdateGraphStatus(ctx, graphInfo)
	}
	defer func() {
		err = dbutil.Knownow().Save(file).Error
		if err != nil {
			logs.ErrorContextf(ctx, "GraphTaskCallback.Savefile failed: %v", err)
			return
		}
	}()
	switch tsk.TaskStatus {
	case task.TaskStatusFail:
		file.GraphStatus = foresttype.TaskStatusFail
		graphInfo.Status = foresttype.GraphStatusFailed
	case task.TaskStatusSuccess:
		file.GraphStatus = foresttype.TaskStatusSuccess
	}
	// 更新payload状态
	return UpdateGraphStatus(ctx, graphInfo)
}

// DeleteTasksByGraphVersion 删除图谱任务
func DeleteTasksByGraphVersion(ctx context.Context, graphVersion uint) error {
	if err := dbutil.Core().WithContext(ctx).Table("core_task ct").
		Where("ct.deleted_at IS NULL").
		Where("app_group = ?", GraphAppGroup).
		Where("ct.subject_id = ?", graphVersion).
		Delete(&task.Task{}).
		Error; err != nil {
		return err
	}
	return nil
}

type GraphTaskCount struct {
	SubjectID    uint  `gorm:"column:subject_id"`
	SuccessCount int64 `gorm:"column:success_count"`
	Count        int64 `gorm:"column:count"`
}

// GetGraphTaskCount 获取图谱任务数量
func GetGraphTaskCount(ctx context.Context, graphVersionID uint) (*GraphTaskCount, error) {
	var result GraphTaskCount

	err := dbutil.Core().WithContext(ctx).
		Table(task.Task{}.TableName()).
		Select(`
            subject_id,
            COUNT(*) AS count,
            SUM(CASE WHEN task_status = 'success' THEN 1 ELSE 0 END) AS success_count
        `).
		Where("subject_id = ?", graphVersionID).
		Where("task_type = ?", GraphFileTask).
		Group("subject_id").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListGraphTaskCount 获取图谱任务数量
func ListGraphTaskCount(ctx context.Context, graphVersionIDs []uint) (map[uint]*GraphTaskCount, error) {
	var result []*GraphTaskCount

	err := dbutil.Core().WithContext(ctx).
		Table(task.Task{}.TableName()).
		Select(`
            subject_id,
            COUNT(*) AS count,
            SUM(CASE WHEN task_status = 'success' THEN 1 ELSE 0 END) AS success_count
        `).
		Where("subject_id in ?", graphVersionIDs).
		Where("task_type = ?", GraphFileTask).
		Group("subject_id").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}
	resultMap := make(map[uint]*GraphTaskCount)
	for _, v := range result {
		resultMap[v.SubjectID] = v
	}
	return resultMap, nil
}
