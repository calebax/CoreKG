package middleware

import (
	"context"

	"github.com/insmtx/corekg/apps/kellm/drivers"
)

func StreamCapability(unsupportedErr error) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, chatCtx *drivers.ChatContext) (*drivers.ProxyResult, error) {
			if chatCtx != nil &&
				chatCtx.OriginalRequest != nil &&
				chatCtx.OriginalRequest.Stream &&
				chatCtx.ModelConfig != nil &&
				!chatCtx.ModelConfig.Capabilities.Stream {
				return nil, unsupportedErr
			}
			return next(ctx, chatCtx)
		}
	}
}
