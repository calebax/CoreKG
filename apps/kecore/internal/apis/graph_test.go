package apis

import (
	"testing"

	"github.com/insmtx/corekg/apps/kecore/internal/apis/graphctl"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtograph"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCopyPreviousVersionManualNodes(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()

	var (
		forestID uint = 1032
		graphID  uint = 333
	)

	ctx := testutils.NewCtx(testutils.WithUin(19))
	err := nebulagraph.InitNebulaConf(ctx)
	assert.Nil(t, err)

	createRes := &dtograph.CreateForestGraphResponse{}
	CreateForestGraph(ctx, &dtograph.CreateForestGraphRequest{
		Request: dtograph.CreateForestGraphEmbedRequest{
			ForestID: forestID,
		},
	}, createRes)
	t.Logf("%+v", createRes)
	if createRes.Code != 0 {
		t.Fatalf("%+v", createRes)
	}

	//===============================================
	parseRes := &graphctl.ParseGraphResponse{}
	graphctl.ParseGraph(ctx, &graphctl.ParseGraphRequest{
		Request: struct {
			GraphID uint `json:"graph_id"`
		}{GraphID: graphID},
	}, parseRes)

	t.Logf("%+v", parseRes)
	if parseRes.Code != 0 {
		t.Fatalf("%+v", parseRes)
	}
}
