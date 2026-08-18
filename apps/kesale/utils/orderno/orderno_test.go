package orderno

import (
	"log"
	"testing"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestGenerateUniqueness(t *testing.T) {
	m := map[string]struct{}{}
	for i := 0; i < 99000; i++ {
		id := Generate("", i)
		if _, ok := m[id]; ok {
			t.Fatalf("duplicate id: %s", id)
		}
		m[id] = struct{}{}
	}
}

func TestGeneratePrefix(t *testing.T) {
	id := Generate("", 1)
	log.Printf("generated id: %s", id)
	if len(id) != 2+17 {
		t.Errorf("expected length %d, got %d", 2+17, len(id))
	}
}

func TestGeneratePaymentTradeNoUniqueness(t *testing.T) {
	m := map[string]struct{}{}
	for i := 0; i < 50000; i++ {
		tradeNo := GeneratePaymentTradeNo()
		if _, ok := m[tradeNo]; ok {
			t.Fatalf("duplicate payment trade no: %s", tradeNo)
		}
		m[tradeNo] = struct{}{}
	}
}

func TestGeneratePaymentTradeNoFormat(t *testing.T) {
	tradeNo := GeneratePaymentTradeNo()
	log.Printf("generated payment trade no: %s", tradeNo)

	// 验证长度为20位
	if len(tradeNo) != 20 {
		t.Errorf("expected length 20, got %d", len(tradeNo))
	}

	// 验证所有字符都是数字
	for _, c := range tradeNo {
		if c < '0' || c > '9' {
			t.Errorf("expected all digits, got non-digit character: %c", c)
		}
	}
}

func TestMigrateCompanyQuota(t *testing.T) {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}

	var (
		cmps     []*accounttype.Company
		cmpQuota []*foresttype.KeCompanyQuota
	)
	if err := dbutil.Account().
		Table(accounttype.TableNameCompany).
		Find(&cmps).
		Error; err != nil {
		panic(err)
	}

	for _, v := range cmps {
		now := time.Now()
		time2099 := time.Date(2099, 1, 1, 0, 0, 0, 0, now.Location())
		cmpQuota = append(cmpQuota, &foresttype.KeCompanyQuota{
			CompanyID:     v.ID,
			SourceType:    foresttype.CompanyQuotaSourceTypeManual,
			PackageLevel:  1,
			OperatorID:    0,
			AgentQuota:    int64(v.Quota.AgentQuota),
			QaQuota:       int64(v.Quota.QAQuota),
			DiskQuota:     v.Quota.DiskQuota,
			EmployeeQuota: int64(v.Quota.EmployeeQuota),
			// 2099
			EffectiveAt: &now,
			ExpireAt:    &time2099,
		})
	}
	if err := dbutil.Knownow().CreateInBatches(&cmpQuota, 50).Error; err != nil {
		t.Fatal(err)
	}

}
