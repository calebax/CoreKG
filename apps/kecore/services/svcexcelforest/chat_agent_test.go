package svcexcelforest

import (
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestRequestGetHeaderNumbers(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	data := []string{
		"重点产业\t重点产业\t重点产业\t重点产业",
		"指标\t2024年（亿元）\t同比±%\t占GDP比重±%",
		"*数字经济核心产业增加值\t6305\t7.1\t28.8",
		"*文化产业增加值\t3448\t6.5\t15.8",
	}
	res, err := RequestGetHeaderNumbers(ctx, data[:2])
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}
