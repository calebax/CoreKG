package svcdashboard

import (
	"testing"
	"time"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func Test_countMultimodalUploadFile(t *testing.T) {
	testutils.Initialize(testutils.AppNameAdmin)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	now := time.Now()
	beginAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endAt := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	req := &StatQuery{
		BeginAt: beginAt,
		EndAt:   endAt,
	}
	res, err := countMultimodalUploadFile(ctx, req)
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}
