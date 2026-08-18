package dtoperm

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type SetResourcePermResponse struct {
	apiobj.BaseResponse
	Response SetResourcePermEmbedResponse `json:"response"`
}

type SetResourcePermEmbedResponse struct {
}

type GetResourcePermResponse struct {
	apiobj.BaseResponse
	Response GetResourcePermEmbedResponse `json:"response"`
}

type AccessResult struct {
	ManagerList []uint
	ViewerList  []uint
	BanList     []uint
	ScopeType   foresttype.PublicScope
}
type GetResourcePermEmbedResponse struct {
	AccessResult AccessResult `json:"access_result"`
}
