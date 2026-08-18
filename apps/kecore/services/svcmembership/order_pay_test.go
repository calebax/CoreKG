package svcmembership

import (
	"testing"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/apps/kecore/services/membership"
	"github.com/insmtx/corekg/apps/kesale"
	"github.com/insmtx/corekg/apps/kesale/callbacks"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestCreateOrder(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)

	kesale.Init(dbutil.Knownow(), global.SettingGroupCore, global.SettingKeySalePay)
	callbacks.AppendGlobalHandlers(membership.NewPaymentHandler())

	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	res, err := CreateOrder(ctx, &dtomembership.CreateOrderRequest{
		Request: dtomembership.CreateOrderEmbedRequest{
			PackageID: 2,
		},
	})
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}
