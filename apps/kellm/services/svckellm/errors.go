package svckellm

import (
	"errors"

	"github.com/insmtx/corekg/apps/kellm/drivers"
)

var (
	ErrMissingAuthorizationHeader = errors.New("missing Authorization header")
	ErrEmptyAuthorizationToken    = errors.New("empty Authorization token")
	ErrInvalidRequestBody         = errors.New("invalid request body")
	ErrModelRequired              = errors.New("model is required")
	ErrModelURLRequired           = errors.New("model_url is required")
	ErrModelNotFound              = errors.New("model config not found")
	ErrUnsupportedModelType       = drivers.ErrUnsupportedModelType
	ErrStreamNotSupported         = errors.New("stream is not supported")
)

const (
	settingGroupKeLLM = "knowledge"
	settingKeyModels  = "proxy_llm_models"
)
