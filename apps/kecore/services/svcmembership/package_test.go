package svcmembership

import (
	"testing"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtomembership"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestListPackage(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)

	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	res, err := ListPackage(ctx, &dtomembership.ListPackageRequest{
		Request: dtomembership.ListPackageEmbedRequest{},
	})
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}
