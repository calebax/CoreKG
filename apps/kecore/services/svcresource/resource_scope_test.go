package svcresource

import (
	"testing"

	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoresource"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestSetResourceScope(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	res, err := SetResourceScope(ctx, &dtoresource.SetResourceScopeRequest{
		Request: dtoresource.SetResourceScopeEmbedRequest{
			ResourceType:   foresttype.CozeResourceTypePlugin,
			ResourceID:     1,
			ViewScopeType:  foresttype.ScopeTypeUser,
			ViewScopeIDs:   []uint{384, 385},
			ManageScopeIDs: []uint{384},
		},
	})
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}

func TestGetResourceScope(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	res, err := GetResourceScope(ctx, &dtoresource.GetResourceScopeRequest{
		Request: dtoresource.GetResourceScopeEmbedRequest{
			ResourceType: foresttype.CozeResourceTypePlugin,
			ResourceIDs:  []uint{1},
		},
	})
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}
