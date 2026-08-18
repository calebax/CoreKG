package printer

import "context"

type Printer interface {
	Send(ctx context.Context, msgId string, msgType string, msg any, isFinal bool)

	SendFinalMsg(ctx context.Context, msgType string, msg any)

	SetCancelFunc(cancel func())

	Close(ctx context.Context)
}
