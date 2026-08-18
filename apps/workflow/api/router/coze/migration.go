package coze

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	cozehandler "github.com/insmtx/corekg/apps/workflow/api/handler/coze"
)

// RegisterMigrationRoutes registers hand-written migration routes.
func RegisterMigrationRoutes(r *server.Hertz) {
	r.POST("/api/internal/space_id_migration", cozehandler.MigrateResourceSpaceID)
}
