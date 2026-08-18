package svckellm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/insmtx/corekg/apps/kellm/drivers"
	_ "github.com/insmtx/corekg/apps/kellm/drivers/openai"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
	svcmiddleware "github.com/insmtx/corekg/apps/kellm/services/svckellm/middleware"
	"github.com/ygpkg/yg-go/settings"
)

func ExtractBearerToken(header http.Header) (string, error) {
	authHeader := strings.TrimSpace(header.Get("Authorization"))
	if authHeader == "" {
		return "", ErrMissingAuthorizationHeader
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
	if token == "" {
		return "", ErrEmptyAuthorizationToken
	}
	return token, nil
}

func ParseChatRequest(payload []byte) (*kellmtype.ChatRequestBody, error) {
	req := &kellmtype.ChatRequestBody{}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequestBody, err)
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, ErrModelRequired
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("%w: messages is required", ErrInvalidRequestBody)
	}
	return req, nil
}

func ProxyChatCompletions(ctx context.Context, header http.Header, payload []byte) (*ProxyResult, error) {
	token, err := ExtractBearerToken(header)
	if err != nil {
		return nil, err
	}

	req, err := ParseChatRequest(payload)
	if err != nil {
		return nil, err
	}

	modelCfg, err := GetModelConfig(req.Model)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(modelCfg.BaseURL) == "" {
		return nil, ErrModelURLRequired
	}

	driver, err := drivers.GetDriver(modelCfg.ModelType)
	if err != nil {
		return nil, ErrUnsupportedModelType
	}

	chatCtx := &drivers.ChatContext{
		Token:           token,
		Header:          header,
		ModelConfig:     modelCfg,
		OriginalRequest: req,
		UpstreamRequest: req,
		Driver:          driver,
	}

	handler := svcmiddleware.Chain(
		svcmiddleware.StreamCapability(ErrStreamNotSupported),
		svcmiddleware.ToolCallFallback(),
	)(svcmiddleware.InvokeDriverHandler())

	return handler(ctx, chatCtx)
}

func GetModelConfig(model string) (*kellmtype.ModelConfig, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, ErrModelRequired
	}

	var cfgs []kellmtype.ModelConfig
	if err := settings.GetYaml(settingGroupKeLLM, settingKeyModels, &cfgs); err != nil {
		return nil, err
	}

	for i := range cfgs {
		if strings.EqualFold(strings.TrimSpace(cfgs[i].ModelName), model) {
			cfgs[i].ModelType = normalizeModelType(cfgs[i].ModelType)
			return &cfgs[i], nil
		}
	}

	return nil, ErrModelNotFound
}

func normalizeModelType(modelType kellmtype.ModelType) kellmtype.ModelType {
	switch strings.ToLower(strings.TrimSpace(string(modelType))) {
	case "", string(kellmtype.ModelTypeOpenAI):
		return kellmtype.ModelTypeOpenAI
	default:
		return kellmtype.ModelType(strings.ToLower(strings.TrimSpace(string(modelType))))
	}
}
