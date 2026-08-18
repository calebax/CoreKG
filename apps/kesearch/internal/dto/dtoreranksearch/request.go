package dtoreranksearch

import (
	"github.com/insmtx/corekg/apps/kesearch/models/reranksearch"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type RerankSearchChunkRequest struct {
	apiobj.BaseRequest
	Request RerankSearchChunkEmbedRequest
}

type RerankSearchChunkEmbedRequest struct {
	Question  string                     `json:"question"`
	ForestIDs []uint                     `json:"forest_ids"`
	FileIDs   []uint                     `json:"file_ids"`
	Config    *reranksearch.SearchConfig `json:"config"`
}

func (opt *RerankSearchChunkRequest) Validity(resp *RerankSearchChunkResponse) {
	if opt.Request.Question == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	if len(opt.Request.ForestIDs) == 0 && len(opt.Request.FileIDs) == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kechat_invalid_params" // 参数错误
		return
	}
	if opt.Request.Config != nil {
		if err := opt.Request.Config.Validity(); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "kesearch_reranksearch_config_error" // 查询条件中的字段不存在
			resp.MessageData = map[string]interface{}{
				"error": err.Error(),
			}
			return
		}
	}
}
