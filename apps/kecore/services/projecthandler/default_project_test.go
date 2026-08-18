package projecthandler

import (
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestCreateDefaultProject(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	res, err := CreateDefaultProject(ctx, 2, 384)
	assert.Nil(t, err)
	t.Logf("res: %s", logs.JSON(res))
}
