package qachatnodes

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/insmtx/corekg/apps/einonodes/nodebase"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

type reporter struct {
	*baseHandler
}

func newReporter() *reporter {
	return &reporter{
		baseHandler: &baseHandler{},
	}
}

func (r *reporter) NoDataReportNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState[*State](ctx, func(_ context.Context, state *State) error {
		noDataCaseName := input.Get(RecordKeyNoDataCaseName).Value
		text, ok := NoDataCaseNameTextMap[noDataCaseName]
		if !ok {
			logs.ErrorContextf(ctx, "[NoDataReportNode] no data case name not found: %s", noDataCaseName)
			text = i18n.T(runtime.GetLanguage(state.Ctx), "kechat_no_data")
		}
		llmchat.WriteContent(state.Ctx, state.QuestionEntity.Source.ReqID, i18n.T(runtime.GetLanguage(state.Ctx), text))
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[NoDataReportNode] ProcessState error: %v", err)
		return nil, err
	}
	return output, nil
}

func (r *reporter) EChartReportNode(ctx context.Context, input nodebase.RecordList, opts ...any) (output nodebase.RecordList, err error) {
	err = compose.ProcessState(ctx, func(_ context.Context, state *State) error {
		answer := state.QuestionEntity.Source.Answer

		for _, record := range input {
			if _, ok := chattype.ValidChartTypeMap[chattype.ChartType(record.Key)]; !ok {
				continue
			}
			answer = answer + "\n" + record.Value + "\n"
			llmchat.WriteStreamsResult(state.Ctx, state.QuestionEntity.Source.ReqID, llmchat.WriteResult{
				Content: record.Value,
				Flag:    llmchat.StreamsFlagECharts,
			})
		}
		state.QuestionEntity.Source.Answer = answer
		return nil
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[EChartReportNode] ProcessState error: %v", err)
		return nil, err
	}
	return input, nil
}
