package dtograph

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateForestGraphRequest struct {
	apiobj.BaseRequest
	Request CreateForestGraphEmbedRequest
}

type CreateForestGraphEmbedRequest struct {
	ForestID  uint   `json:"forest_id"` // 关联的知识森林 ID
	AvatarUrl string `json:"avatar_url"`
}

func (opt *CreateForestGraphRequest) Validity(resp *CreateForestGraphResponse) {
	if opt.Request.ForestID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "kecore_forest_id_empty"
	}
}
