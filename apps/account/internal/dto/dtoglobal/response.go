package dtoglobal

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type GetGlobalInfoResponse struct {
	apiobj.BaseResponse
	Response GetGlobalInfoEmbedResponse
}

type GetGlobalInfoEmbedResponse struct {
	WebsiteInfo WebsiteInfo `json:"website_info"`
}

type WebsiteInfo struct {
	WebsiteLogo string `json:"website_logo"`
	WebsiteName string `json:"website_name"`
}
