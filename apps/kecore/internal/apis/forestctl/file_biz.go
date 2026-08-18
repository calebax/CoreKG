package forestctl

import (
	"path/filepath"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtask"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type PreUploadFileRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID    uint                 `json:"forest_id"`
		ParentID    uint                 `json:"parent_id"`
		FileName    string               `json:"file_name"`
		FileSize    int64                `json:"file_size"`
		SplitConfig *ragtask.SplitConfig `json:"split_config"`
	}
}

func (opt *PreUploadFileRequest) Validity(resp *PreUploadFileResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if opt.Request.FileName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_enter_file_name" // 请输入文件名
		return
	}
	if opt.Request.FileSize == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_size_empty" // 文件大小为空
		return
	}
	if filepath.Ext(opt.Request.FileName) == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_type_empty" // 文件类型为空
		return
	}

	if opt.Request.SplitConfig == nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_config_empty" // 分片配置为空
		return
	}
	if opt.Request.SplitConfig.SplitMode == ragtask.SplitAuto {
		return
	}
	if opt.Request.SplitConfig.SplitMode != ragtask.SplitAuto &&
		opt.Request.SplitConfig.SplitMode != ragtask.SplitRule {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_mode_invalid" // 分片模式错误
		return
	}
	if opt.Request.SplitConfig.ChunkSize > 1024 || opt.Request.SplitConfig.ChunkSize < 256 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_chunk_size_invalid" // 分片大小为空
		return
	}
	if len(opt.Request.SplitConfig.SplitMark) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_mark_empty" // 分片标识符为空
		return
	}
	if opt.Request.SplitConfig.SplitOverlap < 0 || opt.Request.SplitConfig.SplitOverlap > 1 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_overlap_invalid" // 请输入正确的分片重叠度
		return
	}

}

type PreUploadFileResponse struct {
	apiobj.BaseResponse
	Response struct {
		UploadURL string `json:"upload_url"`
		FileID    uint   `json:"file_id"`
	}
}

type UploadFileCallBackRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID uint                  `json:"forest_id"`
		FileID   uint                  `json:"file_id"`
		Status   foresttype.FileStatus `json:"status"`
	}
}

func (opt *UploadFileCallBackRequest) Validity(resp *UploadFileCallBackResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if opt.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_id_empty" // 文件id为空
		return
	}
	if opt.Request.Status == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_file_status_unknown" // 未知的文件状态
		return
	}
}

type UploadFileCallBackResponse struct {
	apiobj.BaseResponse
	Response struct {
		UploadURL string `json:"upload_url"`
		FileID    uint   `json:"file_id"`
	}
}

type ListParseHistoryRequest apiobj.QueryRequest

type ListParseHistoryResponse struct {
	apiobj.BaseResponse
	Response forest.ListParseHistoryResponse
}

func (req *ListParseHistoryRequest) Validate(resp *ListParseHistoryResponse) {
	for _, v := range req.Request.Filters {
		switch v.Field {
		case "recent_days":
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

type RetryParseRequest apiobj.DetailIdRequest

func (r *RetryParseRequest) Validate(resp *apiobj.BaseResponse) {
	if r.Request.ID <= 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_invalid_file_id" // 非法文件id
	}
}

type ResplitChunkRequest struct {
	apiobj.BaseRequest
	Request struct {
		ForestID    uint                 `json:"forest_id"`
		FileID      uint                 `json:"file_id"`
		SplitConfig *ragtask.SplitConfig `json:"split_config"`
	}
}

func (req *ResplitChunkRequest) Validate(resp *ResplitChunkResponse) {
	if req.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_forest" // 请选择知识森林
		return
	}
	if req.Request.FileID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_select_file" // 请选择文件
		return
	}
	if req.Request.SplitConfig == nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_config_empty" // 分片配置为空
		return
	}
	if req.Request.SplitConfig.SplitMode == ragtask.SplitAuto {
		return
	}
	if req.Request.SplitConfig.SplitMode != ragtask.SplitAuto &&
		req.Request.SplitConfig.SplitMode != ragtask.SplitRule {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_mode_invalid" // 分片模式错误
		return
	}
	if req.Request.SplitConfig.ChunkSize > 1024 || req.Request.SplitConfig.ChunkSize < 256 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_chunk_size_invalid" // 分片大小为空
		return
	}
	if len(req.Request.SplitConfig.SplitMark) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_mark_empty" // 分片标识符为空
		return
	}
	if req.Request.SplitConfig.SplitOverlap < 0 || req.Request.SplitConfig.SplitOverlap > 1 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_split_overlap_invalid" // 请输入正确的分片重叠度
		return
	}
}

type ResplitChunkResponse struct {
	apiobj.BaseResponse
}
