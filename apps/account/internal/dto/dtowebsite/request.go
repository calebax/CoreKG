package dtowebsite

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type UpdateWebsiteInfoRequest struct {
	apiobj.BaseRequest
	Request UpdateWebsiteInfoEmbedRequest
}

type UpdateWebsiteInfoEmbedRequest struct {
	WebSiteInfo WebSiteInfo `json:"website_info"`
}

func (opt *UpdateWebsiteInfoRequest) Validity(resp *UpdateWebsiteInfoResponse) {
	if opt.Request.WebSiteInfo.WebsiteLogo == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_website_logo_empty"
		return
	}
	if opt.Request.WebSiteInfo.WebsiteName == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_website_name_empty"
		return
	}
}

type WebSiteInfo struct {
	WebsiteLogo string `json:"website_logo"`
	WebsiteName string `json:"website_name"`
}
