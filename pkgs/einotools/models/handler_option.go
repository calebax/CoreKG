package models

type HandlerOptions struct {
	Debug         bool
	SummaryMode   bool
	ReactStream   *bool
	SummaryStream *bool
}
type HandlerOption func(*HandlerOptions)

func WithSummaryMode(mode bool) HandlerOption {
	return func(o *HandlerOptions) {
		o.SummaryMode = mode
	}
}

func WithAgentStageStreamMode(reactStream, summaryStream bool) HandlerOption {
	return func(o *HandlerOptions) {
		o.ReactStream = &reactStream
		o.SummaryStream = &summaryStream
	}
}
