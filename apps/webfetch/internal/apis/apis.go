package apis

import (
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

// RegistryRouter registers the authenticated WebFetch action.
func RegistryRouter(eng *server.Router, handler *Handler, apiKey string) {
	eng.P("webfetch.Fetch", hasAPIKey(apiKey), handler.Fetch)
}
