package drivers

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

type ProxyResult struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	BodyReader io.ReadCloser
	Stream     bool
}

type ChatContext struct {
	Token                string
	Header               http.Header
	ModelConfig          *kellmtype.ModelConfig
	OriginalRequest      *kellmtype.ChatRequestBody
	UpstreamRequest      *kellmtype.ChatRequestBody
	Driver               Driver
	NeedToolCallFallback bool
}

type Driver interface {
	Type() kellmtype.ModelType
	ChatCompletions(ctx context.Context, chatCtx *ChatContext) (*ProxyResult, error)
}

var registeredDrivers = map[kellmtype.ModelType]Driver{}

func RegisterDriver(driver Driver) {
	if driver == nil {
		return
	}
	registeredDrivers[normalizeModelType(driver.Type())] = driver
}

func GetDriver(modelType kellmtype.ModelType) (Driver, error) {
	driver, ok := registeredDrivers[normalizeModelType(modelType)]
	if !ok {
		return nil, ErrUnsupportedModelType
	}
	return driver, nil
}

func normalizeModelType(modelType kellmtype.ModelType) kellmtype.ModelType {
	switch strings.ToLower(strings.TrimSpace(string(modelType))) {
	case "", string(kellmtype.ModelTypeOpenAI):
		return kellmtype.ModelTypeOpenAI
	default:
		return kellmtype.ModelType(strings.ToLower(strings.TrimSpace(string(modelType))))
	}
}
