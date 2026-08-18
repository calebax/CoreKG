package dtoproject

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateProjectRequest struct {
	apiobj.BaseRequest
	Request CreateProjectEmbedRequest
}

type CreateProjectEmbedRequest struct {
	Name string `json:"name"`
}

func (opt *CreateProjectRequest) Validity(resp *CreateProjectResponse) {
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入项目名称"
		return
	}
}

type GetProjectInfoRequest struct {
	apiobj.BaseRequest
	Request GetProjectInfoEmbedRequest
}
type GetProjectInfoEmbedRequest struct {
	ID uint `json:"id"`
}

func (opt *GetProjectInfoRequest) Validity(resp *GetProjectInfoResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请选择项目"
		return
	}
}

type DeleteProjectRequest struct {
	apiobj.BaseRequest
	Request DeleteProjectEmbedRequest
}
type DeleteProjectEmbedRequest struct {
	ID         uint `json:"id"`
	MoveToFree bool `json:"move_to_free"`
}

func (opt *DeleteProjectRequest) Validity(resp *DeleteProjectResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
}

type RenameProjectRequest struct {
	apiobj.BaseRequest
	Request RenameProjectEmbedRequest
}
type RenameProjectEmbedRequest struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (opt *RenameProjectRequest) Validity(resp *RenameProjectResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请选择项目"
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请输入项目名称"
		return
	}
}

type ListProjectRequest struct {
	apiobj.BaseRequest
	Request ListProjectEmbedRequest
}
type ListProjectEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListProjectRequest) Validity(resp *ListProjectResponse) {
	if opt.Request.Offset < 0 || opt.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "offset和limit必须大于0"
		return
	}
	for _, v := range opt.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "name",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "orderBy不能为空"
			return
		}
	}
	for _, v := range opt.Request.Filters {
		switch v.Field {
		case "name", "created_at", "updated_at":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的字段只能有一个值"
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "查询条件中的值不能为空"
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "查询条件中的字段不存在, " + v.Field
			return
		}
	}
}

type GetDefaultProjectRequest struct {
	apiobj.BaseRequest
	Request GetDefaultProjectEmbedRequest
}
type GetDefaultProjectEmbedRequest struct {
	FileID  uint `json:"file_id"`
	AgentID uint `json:"agent_id"`
}

func (opt *GetDefaultProjectRequest) Validity(resp *GetDefaultProjectResponse) {

}

type ListProjectItemRequest struct {
	apiobj.BaseRequest
	Request ListProjectItemEmbedRequest
}
type ListProjectItemEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListProjectItemRequest) Validity(resp *ListProjectItemResponse) {
}

type GetProjectItemRequest struct {
	apiobj.BaseRequest
	Request GetProjectItemEmbedRequest
}
type GetProjectItemEmbedRequest struct {
	//项目项ID
	ID uint `json:"id"`
}

func (opt *GetProjectItemRequest) Validity(resp *GetProjectItemResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
}
