package dtoproject

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/project"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type CreateProjectResponse struct {
	apiobj.BaseResponse
	Response CreateProjectEmbedResponse
}

type CreateProjectEmbedResponse struct {
	*foresttype.KnownowProject
}

type GetProjectInfoResponse struct {
	apiobj.BaseResponse
	Response GetProjectInfoEmbedResponse
}
type GetProjectInfoEmbedResponse struct {
	*project.ProjectInfo
}

type DeleteProjectResponse struct {
	apiobj.BaseResponse
	Response DeleteProjectEmbedResponse
}
type DeleteProjectEmbedResponse struct {
}

type RenameProjectResponse struct {
	apiobj.BaseResponse
	Response RenameProjectEmbedResponse
}
type RenameProjectEmbedResponse struct {
}

type ListProjectResponse struct {
	apiobj.BaseResponse
	Response ListProjectEmbedResponse
}
type ListProjectEmbedResponse struct {
	Project *project.ProjectInfoList `json:"project"`
}

type GetDefaultProjectResponse struct {
	apiobj.BaseResponse
	Response GetDefaultProjectEmbedResponse
}

type ProjItem struct {
	ProjectID uint `json:"project_id"`
	SessionID uint `json:"session_id"`
}

type GetDefaultProjectEmbedResponse struct {
	File  ProjItem `json:"file"`
	Agent ProjItem `json:"agent"`
}

type ListProjectItemResponse struct {
	apiobj.BaseResponse
	Response ListProjectItemEmbedResponse
}

type Item struct {
	// 项目ID
	ID uint `json:"id"`
	// 项目名称
	Name string `json:"name"`
	// 项目类型
	ProjectType foresttype.ProjectType `json:"project_type"`
}

type ListProjectItemEmbedResponse struct {
	apiobj.QueryResponse
	Data []*Item `json:"data"`
}

type GetProjectItemResponse struct {
	apiobj.BaseResponse
	Response GetProjectItemEmbedResponse
}
type GetProjectItemEmbedResponse struct {
	// 项目ID
	ID uint `json:"id"`
	// 项目名称
	Name string `json:"name"`
	// 项目类型
	ProjectType foresttype.ProjectType `json:"project_type"`
	// Forest
	Forest []*project.ProjectForest `json:"forest" gorm:"-"`
}
