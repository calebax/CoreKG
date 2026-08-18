package membership

import (
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestQuotaCheck(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(305))

	valid, balance, err := NewQuotaManager().Check(ctx, &QuotaCheckReq{
		CompanyID:    305,
		ResourceType: QuotaResourceTypeDisk,
	})
	assert.Nil(t, err)
	t.Log("valid:", valid)
	t.Log("balance:", balance)
}

func TestQuotaCheckArticle(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(305))

	valid, balance, err := NewQuotaManager().Check(ctx, &QuotaCheckReq{
		CompanyID:    305,
		ResourceType: QuotaResourceTypeArticle,
	})
	assert.Nil(t, err)
	t.Log("valid:", valid)
	t.Log("balance:", balance)
}

func TestQuotaQuery(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(19))

	res, err := NewQuotaManager().Query(ctx, &QuotaQueryReq{
		CompanyID: 2,
	})
	assert.Nil(t, err)
	t.Log("res:", logs.JSON(res))
}
