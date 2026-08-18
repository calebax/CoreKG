package middleware

import (
	"context"

	"github.com/insmtx/corekg/apps/kellm/drivers"
)

type Handler func(ctx context.Context, chatCtx *drivers.ChatContext) (*drivers.ProxyResult, error)

type Middleware func(next Handler) Handler

func Chain(middlewares ...Middleware) func(Handler) Handler {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func InvokeDriverHandler() Handler {
	return func(ctx context.Context, chatCtx *drivers.ChatContext) (*drivers.ProxyResult, error) {
		return chatCtx.Driver.ChatCompletions(ctx, chatCtx)
	}
}
