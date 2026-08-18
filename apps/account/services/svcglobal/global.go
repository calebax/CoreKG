package svcglobal

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtoglobal"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

func GetGlobalInfo(ctx *gin.Context, req *dtoglobal.GetGlobalInfoRequest) (res *dtoglobal.GetGlobalInfoResponse, err error) {
	res = &dtoglobal.GetGlobalInfoResponse{}
	var websiteInfo dtoglobal.WebsiteInfo

	if err := settings.GetYaml("core", "website-info", &websiteInfo); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logs.ErrorContextf(ctx, "Failed to get website info: %v", err)
			return res, err
		}
		logs.WarnContextf(ctx, "Failed to get website info: %v", err)
	}
	res.Response.WebsiteInfo = websiteInfo

	return res, nil
}
