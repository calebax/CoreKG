package qapair

import (
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

func TestListQAPair(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()
	ctx := testutils.NewCtx(testutils.WithUin(384))
	req := ListQAPairRequest{
		Request: struct {
			apiobj.PageQuery
			ForestID uint `json:"forest_id"`
		}{
			ForestID: 884,
			PageQuery: apiobj.PageQuery{
				Filters: []apiobj.Filter{
					{Field: "qa_question", Value: []string{"篮球"}},
				},
				Limit:   10,
				OrderBy: []string{"updated_at desc"},
			},
		},
	}
	resp := &ListQAPairResponse{}
	ListQAPair(ctx, &req, resp)
	t.Log(resp)
}
