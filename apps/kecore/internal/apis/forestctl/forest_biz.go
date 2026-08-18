package forestctl

import (
	"path"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		Name              string                             `json:"name"`
		AvatarUrl         string                             `json:"avatar_url"`
		Decription        string                             `json:"description"`
		PublicScope       foresttype.PublicScope             `json:"public_scope"`
		ForestType        foresttype.ForestType              `json:"forest_type"`         // 知识库类型，folder：文件夹(目录)，standard：标准类型(其他非数据库的知识库)，excel：excel 文件，db：数据库
		DataSourceType    foresttype.ForestDataSourceType    `json:"data_source_type"`    // 数据源类型，standard：标准类型(其他非数据库的知识库)，excel：excel 文件，db：数据库
		DataSourceSubtype foresttype.ForestDataSourceSubtype `json:"data_source_subtype"` // 数据源子类型，standard：标准类型(其他非数据库的知识库)，excel：excel 文件，mysql：mysql
	}
}

func (opt *CreateForestRequest) Validity(resp *CreateForestResponse) {
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty" // 名称不能为空
		return
	}
	if opt.Request.PublicScope == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_public_scope_empty" // 公开范围不能为空
		return
	}
	if opt.Request.ForestType == "" {
		opt.Request.ForestType = foresttype.ForestTypeFile
	}

	if opt.Request.DataSourceType == "" {
		opt.Request.DataSourceType = foresttype.ForestDataSourceStandard
		opt.Request.DataSourceSubtype = foresttype.ForestDataSourceSubtypeStandard
	}
	if _, ok := foresttype.ForestDataSourceTypeMap[opt.Request.DataSourceType]; !ok {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_data_source_type" // 未知数据源类型
		return
	}
	if _, ok := foresttype.ForestDataSourceSubtypeMap[opt.Request.DataSourceSubtype]; !ok {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_data_source_subtype" // 未知数据源子类型
		return
	}

	switch opt.Request.ForestType {
	case foresttype.ForestTypeFile, foresttype.ForestTypeCAD,
		foresttype.ForestTypeData, foresttype.ForestTypeQA:
		return
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_unknown_forest_type" // 未知知识库类型
		return
	}
}

type CreateForestResponse struct {
	apiobj.BaseResponse
	Response struct {
		ForestID uint `json:"forest_id"`
	}
}

type ListForestRequest struct {
	apiobj.BaseRequest
	Request apiobj.PageQuery
}

func (req *ListForestRequest) Validity(resp *apiobj.BaseResponse) {
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "created_at", "updated_at", "name", "knowledge_status",
			"created_at desc", "updated_at desc":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_order_by_empty" // orderBy不能为空
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "auto", "name", "created_at", "updated_at", "knowledge_status":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_value_invalid" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_value_empty" // 查询条件中的值不能为空
				return
			}
		case "forest_type":
			if len(v.Value) == 0 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_field_empty" // 查询条件中的字段不能空
				return
			}
			for _, forestType := range v.Value {
				if forestType == "" {
					resp.Code = errcode.ErrCode_BadRequest
					resp.Message = "kecore_filter_value_empty" // 查询条件中的值不能为空
					return
				}
			}

		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_filter_field_invalid, " + v.Field // 查询条件中的字段不存在
			return
		}
	}
}

type ListForestResponse struct {
	apiobj.BaseResponse
	Response forest.ForestInfoItemList
}

type ModifyForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID                uint                               `json:"id"`
		Name              string                             `json:"name"`
		AvatarUrl         string                             `json:"avatar_url"`
		Decription        string                             `json:"description"`
		ForestType        foresttype.ForestType              `json:"forest_type"`
		DataSourceType    foresttype.ForestDataSourceType    `json:"data_source_type"`    // 数据源类型，standard：标准类型(其他非数据库的知识库)，excel：excel 文件，db：数据库
		DataSourceSubtype foresttype.ForestDataSourceSubtype `json:"data_source_subtype"` // 数据源子类型，standard：标准类型(其他非数据库的知识库)，excel：excel 文件，mysql：mysql
	}
}

func (opt *ModifyForestRequest) Validity(resp *ModifyForestResponse) {
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_new_name" // 请输入新名称
		return
	}
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_to_modify" // 请选择要修改的知识森林
		return
	}
}

type ModifyForestResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type GetForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

func (opt *GetForestRequest) Validity(resp *GetForestResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_to_modify" // 请选择要修改的知识森林
		return
	}
}

type GetForestResponse struct {
	apiobj.BaseResponse
	Response struct {
		Data      *forest.ForestInfoItemstruct `json:"data"`
		GraphInfo *foresttype.ForestGraphInfo  `json:"graph_info"`
	}
}

type DeleteForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	}
}

func (opt *DeleteForestRequest) Validity(resp *DeleteForestResponse) {
	if opt.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest_to_delete" // 请选择要删除的知识森林
		return
	}
}

type DeleteForestResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type ListFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		apiobj.PageQuery
		ForestID uint   `json:"forest_id"`
		ImageUrl string `json:"image_url"`
	}
}

func (req *ListFileRequest) Validity(resp *ListFileResponse) {
	if req.Request.ImageUrl == "" {
		return
	}
	if req.Request.Offset < 0 || req.Request.Limit < 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_offset_limit_invalid" // offset和limit必须大于0
		return
	}
	for _, v := range req.Request.OrderBy {
		switch v {
		case "id desc", "id", "created_at desc", "created_at", "size desc", "size", "name desc", "name", "ext desc", "ext":
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_order_by_missing" // 未包含orderBy条件
			return
		}
	}
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "parent_id", "forest_id", "name", "parse_status", "is_dir", "enable":
			if len(v.Value) != 1 {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_value_invalid" // 查询条件中的字段只能有一个值
				return
			}
			if v.Value[0] == "" {
				resp.Code = errcode.ErrCode_BadRequest
				resp.Message = "kecore_filter_value_empty" // 查询条件中的值不能为空
				return
			}
		default:
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kecore_filter_field_invalid, " + v.Field // 查询条件中的字段不存在
			return
		}
	}
}

type ListFileResponse forest.QueryForestFileResponse

type CreateDirRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		ParentID uint   `json:"parent_id"`
		Name     string `json:"name"`
	}
}

func (opt *CreateDirRequest) Validity(resp *CreateDirResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if opt.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_directory_name" // 请输入文件夹名称
		return
	}
}

type CreateDirResponse struct {
	apiobj.BaseResponse
	Response struct {
		ID uint `json:"id"`
	}
}

type DeleteDirRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileIDs []uint `json:"file_id"`
	}
}

func (opt *DeleteDirRequest) Validity(resp *DeleteDirResponse) {
	if len(opt.Request.FileIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty" // 文件id为空
		return
	}
}

type DeleteDirResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type UploadFileRequest struct {
	apiobj.BaseRequest
	Request struct{}
}

type UploadFileResponse struct {
	apiobj.BaseResponse
	Response UploadFileEmbeddedResponse
}

type UploadFileEmbeddedResponse struct {
	// ForestFileID 知识库文件 ID
	ForestFileID uint `json:"forest_file_id"`
}

type RenamePathRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID  uint   `json:"file_id"`
		NewName string `json:"new_name"`
	}
}

func (opt *RenamePathRequest) Validity(resp *RenamePathResponse) {
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if opt.Request.NewName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_new_name" // 请输入新的名称
		return
	}
}

type RenamePathResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type MovePathRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID      uint `json:"file_id"`
		DstParentID uint `json:"dst_parent_id"`
	}
}

func (opt *MovePathRequest) Validity(resp *MovePathResponse) {
	if opt.Request.FileID == opt.Request.DstParentID {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_move_to_same_directory" // 请勿移动到当前目录
		return
	}
}

type MovePathResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type PreviewFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint   `json:"forest_id"`
		Path     string `json:"path"`
	}
}

func (opt *PreviewFileRequest) Validity(resp *PreviewFileResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if opt.Request.Path == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_file" // 请选择文件
		return
	}
	if path.Ext(opt.Request.Path) != ".pdf" && path.Ext(opt.Request.Path) != ".PDF" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_not_previewable" // 仅支持pdf文件预览
		return
	}
}

type PreviewFileResponse struct {
	apiobj.BaseResponse
	Response struct{}
}

type GetFileInfoRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

func (opt *GetFileInfoRequest) Validity(resp *GetFileInfoResponse) {
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty" // 文件id为空
		return
	}
}

type GetFileInfoResponse struct {
	apiobj.BaseResponse
	Response struct {
		*foresttype.KnownowForestFile
		Forest        *foresttype.KnownowForest
		ParentIDArr   []uint   `json:"parent_id_arr"`
		ParentPathArr []string `json:"parent_path_arr"`
	}
}

type PreviewFileByURLRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID     uint `json:"file_id"`
		IsDownLoad bool `json:"is_download"`
	}
}

func (opt *PreviewFileByURLRequest) Validity(resp *PreviewFileByURLResponse) {
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty"
		return
	}
}

type PreviewFileByURLResponse struct {
	apiobj.BaseResponse
	Response struct {
		URL string `json:"url"`
	}
}

type RecentlyForestRequest struct {
	apiobj.BaseRequest
	Request struct {
	}
}

func (opt *RecentlyForestRequest) Validity(resp *RecentlyForestResponse) {

}

type RecentlyForestResponse struct {
	apiobj.BaseResponse
	Response struct {
		ForestCount int64                       `json:"forest_count"`
		FileCount   int64                       `json:"file_count"`
		Forests     []*foresttype.KnownowForest `json:"forest"`
		CreateCount int64                       `json:"create_count"`
	}
}

type GetFilePathRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint `json:"forest_id"`
		FileID   uint `json:"file_id"`
	}
}
type GetFilePathResponse struct {
	apiobj.BaseResponse
	Response struct {
		PathIds     []uint   `json:"path_ids"`
		PathStrings []string `json:"path_strings"`
	}
}

func (opt *GetFilePathRequest) Validity(resp *GetFilePathResponse) {
	if opt.Request.ForestID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty" // 森林id为空
		return
	}
	if opt.Request.FileID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty" // 文件id为空
		return
	}
}

type DeleteFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		FileID uint `json:"file_id"`
	}
}

type DeleteFileResponse struct {
	apiobj.BaseResponse
}

func (opt *DeleteFileRequest) Validity(resp *DeleteFileResponse) {
	if opt.Request.FileID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty"
		return
	}
}

type ListExcelSheetRequest struct {
	apiobj.BaseRequest
	Request ListExcelSheetEmbedRequest
}

type ListExcelSheetEmbedRequest struct {
	ForestFileIDs []uint `json:"forest_file_ids" validate:"required,gt=0"` // 知识库文件 id 列表
}

func (opt *ListExcelSheetRequest) Validity(resp *ListExcelSheetResponse) {
	if len(opt.Request.ForestFileIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty"
		return
	}
}

type ListExcelSheetResponse struct {
	apiobj.BaseResponse
	Response ListExcelSheetEmbedResponse
}

type ListExcelSheetEmbedResponse struct {
	SheetList []ExcelSheetListItem `json:"sheet_list"` // excel sheet 的列表
}

type ExcelSheetListItem struct {
	ForestFileID uint   `json:"forest_file_id"` // excel sheet 的所属 forest file id
	ExcelSheetID uint   `json:"excel_sheet_id"` // excel sheet 的 id
	SheetName    string `json:"sheet_name"`     // sheet 名称
}

type GetNameByModuleIDsRequest struct {
	apiobj.BaseRequest
	Request GetNameByModuleIDsEmbedRequest
}

type GetNameByModuleIDsEmbedRequest struct {
	// ModuleIDList 模块 id 列表
	ModuleIDList []ModuleIDListItem `json:"module_id_list" validate:"required,gt=0"`
}

type ModuleIDListItem struct {
	// Module 模块，forest：知识库，database：数据库，table：表
	Module foresttype.ForestModule `json:"module"`
	// IDs id 列表
	IDs []uint `json:"ids" validate:"required,gt=0"`
}

func (opt *GetNameByModuleIDsRequest) Validity(resp *GetNameByModuleIDsResponse) {
	if len(opt.Request.ModuleIDList) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_module_id_list_empty"
		return
	}
}

type GetNameByModuleIDsResponse struct {
	apiobj.BaseResponse
	Response GetNameByModuleIDsEmbedResponse
}

type GetNameByModuleIDsEmbedResponse struct {
	// NameList 名称列表
	NameList []GetNameByModuleIDsNameListItem `json:"name_list"`
}

type GetNameByModuleIDsNameListItem struct {
	// ID 主键 id
	ID uint `json:"id"`
	// Module 模块，forest：知识库，database：数据库，table：表
	Module foresttype.ForestModule
	// Name 名称
	Name string `json:"name"`
}

type RenameForestRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
}

const MaxForestName = 50

func (opt *RenameForestRequest) Validity(resp *apiobj.BaseResponse) {
	if opt.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_id_empty"
		return
	}
	if len(opt.Request.Name) <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_empty"
	}
	if len(opt.Request.Name) > MaxForestName {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_name_too_long"
		return
	}
}
