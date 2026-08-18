package corearticle

import (
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/apis/constants"
)

func TestAIWriteExecutor_Execute(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()

	ctx := testutils.NewCtx(testutils.WithUin(384))
	ctx.Set(constants.CtxKeyLang, "")

	executor := NewAIWriteExecutor(AIWriteExecutorParams{
		Ctx:         ctx,
		GinCtx:      nil,
		Cmd:         foresttype.CmdExpansion,
		Content:     "中集来福士",
		ForestIDs:   "[1051]",
		RequestID:   "test-request-id",
		CompanyID:   1,
		Uin:         384,
		ChatModelID: 0,
	})

	var output strings.Builder
	err := executor.Execute(&output)

	assert.Nil(t, err)
	t.Logf("output: %s", output.String())
}
