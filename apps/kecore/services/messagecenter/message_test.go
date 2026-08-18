package messagecenter

import (
	"testing"
	"time"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestSendMessage(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	req := &SendMessageReq{
		UserID:       1,
		Uin:          384,
		TemplateName: "announcement_new_feature",
		SourceType:   "announcement",
		SourceID:     1,
		MessageParams: map[string]string{
			"feature_name":    "公告 1",
			"announcement_id": "1",
		},
	}
	res, err := NewMessage().SendMessage(ctx, req)
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
}

func TestBatchSendMessage(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	reqItem := &SendMessageReq{
		UserID:       1,
		Uin:          384,
		TemplateName: "announcement_new_feature",
		SourceType:   "announcement",
		SourceID:     1,
		MessageParams: map[string]string{
			"feature_name":    "公告 2",
			"announcement_id": "2",
		},
	}
	req := []*SendMessageReq{reqItem}
	res, err := NewMessage().BatchSendMessage(ctx, req)
	assert.Nil(t, err)
	t.Log(logs.JSON(res))
	time.Sleep(time.Second)
}

func TestMarkAsRead(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))

	err := NewMessage().MarkAsRead(ctx, 1)
	assert.Nil(t, err)
}
