package svcmembership

import (
	"testing"
	"time"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNotifyExpiringQuotas(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	_ = testutils.NewCtx(testutils.WithUin(384))

	res, err := NotifyExpiringQuotas()
	assert.Nil(t, err)
	t.Log(res)
	time.Sleep(time.Second)
}
