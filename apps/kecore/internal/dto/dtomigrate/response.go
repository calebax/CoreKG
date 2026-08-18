package dtomigrate

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type MigrateInterfaceResponse struct {
	apiobj.BaseResponse
	Response MigrateInterfaceEmbedResponse
}

type MigrateInterfaceEmbedResponse struct {
}
