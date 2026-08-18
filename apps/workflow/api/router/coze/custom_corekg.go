package coze

import (
	"github.com/cloudwego/hertz/pkg/app/server"

	cozehandler "github.com/insmtx/corekg/apps/workflow/api/handler/coze"
)

// RegisterCorekgRoutes registers hand-written corekg related routes.
func RegisterCorekgRoutes(r *server.Hertz) {
	r.POST("/api/internal/space_sync", cozehandler.SpaceSyncWebhook)
	r.POST("/api/internal/agent/external_info", cozehandler.GetAgentShortLinkCode)
	r.POST("/api/internal/agent/set_external_status", cozehandler.SetAgentExternalStatus)
	r.POST("/api/permission_api/pat/get_personal_access_token", cozehandler.GetOrCreateCurrentUserAPIKey)
	r.POST("/api/public/agent/external_token", cozehandler.PublicGetAgentUserIDByShortCode)
	r.POST("/api/playground_api/produce/create_bot", corekgCreateResourcePermissionMw(), cozehandler.ProduceCreateBot)
	r.POST("/api/playground_api/bot_config/create", cozehandler.CreateBotConfig)
}
