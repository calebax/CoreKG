package article

import (
	"context"
	"testing"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func TestMigrateGraphAndArticleQuota(t *testing.T) {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}

	var cmps []*accounttype.Company
	if err := dbutil.Account().Table(accounttype.TableNameCompany).
		Where("deleted_at IS NULL").
		Find(&cmps).
		Error; err != nil {
		panic(err)
	}
	if err := dbutil.Account().Transaction(func(tx *gorm.DB) error {
		for _, v := range cmps {
			if v.Quota == nil || v.Quota.QAQuota == 0 {
				v.Quota = &accounttype.ResourceQuota{DiskQuota: 10 * forest.GB, QAQuota: 1000, AgentQuota: 5, EmployeeQuota: 5, GraphQuota: 5, ArticleQuota: 5}
			} else {
				v.Quota.ArticleQuota = 5
				v.Quota.GraphQuota = 5
			}
			logs.DebugContextf(context.Background(), "company: %+v", v)
			if err := dbutil.Account().Save(v).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
