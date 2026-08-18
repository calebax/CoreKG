package einodemo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/stretchr/testify/assert"
)

func TestSingleGraph(t *testing.T) {
	ctx := context.Background()
	r, err := SingleGraph(ctx)
	assert.NoError(t, err)
	out, err := r.Invoke(ctx, nodebase.RecordList{
		&nodebase.Record{
			Key:   "question",
			Value: "你好",
		},
	})
	assert.NoError(t, err)
	jsonStr, err := json.Marshal(out)
	assert.NoError(t, err)
	t.Log(string(jsonStr))
}
