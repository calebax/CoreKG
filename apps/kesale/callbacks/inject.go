package callbacks

import "context"

func OnPaymentCompletedHandler(ctx context.Context,
	payInfo *PayInfo, handlers []Handler) context.Context {

	for i := len(handlers) - 1; i >= 0; i-- {
		newCtx := handlers[i].OnPaymentCompleted(ctx, payInfo)
		if newCtx != nil {
			ctx = newCtx
		}
	}

	return ctx
}
