package svcproject

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoproject"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/project"
	"github.com/insmtx/corekg/apps/kecore/services/projecthandler"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// CreateProject 创建项目
func CreateProject(ctx *gin.Context, req *dtoproject.CreateProjectRequest) (res *dtoproject.CreateProjectResponse, err error) {
	res = &dtoproject.CreateProjectResponse{}
	companyID := runtime.CompanyID(ctx)
	uin := runtime.Uin(ctx)
	// milliTimestamp := time.Now().UnixMilli()
	// name := fmt.Sprintf("project-%d", milliTimestamp)
	// ex, err := project.ExistProject(ctx, companyID, name)
	// if err != nil {
	// 	res.Code = errcode.ErrCode_InternalError
	// 	res.Message = "查询重复项目失败"
	// 	return res, nil
	// }
	// if ex {
	// 	res.Code = errcode.ErrCode_InternalError
	// 	res.Message = "项目名称重复"
	// 	return res, nil
	// }
	projectEntity := &foresttype.KnownowProject{
		CompanyID: companyID,
		Uin:       uin,
		//Name:        i18n.T(runtime.GetLanguage(ctx), "kecore_default_project_name"),
		Name:        req.Request.Name,
		ProjectType: foresttype.ProjectTypeCustom,
	}
	if err := forest.NewKeProjectDao().Insert(ctx, projectEntity); err != nil {
		logs.ErrorContextf(ctx, "[CreateProject] insert project failed, err: %v", err)
		return nil, err
	}
	res.Response.KnownowProject = projectEntity
	return res, nil
}

// GetProjectInfo 获取项目信息
func GetProjectInfo(ctx *gin.Context, req *dtoproject.GetProjectInfoRequest) (res *dtoproject.GetProjectInfoResponse, err error) {
	res = &dtoproject.GetProjectInfoResponse{}
	proInfo, err := project.GetProjectInfoByID(ctx, req.Request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = errcode.ErrCode_NotFound
			res.Message = "项目不存在"
			return res, nil
		}
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取项目失败"
		return res, nil
	}
	res.Response.ProjectInfo = proInfo
	return res, nil
}

// DeleteProject 删除项目
func DeleteProject(ctx *gin.Context, req *dtoproject.DeleteProjectRequest) (res *dtoproject.DeleteProjectResponse, err error) {
	res = &dtoproject.DeleteProjectResponse{}
	err = dbutil.Knownow().Transaction(func(db *gorm.DB) error {
		// 删除项目
		err = project.DeleteProject(ctx, db, req.Request.ID)
		if err != nil {
			logs.ErrorContextf(ctx, "[DeleteProject] svcproject.DeleteProject(id:%v) failed, err: %v", req.Request.ID, err)
			return err
		}

		if !req.Request.MoveToFree {
			// 删除项目下的会话
			if err = project.DeleteProjectWithSessions(ctx, dbutil.Chat(), req.Request.ID); err != nil {
				logs.ErrorContextf(ctx, "[DeleteProject] svcproject.DeleteProjectWithSessions(id:%v) failed, err: %v", req.Request.ID, err)
				return err
			}
		} else {
			// 更新项目下的会话
			if err = project.UnsetSessionSubject(ctx, dbutil.Chat(), req.Request.ID); err != nil {
				logs.ErrorContextf(ctx, "[DeleteProject] svcproject.UnsetSessionSubject(id:%v) failed, err: %v", req.Request.ID, err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteProject] DeleteProject(id:%v) failed, err: %v", req.Request.ID, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_delete_project_failed"
		return res, nil
	}
	return res, nil
}

// RenameProject 项目重命名
func RenameProject(ctx *gin.Context, req *dtoproject.RenameProjectRequest) (res *dtoproject.RenameProjectResponse, err error) {
	res = &dtoproject.RenameProjectResponse{}
	pro, err := project.GetProjectByID(ctx, req.Request.ID)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取项目失败"
		return res, nil
	}

	if pro.ProjectType == foresttype.ProjectTypeAgentQA ||
		pro.ProjectType == foresttype.ProjectTypeForestQA {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "kecore_dont_modify_default_project"
		return res, nil
	}

	if pro.Name == req.Request.Name {
		return
	}
	// companyID := runtime.CompanyID(ctx)
	// ex, err := project.ExistProject(ctx, companyID, req.Request.Name)
	// if err != nil {
	// 	res.Code = errcode.ErrCode_InternalError
	// 	res.Message = "查询同名项目失败"
	// 	return res, nil
	// }
	// if ex {
	// 	res.Code = errcode.ErrCode_InternalError
	// 	res.Message = "项目名称重复"
	// 	return res, nil
	// }
	pro.Name = req.Request.Name
	err = project.UpdateProject(ctx, pro)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "更新项目失败"
		return res, nil
	}
	return res, nil
}

// ListProject 获取项目列表
func ListProject(ctx *gin.Context, req *dtoproject.ListProjectRequest) (res *dtoproject.ListProjectResponse, err error) {
	// TODO: 这里需要优化，逻辑需要重构，返回给前端的数据需要简化
	res = &dtoproject.ListProjectResponse{
		Response: dtoproject.ListProjectEmbedResponse{
			Project: &project.ProjectInfoList{},
		},
	}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	req.Request.Uin = runtime.Uin(ctx)

	projectInfoList, err := project.ListProject(ctx, req.Request.PageQuery)
	if err != nil {
		res.Code = errcode.ErrCode_InternalError
		res.Message = "获取项目列表失败"
		return res, nil
	}

	personalProjectTotal, err := forest.NewKeProjectDao().CountByCond(ctx, &forest.KeProjectCond{
		BaseCond: forest.BaseCond{
			OrderBy: req.Request.OrderBy,
		},
		CompanyID:       req.Request.CompanyID,
		Uin:             req.Request.Uin,
		ProjectTypeList: []foresttype.ProjectType{foresttype.ProjectTypeForestQA, foresttype.ProjectTypeAgentQA},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[ListProject] CountByCond failed, err: %v", err)
		return nil, err
	}

	personalProjectInfoList := make([]*project.ProjectInfo, 0)
	if personalProjectTotal == 0 {
		// 创建知识库问答和智能体问答项目
		projectMap, err := projecthandler.CreateDefaultProject(ctx, req.Request.CompanyID, req.Request.Uin)
		if err != nil {
			logs.ErrorContextf(ctx, "[ListProject] CreateDefaultProject failed, err: %v", err)
			return nil, err
		}
		forestQAProjectEntity := projectMap[foresttype.ProjectTypeForestQA]
		agentQAProjectEntity := projectMap[foresttype.ProjectTypeAgentQA]
		if forestQAProjectEntity.ID > 0 {
			personalProjectInfoList = append(personalProjectInfoList, &project.ProjectInfo{KnownowProject: forestQAProjectEntity})
		}
		if agentQAProjectEntity.ID > 0 {
			personalProjectInfoList = append(personalProjectInfoList, &project.ProjectInfo{KnownowProject: agentQAProjectEntity})
		}
	}

	if len(projectInfoList.Data) == 0 {
		projectInfoList.Data = personalProjectInfoList
		// 创建默认项目
		createRes, err := CreateProject(ctx, &dtoproject.CreateProjectRequest{})
		if err != nil || createRes.Code != 0 {
			res.Code = errcode.ErrCode_InternalError
			res.Message = "创建失败"
			return res, nil
		}
		projectInfoList.Data = append(projectInfoList.Data, &project.ProjectInfo{KnownowProject: *createRes.Response.KnownowProject})
		res.Response.Project.Total += int64(len(personalProjectInfoList))
	} else if len(personalProjectInfoList) > 0 {
		personalProjectInfoList = append(personalProjectInfoList, projectInfoList.Data...)
		projectInfoList.Data = personalProjectInfoList
		res.Response.Project.Total += int64(len(personalProjectInfoList))
	}
	res.Response.Project = projectInfoList
	return res, nil
}

func GetDefaultProject(ctx *gin.Context, req *dtoproject.GetDefaultProjectRequest) (res *dtoproject.GetDefaultProjectResponse, err error) {
	res = &dtoproject.GetDefaultProjectResponse{}

	var (
		uin = runtime.Uin(ctx)
	)

	projs, err := forest.NewKeProjectDao().GetListByCond(ctx, &forest.KeProjectCond{
		Uin:             uin,
		ProjectTypeList: []foresttype.ProjectType{foresttype.ProjectTypeForestQA, foresttype.ProjectTypeAgentQA},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GetDefaultProject] GetByCond failed, err: %v", err)
		return nil, err
	}

	for _, proj := range projs {
		if proj.ProjectType == foresttype.ProjectTypeForestQA {
			res.Response.File.ProjectID = proj.ID
		}
		if proj.ProjectType == foresttype.ProjectTypeAgentQA {
			res.Response.Agent.ProjectID = proj.ID
		}
	}

	if req.Request.FileID > 0 {
		sf, err := chat.NewChatSessionsDao().GetByCond(ctx, &chat.ChatSessionsCond{
			FileID:     req.Request.FileID,
			SubjectIDs: []uint{res.Response.File.ProjectID},
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[GetDefaultProject] GetByCond failed, err: %v", err)
			return res, err
		}
		if sf.ID == 0 {
			logs.ErrorContextf(ctx, "GetByCond failed, no session found")
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_get_session_failed"
			return res, nil
		}

		res.Response.File.SessionID = sf.ID
	}
	if req.Request.AgentID > 0 {
		sg, err := chat.NewChatSessionsDao().GetByCond(ctx, &chat.ChatSessionsCond{
			AgentID:    req.Request.AgentID,
			SubjectIDs: []uint{res.Response.Agent.ProjectID},
		})
		if err != nil {
			logs.ErrorContextf(ctx, "[GetDefaultProject] GetByCond failed, err: %v", err)
			return res, err
		}

		if sg.ID == 0 {
			logs.ErrorContextf(ctx, "GetByCond failed, no session found")
			res.Code = errcode.ErrCode_InternalError
			res.Message = "kecore_get_session_failed"
			return res, nil
		}

		res.Response.Agent.SessionID = sg.ID

	}

	return res, nil
}

func ListProjectItem(ctx *gin.Context, req *dtoproject.ListProjectItemRequest) (res *dtoproject.ListProjectItemResponse, err error) {
	res = &dtoproject.ListProjectItemResponse{}
	req.Request.CompanyID = runtime.CompanyID(ctx)
	req.Request.Uin = runtime.Uin(ctx)

	cond := &forest.KeProjectCond{
		BaseCond: forest.BaseCond{
			Offset:  req.Request.Offset,
			Limit:   req.Request.Limit,
			OrderBy: req.Request.OrderBy,
		},
		CompanyID: req.Request.CompanyID,
		Uin:       req.Request.Uin,
	}

	for _, filter := range req.Request.Filters {
		switch filter.Field {
		case "name":
			cond.NameLike = filter.Value[0]
		default:
			logs.ErrorContextf(ctx, "ListProjectItem invalid filtering field: %s", filter.Field)
			return res, fmt.Errorf("invalid filtering field: %s", filter.Field)
		}
	}

	projs, total, err := forest.NewKeProjectDao().GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListProjectItem] GetPageListByCond failed, err: %v", err)
		return
	}

	itemList := make([]*dtoproject.Item, 0)
	if total == 0 {
		// 创建知识库问答和智能体问答项目
		projectMap, err2 := projecthandler.CreateDefaultProject(ctx, req.Request.CompanyID, req.Request.Uin)
		if err2 != nil {
			logs.ErrorContextf(ctx, "[ListProjectItem] CreateDefaultProject failed, err: %v", err)
			return res, err2
		}
		forestQAProjectEntity := projectMap[foresttype.ProjectTypeForestQA]
		agentQAProjectEntity := projectMap[foresttype.ProjectTypeAgentQA]
		if forestQAProjectEntity.ID > 0 {
			itemList = append(itemList, &dtoproject.Item{
				ID:          forestQAProjectEntity.ID,
				Name:        forestQAProjectEntity.Name,
				ProjectType: foresttype.ProjectTypeForestQA,
			})
		}
		if agentQAProjectEntity.ID > 0 {
			itemList = append(itemList, &dtoproject.Item{
				ID:          agentQAProjectEntity.ID,
				Name:        agentQAProjectEntity.Name,
				ProjectType: foresttype.ProjectTypeAgentQA,
			})
		}
	}
	res.Response.Limit = req.Request.Limit
	res.Response.Total = total

	if len(projs) > 0 {
		for _, v := range projs {
			itemList = append(itemList, &dtoproject.Item{
				ID:          v.ID,
				Name:        v.Name,
				ProjectType: v.ProjectType,
			})
		}
		res.Response.Total += int64(len(itemList))
	}
	res.Response.Data = itemList
	return
}

func GetProjectItem(ctx *gin.Context, req *dtoproject.GetProjectItemRequest) (res *dtoproject.GetProjectItemResponse, err error) {
	res = &dtoproject.GetProjectItemResponse{}
	proInfo, err := project.GetProjectInfoByID(ctx, req.Request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = errcode.ErrCode_NotFound
			res.Message = "kecore_project_not_found"
			return
		}
		res.Code = errcode.ErrCode_InternalError
		res.Message = "kecore_query_project_failed"
		return
	}
	res.Response.ID = proInfo.ID
	res.Response.Name = proInfo.Name
	res.Response.ProjectType = proInfo.ProjectType

	res.Response.Forest = proInfo.Forest
	return
}
