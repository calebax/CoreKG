package svcexcelforest

import (
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzeXlsx(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	req := AnalyzeXlsxReq{
		ForestFileID: 6194,
	}
	err := AnalyzeXlsx(ctx, &req)
	assert.Nil(t, err)
}
