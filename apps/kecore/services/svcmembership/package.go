package svcmembership

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/runtime"
)

func ListPackage(ctx *gin.Context, req *dtomembership.ListPackageRequest) (res *dtomembership.ListPackageResponse, err error) {
	packageEntityList, err := forest.NewKePackageDao().GetListByCond(ctx, &forest.KePackageCond{})
	if err != nil {
		return nil, err
	}

	// 统计已购买的配额记录
	now := time.Now()
	packageLevelCountMap, err := forest.NewKeCompanyQuotaDao().GetGroupCountByPackageLevel(ctx, &forest.KeCompanyQuotaCond{
		CompanyID:     runtime.CompanyID(ctx),
		SourceType:    foresttype.CompanyQuotaSourceTypeOrder,
		ExpireAtStart: &now,
	})
	if err != nil {
		return nil, err
	}
	list := make([]dtomembership.ListPackageItem, 0, len(packageEntityList))
	for _, v := range packageEntityList {
		item := dtomembership.ListPackageItem{
			PackageID:       v.ID,
			Name:            v.Name,
			Description:     v.Description,
			PackageLevel:    v.Level,
			Price:           v.Price,
			SalePrice:       v.SalePrice,
			AgentQuota:      v.AgentQuota,
			QaQuota:         v.QaQuota,
			DiskQuota:       v.DiskQuota,
			EmployeeQuota:   v.EmployeeQuota,
			ArticleQuota:    v.ArticleQuota,
			Edition:         v.Edition,
			AdditionalNotes: v.Extra.AdditionalNotes,
			IsPurchased:     false,
		}
		if len(item.AdditionalNotes) == 0 {
			item.AdditionalNotes = make([]string, 0)
		}
		if v.Edition != foresttype.PackageEditionFreeTrail {
			item.IsPurchased = packageLevelCountMap[item.PackageLevel] > 0
		}
		list = append(list, item)
	}

	// 如果所有套餐的 IsPurchased 都为 false，则将免费版的 IsPurchased 设置为 true
	allNotPurchased := true
	for i := range list {
		if list[i].IsPurchased {
			allNotPurchased = false
			break
		}
	}
	if allNotPurchased {
		for i := range list {
			if list[i].Edition == foresttype.PackageEditionFreeTrail {
				list[i].IsPurchased = true
			}
		}
	}

	// 获取当前用户已购买的套餐ID集合
	res = &dtomembership.ListPackageResponse{}
	res.Response.List = list
	return res, nil
}
