package dtomigrate

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type MigrateInterfaceRequest struct {
	apiobj.BaseRequest
	Request MigrateInterfaceEmbedRequest
}



type MigrateInterfaceEmbedRequest struct {
	// 业务类型
	BusinessType string `json:"business_type"`
}

func (opt *MigrateInterfaceRequest) Validity(resp *MigrateInterfaceResponse) {
}
