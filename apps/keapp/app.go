package keapp

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/insmtx/corekg/apps/keapp/internal/apis"
	"github.com/insmtx/corekg/apps/keapp/services/svcweb"
	"github.com/insmtx/corekg/apps/keapp/worker"
	"github.com/ygpkg/yg-go/apis/runtime/server"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

func Routers(eng *server.Router) error {
	apis.RegistryRouter(eng)
	return nil
}

func Migrates(db *gorm.DB) error {
	return nil
}

var onceStart sync.Once

func RunJob() error {
	onceStart.Do(func() {
	})
	return nil
}

func StartWorker(ctx context.Context, nc *nats.Conn, cfg worker.Config) {
	svcweb.SetNATSConn(nc)
	go func() {
		if err := worker.Start(ctx, nc, cfg); err != nil {
			logs.ErrorContextf(ctx, "[keapp] start worker failed: %v", err)
		}
	}()
}
