package apis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygpkg/yg-go/apis/runtime/middleware"
	"github.com/ygpkg/yg-go/apis/runtime/server"
)

func TestRegistryRouterProtectsFetchAction(t *testing.T) {
	reader := &fakeReader{}
	handler, err := NewHandler(HandlerOptions{Reader: reader})
	if err != nil {
		t.Fatal(err)
	}
	router := server.NewRouter("/v3/", server.WithMiddleware(middleware.GenerateRequestID(nil)))
	RegistryRouter(router, handler, "yg-test-token")

	unauthorized := httptest.NewRecorder()
	router.GinEngine().ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/v3/webfetch.Fetch", strings.NewReader(`{}`)))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	authorizedRequest := httptest.NewRequest(http.MethodPost, "/v3/webfetch.Fetch", strings.NewReader(`{"url":"https://example.com"}`))
	authorizedRequest.Header.Set("Authorization", "Bearer yg-test-token")
	authorized := httptest.NewRecorder()
	router.GinEngine().ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK || !strings.Contains(authorized.Body.String(), `"final_url":"https://example.com"`) {
		t.Fatalf("authorized response = %d %s", authorized.Code, authorized.Body.String())
	}
}
