package apis

import (
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

// RegistryRouter registers the authenticated WebSearch action.
func RegistryRouter(eng *server.Router, handler *Handler, apiKey string) {
	eng.P("websearch.Search", hasAPIKey(apiKey), handler.Search)
}
