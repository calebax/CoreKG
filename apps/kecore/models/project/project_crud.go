package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateProject 创建项目
func CreateProject(ctx context.Context, project *foresttype.KnownowProject) error {
	err := dbutil.Knownow().WithContext(ctx).Create(project).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateProject error: %v", err)
		return err
	}
	return nil
}

// GetProjectByID 根据ID获取项目
func GetProjectByID(ctx context.Context, id uint) (*foresttype.KnownowProject, error) {
	project := &foresttype.KnownowProject{}
	err := dbutil.Knownow().WithContext(ctx).Where("id = ?", id).First(project).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetProjectByID error: %v", err)
		return nil, err
	}
	return project, nil
}

// GetProjectByID 根据ID获取项目
func GetProjectInfoByID(ctx context.Context, id uint) (*ProjectInfo, error) {
	project := &foresttype.KnownowProject{}
	err := dbutil.Knownow().WithContext(ctx).Where("id = ?", id).First(project).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetProjectByID error: %v", err)
		return nil, err
	}
	projectInfo := &ProjectInfo{
		KnownowProject: *project,
	}
	forests, err := forest.GetForestByIDS(project.ForestIDList.Slice())
	if err != nil {
		logs.ErrorContextf(ctx, "GetForestByIDS error: %v", err)
		return nil, err
	}
	for _, v := range forests {
		projectInfo.Forest = append(projectInfo.Forest, &ProjectForest{
			ID:         v.ID,
			Name:       v.Name,
			ForestType: v.ForestType,
		})
	}
	return projectInfo, nil
}

// UpdateProject 更新项目
func UpdateProject(ctx context.Context, project *foresttype.KnownowProject) error {
	err := dbutil.Knownow().WithContext(ctx).Save(project).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateProject error: %v", err)
		return err
	}
	return nil
}

// ExistProject 检查项目是否存在
func ExistProject(ctx context.Context, companyID uint, name string) (bool, error) {
	var count int64
	err := dbutil.Knownow().WithContext(ctx).
		Model(&foresttype.KnownowProject{}).Where("company_id = ?", companyID).
		Where("name = ?", name).Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "ExistProject error: %v", err)
		return false, err
	}
	return count > 0, nil
}

// ListProject
func ListProject(ctx context.Context, opt apiobj.PageQuery) (*ProjectInfoList, error) {
	queryList := &ProjectInfoList{}
	sql := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.KnownowProject{}.TableName()+" AS p").
		Where("p.deleted_at IS NULL").
		Where("p.uin = ?", opt.Uin)
	for _, filter := range opt.Filters {
		switch filter.Field {
		// case "uin":
		// 	sql = sql.Where("p.uin = ?", filter.Value[0])
		case "name":
			sql = sql.Where("p.name LIKE ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		default:
			logs.ErrorContextf(ctx, "ListProject invalid filtering field: %s", filter.Field)
			return nil, fmt.Errorf("invalid filtering field: %s", filter.Field)
		}
	}
	// 处理 BeginTime 和 EndTime
	if !opt.BeginTime.IsZero() {
		sql = sql.Where("p.created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		sql = sql.Where("p.created_at <= ?", opt.EndTime)
	}

	if err := sql.Count(&queryList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "ListProject Statistical project failed: %v", err)
		return nil, err
	}
	if queryList.Total == 0 {
		return queryList, nil
	}
	// 排序逻辑
	if len(opt.OrderBy) > 0 {
		// 用户自定义的排序优先级在 sort 之后
		sql = sql.Order("p.sort DESC").Order(strings.Join(opt.OrderBy, ","))
	} else {
		// 默认排序：sort 排第一
		sql = sql.Order("p.sort DESC")
	}
	sql = sql.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		sql = sql.Limit(opt.Limit)
	}
	if err := sql.Find(&queryList.Data).Error; err != nil {
		logs.ErrorContextf(ctx, "ListProject Retrieval project failed: %v", err)
		return nil, err
	}
	queryList.Limit = opt.Limit
	queryList.Offset = opt.Offset

	return queryList, nil
}

// ProjectInfoList
type ProjectInfoList struct {
	apiobj.QueryResponse
	Data []*ProjectInfo `json:"data"`
}

// ProjectInfo ProjectInfo
type ProjectInfo struct {
	foresttype.KnownowProject
	Forest []*ProjectForest `json:"forest" gorm:"-"`
}

type ProjectForest struct {
	ID         uint                  `json:"id"`
	Name       string                `json:"name"`
	ForestType foresttype.ForestType `json:"forest_type"`
}

// GetProjectList 获取项目列表
func GetProjectList(ctx context.Context, companyID uint) ([]*foresttype.KnownowProject, error) {
	projectList := make([]*foresttype.KnownowProject, 0)
	err := dbutil.Knownow().WithContext(ctx).
		Where("company_id = ?", companyID).
		Find(&projectList).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetProjectList error: %v", err)
		return nil, err
	}
	return projectList, nil
}

// DeleteProject 删除项目
func DeleteProject(ctx context.Context, tx *gorm.DB, id uint) error {
	err := tx.WithContext(ctx).
		Where("id = ?", id).
		Where("project_type NOT IN (?)", []foresttype.ProjectType{foresttype.ProjectTypeAgentQA, foresttype.ProjectTypeForestQA}).
		Delete(&foresttype.KnownowProject{}).Error
	if err != nil {
		logs.WarnContextf(ctx, "DeleteProject error: %v", err)
		return err
	}
	return nil
}

func DeleteProjectWithSessions(ctx context.Context, tx *gorm.DB, id uint) error {
	err := tx.WithContext(ctx).
		Where("subject_id = ?", id).
		Delete(&chattype.ChatSession{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteProjectWithSessions(id:%v) error: %v", id, err)
		return err
	}
	return nil
}

// UnsetSessionSubject will unset session's project id ref
func UnsetSessionSubject(ctx context.Context, tx *gorm.DB, projectID uint) error {
	err := tx.WithContext(ctx).
		Table(chattype.TableNameChatSessions).
		Where("subject_id = ?", projectID).
		Where("deleted_at IS NULL").
		Updates(map[string]interface{}{
			"subject_id": 0,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UnsetSessionSubject(projectID:%v) error: %v", projectID, err)
		return err
	}
	return nil
}
