package svccommon

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtocommon"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/services/membership"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func GetCommonInfo(ctx *gin.Context, req *dtocommon.GetCommonInfoRequest) (res *dtocommon.GetCommonInfoResponse, err error) {

	res = &dtocommon.GetCommonInfoResponse{}

	companyID := runtime.CompanyID(ctx)

	// 公司配额
	companyQuota, err := getCompanyQuota(ctx, companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "[svccommon.GetCommonInfo] getCompanyQuota failed, companyID: %d, err: %v", companyID, err)
		return res, nil
	}

	res = &dtocommon.GetCommonInfoResponse{
		Response: dtocommon.GetCommonInfoEmbedResponse{
			CompanyQuota: *companyQuota,
		},
	}
	return res, nil
}

func getCompanyQuota(ctx *gin.Context, companyID uint) (*dtocommon.CompanyQuota, error) {
	quotaRes, err := membership.NewQuotaManager().Query(ctx, &membership.QuotaQueryReq{
		CompanyID: companyID,
	})
	if err != nil {
		return nil, err
	}
	// 统计已购买的配额记录
	now := time.Now()
	purchasedCount, err := forest.NewKeCompanyQuotaDao().CountByCond(ctx, &forest.KeCompanyQuotaCond{
		CompanyID:     companyID,
		SourceType:    foresttype.CompanyQuotaSourceTypeOrder,
		ExpireAtStart: &now,
	})
	if err != nil {
		return nil, err
	}

	ui := runtime.Uin(ctx)
	usr, err := user.GetUserByUin(ctx, ui)
	if err != nil {
		return nil, fmt.Errorf("getCompanyQuota: get user by id failed, userID: %d, err: %v", ui, err)
	}

	companyQuota := &dtocommon.CompanyQuota{
		IsPurchased:       purchasedCount > 0,
		AgentQuota:        quotaRes.AgentQuota,
		AgentQuotaUsed:    quotaRes.AgentQuotaUsed,
		QaQuota:           quotaRes.QaQuota,
		QaQuotaUsed:       quotaRes.QaQuotaUsed,
		DiskQuota:         quotaRes.DiskQuota,
		DiskQuotaUsed:     quotaRes.DiskQuotaUsed,
		EmployeeQuota:     quotaRes.EmployeeQuota,
		EmployeeQuotaUsed: quotaRes.EmployeeQuotaUsed,
		ArticleQuota:      quotaRes.ArticleQuota,
		ArticleQuotaUsed:  quotaRes.ArticleQuotaUsed,
		CompanyQuota:      usr.CompanyQuota,
	}
	return companyQuota, nil
}
