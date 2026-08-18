package svcmembership

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/account"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/services/messagecenter"
	"github.com/insmtx/corekg/pkgs/utils/ginctx"
	"github.com/ygpkg/yg-go/logs"
)

func NotifyExpiringQuotas() (string, error) {
	ctx := ginctx.InitJobCtx("NotifyExpiringQuotas", "")
	logs.InfoContext(ctx, "[NotifyExpiringQuotas] start notify expiring quotas")
	if err := notifyExpiringQuotas(ctx); err != nil {
		logs.ErrorContextf(ctx, "[NotifyExpiringQuotas] notify expiring quotas fail, err: %v", err)
		return "", err
	}
	return "", nil
}

// notifyExpiringQuotas 通知即将过期的配额
func notifyExpiringQuotas(ctx context.Context) error {

	// 获取7天内即将过期的配额
	expireAtStart := time.Now()
	sevenDaysLater := expireAtStart.Add(time.Hour * 24 * 7)

	quotaList, err := forest.NewKeCompanyQuotaDao().GetListByCond(ctx, &forest.KeCompanyQuotaCond{
		ExpireAtStart: &expireAtStart,
		ExpireAtEnd:   &sevenDaysLater,
	})
	if err != nil {
		return fmt.Errorf("get expiring quotas fail, err: %v", err)
	}

	if len(quotaList) == 0 {
		logs.InfoContextf(ctx, "no expiring quotas found in next 7 days")
		return nil
	}

	// 获取消息服务实例
	msgService := messagecenter.NewMessage()

	var companyIDs []uint
	companyExpireDateMap := make(map[uint]string)
	companySourceIDMap := make(map[uint]uint)

	// 遍历每个即将过期的配额，发送消息
	for _, quota := range quotaList {
		if quota.ExpireAt == nil {
			continue
		}
		companyIDs = append(companyIDs, quota.CompanyID)
		companyExpireDateMap[quota.CompanyID] = quota.ExpireAt.Format("2006年01月02日")
		companySourceIDMap[quota.CompanyID] = quota.ID
	}
	// 获取 uin 列表
	uinEntityList, err := account.NewUserIdentificationDao().GetListByCond(ctx, &account.UserIdentificationCond{
		SubjectType: accounttype.SubjectTypeCompany,
		SubjectIDs:  companyIDs,
	})

	if err != nil {
		logs.ErrorContextf(ctx, "[NotifyExpiringQuotas] get uin list fail, company_ids: %s, err: %v", logs.JSON(companyIDs), err)
		return err
	}
	sendMessageItem := make([]*messagecenter.SendMessageReq, 0, len(uinEntityList))
	for _, v := range uinEntityList {
		sendMessageItem = append(sendMessageItem, &messagecenter.SendMessageReq{
			TemplateName: foresttype.MessageTemplateNameOrderAboutToExpire,
			CompanyID:    v.SubjectID,
			UserID:       v.UserID,
			Uin:          v.ID,
			SourceType:   foresttype.MessageSourceTypeCompany,
			SourceID:     companySourceIDMap[v.SubjectID],
			MessageParams: map[string]string{
				"expire_date": companyExpireDateMap[v.SubjectID],
			},
		})
	}
	_, err = msgService.BatchSendMessage(ctx, sendMessageItem)
	if err != nil {
		logs.ErrorContextf(ctx, "[NotifyExpiringQuotas] batch send expiring quota message fail, err: %v", err)
		return err
	}

	return nil
}
