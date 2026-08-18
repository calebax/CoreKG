package projecthandler

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

func CreateDefaultProject(ctx *gin.Context, companyID, uin uint) (map[foresttype.ProjectType]foresttype.KnownowProject, error) {
	existDefaultProjectList, err := forest.NewKeProjectDao().GetListByCond(ctx, &forest.KeProjectCond{
		Uin:             uin,
		ProjectTypeList: []foresttype.ProjectType{foresttype.ProjectTypeForestQA, foresttype.ProjectTypeAgentQA},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateDefaultProject] get exist default project list failed, uin: %d, err: %v", uin, err)
		return nil, err
	}
	projectTypeMap := make(map[foresttype.ProjectType]foresttype.KnownowProject)
	for _, v := range existDefaultProjectList {
		projectTypeMap[v.ProjectType] = v
	}
	var projectEntityList foresttype.KnownowProjectList
	if _, ok := projectTypeMap[foresttype.ProjectTypeForestQA]; !ok {
		projectEntityList = append(projectEntityList, foresttype.KnownowProject{
			Name:        i18n.T(runtime.GetLanguage(ctx), "kecore_default_forest_qa_project_name"),
			ProjectType: foresttype.ProjectTypeForestQA,
			Uin:         uin,
			CompanyID:   companyID,
			Sort:        foresttype.ProjectSortForestQA,
		})
	}
	if _, ok := projectTypeMap[foresttype.ProjectTypeAgentQA]; !ok {
		projectEntityList = append(projectEntityList, foresttype.KnownowProject{
			Name:        i18n.T(runtime.GetLanguage(ctx), "kecore_default_agent_qa_project_name"),
			ProjectType: foresttype.ProjectTypeAgentQA,
			Uin:         uin,
			CompanyID:   companyID,
			Sort:        foresttype.ProjectSortAgentQA,
		})
	}

	// 说明已经有了默认项目，直接返回
	if len(projectEntityList) == 0 {
		return projectTypeMap, nil
	}

	if err := forest.NewKeProjectDao().BatchInsert(ctx, projectEntityList); err != nil {
		logs.ErrorContextf(ctx, "[CreateDefaultProject] create project fail, uin:%d, err: %v", uin, err)
		return nil, err
	}
	for _, v := range projectEntityList {
		projectTypeMap[v.ProjectType] = v
	}
	return projectTypeMap, nil
}

// MigrateDefaultProject 迁移默认项目，暂时不做迁移，直接创建
func MigrateDefaultProject(ctx *gin.Context) error {
	return nil
}
