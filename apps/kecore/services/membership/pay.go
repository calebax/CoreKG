/*
 * @Author: morehao morehao@qq.com
 * @Date: 2025-11-28 20:34:00
 * @LastEditors: morehao morehao@qq.com
 * @LastEditTime: 2025-11-28 21:10:13
 * @FilePath: /roc/apps/kecore/services/membership/pay.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package membership

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesale/callbacks"
	"github.com/insmtx/corekg/apps/kesale/models"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// PaymentHandler 支付完成处理器
type PaymentHandler struct {
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

// BusinessType 返回业务类型
func (h *PaymentHandler) BusinessType() models.BusinessType {
	return models.BusinessTypeSubscription
}

// OnPaymentCompleted 支付完成回调处理
func (h *PaymentHandler) OnPaymentCompleted(ctx context.Context, payInfo *callbacks.PayInfo) context.Context {

	ctx = logs.WithContextFields(ctx, "orderSN", payInfo.OrderSN)

	logs.InfoContextf(ctx, "[PaymentHandler.OnPaymentCompleted] start callback, payInfo: %s", logs.JSON(payInfo))

	now := time.Now()
	existingQuotaEntityList, err := forest.NewKeCompanyQuotaDao().GetListByCond(ctx, &forest.KeCompanyQuotaCond{
		CompanyID:       payInfo.CompanyID,
		SourceType:      foresttype.CompanyQuotaSourceTypeOrder,
		ExpireAtStart:   &now,
		MinPackageLevel: foresttype.PackageLevel1,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[PaymentHandler.OnPaymentCompleted] get existing quota fail, companyID: %d, err: %v", payInfo.CompanyID, err)
		return nil
	}
	packageLevelMap := make(map[foresttype.PackageLevel]foresttype.KeCompanyQuota)
	for _, v := range existingQuotaEntityList {
		packageLevelMap[v.PackageLevel] = v
	}

	packageQuotaList, err := h.calcQuota(ctx, payInfo.Product)
	if err != nil {
		logs.ErrorContextf(ctx, "[PaymentHandler.OnPaymentCompleted] calc quota fail, companyID: %d, err: %v", payInfo.CompanyID, err)
		return nil
	}
	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, v := range packageQuotaList {
			existingQuotaEntity, ok := packageLevelMap[v.PackageLevel]
			if ok {
				expireAt := existingQuotaEntity.ExpireAt.Add(time.Hour * 24 * time.Duration(v.Days))
				updateMap := map[string]any{
					"expire_at": expireAt,
				}
				if err := forest.NewKeCompanyQuotaDao().WithTx(tx).UpdateMap(ctx, existingQuotaEntity.ID, updateMap); err != nil {
					return err
				}
			} else {
				// 当天的第一秒（00:00:00）
				effectiveAt := time.Now()
				// 到期日期的最后一秒（23:59:59.999）
				expireAt := effectiveAt.Add(time.Hour * 24 * time.Duration(v.Days))
				entity := &foresttype.KeCompanyQuota{
					CompanyID:     payInfo.CompanyID,
					SourceType:    foresttype.CompanyQuotaSourceTypeOrder,
					PackageLevel:  v.PackageLevel,
					OperatorID:    payInfo.UinID,
					AgentQuota:    v.AgentQuota,
					QaQuota:       v.QaQuota,
					DiskQuota:     v.DiskQuota,
					EmployeeQuota: v.EmployeeQuota,
					ArticleQuota:  v.ArticleQuota,
					EffectiveAt:   &effectiveAt,
					ExpireAt:      &expireAt,
				}
				if err := forest.NewKeCompanyQuotaDao().WithTx(tx).Insert(ctx, entity); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if txErr != nil {
		logs.ErrorContextf(ctx, "[PaymentHandler.OnPaymentCompleted] transaction fail, companyID: %d, err: %v", payInfo.CompanyID, txErr)
		return nil
	}

	return nil
}

func (h *PaymentHandler) calcQuota(ctx context.Context, products []*models.OrderProduct) ([]PackageQuota, error) {
	productCountMap := make(map[uint]int64)
	var packageIDs []uint
	for _, product := range products {
		productCountMap[product.ProductID] += int64(product.Quantity)
		packageIDs = append(packageIDs, product.ProductID)
	}
	packageEntityList, err := forest.NewKePackageDao().GetListByCond(ctx, &forest.KePackageCond{
		IDs: packageIDs,
	})
	if err != nil {
		return nil, err
	}
	packageEntityMap := packageEntityList.ToMap()
	var packageQuotaList []PackageQuota
	for packageID, count := range productCountMap {
		packageEntity, ok := packageEntityMap[packageID]
		if !ok {
			continue
		}
		renewalDays := foresttype.PackageLevelDaysMap[packageEntity.Level] * count
		item := PackageQuota{
			PackageID:     packageID,
			PackageLevel:  packageEntity.Level,
			Quantity:      count,
			Days:          renewalDays,
			AgentQuota:    packageEntity.AgentQuota,
			QaQuota:       packageEntity.QaQuota,
			DiskQuota:     packageEntity.DiskQuota,
			EmployeeQuota: packageEntity.EmployeeQuota,
			ArticleQuota:  packageEntity.ArticleQuota,
		}
		packageQuotaList = append(packageQuotaList, item)
	}
	return packageQuotaList, nil
}
