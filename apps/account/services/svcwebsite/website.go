package svcwebsite

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtowebsite"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

func UpdateWebsiteInfo(ctx *gin.Context, req *dtowebsite.UpdateWebsiteInfoRequest) (res *dtowebsite.UpdateWebsiteInfoResponse, err error) {
	res = &dtowebsite.UpdateWebsiteInfoResponse{}
	var webInfo dtowebsite.WebSiteInfo
	if err := settings.GetYaml("core", "website-info", &webInfo); err != nil {
		logs.ErrorContextf(ctx, "Failed to get website info: %v", err)
		return res, err
	}
	webInfo.WebsiteLogo = req.Request.WebSiteInfo.WebsiteLogo
	webInfo.WebsiteName = req.Request.WebSiteInfo.WebsiteName
	if err := settings.SetYaml("core", "website-info", webInfo); err != nil {
		logs.ErrorContextf(ctx, "Failed to update website info: %v", err)
		return res, err
	}
	return res, nil
}
