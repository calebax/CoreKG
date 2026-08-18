package apikey

import (
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type QueryListAPIKeyResponse struct {
	apiobj.QueryResponse
	Data []*accounttype.APIKey
}
