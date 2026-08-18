package jobs

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

var resultNC *nats.Conn

func InitResultConsumer(nc *nats.Conn) {
	resultNC = nc
}

func RunRoutines(ctx context.Context) error {
	if resultNC == nil {
		logs.WarnContextf(ctx, "[jobs] NATS 连接未设置，跳过 result consumer")
		return nil
	}

	if err := StartForestResultConsumers(ctx, resultNC); err != nil {
		return err
	}

	if err := StartGraphResultConsumers(ctx, resultNC); err != nil {
		return err
	}

	return nil
}
