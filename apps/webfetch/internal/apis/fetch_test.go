package apis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/webfetch/models/domain"
	"github.com/ygpkg/yg-go/apis/constants"
)

type fakeReader struct{ request domain.ReadRequest }

func newTestRouter(options HandlerOptions) (*gin.Engine, error) {
	handler, err := NewHandler(options)
	if err != nil {
		return nil, err
	}
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set(constants.CtxKeyRequestID, "req_test")
		ctx.Next()
	})
	router.POST("/v1/webfetch", handler.Fetch)
	return router, nil
}

func (fake *fakeReader) Read(_ context.Context, request domain.ReadRequest) (domain.ReadResponse, error) {
	fake.request = request
	return domain.ReadResponse{URL: request.URL, FinalURL: request.URL, Title: "Example", SourceType: domain.SourceTypeHTML, ContentType: "text/html", StatusCode: 200, Content: "# Example", ContentFormat: domain.OutputFormatMarkdown, ContentLength: 9, Meta: domain.ReadMeta{Transport: domain.ReadTransportHTTP, TookMS: 3}, Warnings: []domain.ReadWarning{}}, nil
}

func TestPOSTReadContract(t *testing.T) {
	reader := &fakeReader{}
	router, err := newTestRouter(HandlerOptions{Reader: reader})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/webfetch", strings.NewReader(`{"url":"https://example.com","timeout":"1500ms","output":{"format":"markdown","max_chars":30000}}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["request_id"] == "" {
		t.Fatal("request_id missing")
	}
	if reader.request.URL != "https://example.com" {
		t.Fatalf("url=%q", reader.request.URL)
	}
	document := body["document"].(map[string]any)
	if document["content_type"] != "text/html" || document["status_code"] != float64(200) {
		t.Fatalf("document=%v", document)
	}
	if body["usage"].(map[string]any)["units"] != float64(1) {
		t.Fatalf("usage=%v", body["usage"])
	}
}

func TestPOSTReadRejectsNonStrictJSON(t *testing.T) {
	reader := &fakeReader{}
	router, err := newTestRouter(HandlerOptions{Reader: reader})
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"url":"https://example.com","unknown":true}`,
		`{"url":"https://example.com"}{"url":"https://example.org"}`,
		`{"url":"https://example.com","timeout":"1.5s"}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/v1/webfetch", strings.NewReader(body))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body=%s status=%d, want %d", body, response.Code, http.StatusBadRequest)
		}
	}
}
