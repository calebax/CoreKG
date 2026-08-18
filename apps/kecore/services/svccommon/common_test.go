package svccommon

import (
	"testing"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtocommon"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestGetCommonInfo(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(19))

	res, err := GetCommonInfo(ctx, &dtocommon.GetCommonInfoRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	t.Log("res:", logs.JSON(res))
}
