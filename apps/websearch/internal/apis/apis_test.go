package apis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/insmtx/corekg/apps/websearch/models/cursor"
	"github.com/ygpkg/yg-go/apis/runtime/middleware"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func TestRegistryRouterProtectsSearchAction(t *testing.T) {
	searcher := &fakeSearcher{}
	codec, err := cursor.New("secret", time.Minute, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(HandlerOptions{Searcher: searcher, Cursor: codec, EnabledProviders: []string{"bing"}, ProviderVisibility: "public"})
	if err != nil {
		t.Fatal(err)
	}
	router := server.NewRouter("/v3/", server.WithMiddleware(middleware.GenerateRequestID(nil)))
	RegistryRouter(router, handler, "yg-test-token")

	unauthorized := httptest.NewRecorder()
	router.GinEngine().ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/v3/websearch.Search", strings.NewReader(`{}`)))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	authorizedRequest := httptest.NewRequest(http.MethodPost, "/v3/websearch.Search", strings.NewReader(`{"query":"go","limit":1}`))
	authorizedRequest.Header.Set("Authorization", "Bearer yg-test-token")
	authorized := httptest.NewRecorder()
	router.GinEngine().ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK || !strings.Contains(authorized.Body.String(), `"query":"go"`) {
		t.Fatalf("authorized response = %d %s", authorized.Code, authorized.Body.String())
	}
}
