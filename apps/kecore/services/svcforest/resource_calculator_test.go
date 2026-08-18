package svcforest

import (
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestResourceMetrics(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	inst := NewResourceCalculator(ResourceCalcTypeQAPair)
	res, err := inst.Metrics(ctx, 884)
	assert.Nil(t, err)
	t.Logf("res: %s", logs.JSON(res))
}
