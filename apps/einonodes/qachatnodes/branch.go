package qachatnodes

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/ygpkg/yg-go/logs"
)

type branch struct {
}

func newBranch() *branch {
	return &branch{}
}

func (b *branch) CheckStatementRes(ctx context.Context, input nodebase.RecordList) (next string, err error) {
	next = NodeExcelTransferSQLToAnswerExecutor
	_ = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		sqlResValue := input.Get(RecordKeyQueryStatementRes).Value
		var sqlResList []map[string]any
		if err := json.Unmarshal([]byte(sqlResValue), &sqlResList); err != nil {
			logs.ErrorContextf(state.Ctx, "[CheckStatementRes] unmarshal sql res error: %v", err)
			return err
		}
		if len(sqlResList) == 0 {
			next = NodeNoDataReporter
			return nil
		}
		return nil
	})
	return next, nil
}
