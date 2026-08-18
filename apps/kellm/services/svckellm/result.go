package svckellm

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/insmtx/corekg/apps/kellm/drivers"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

type ProxyResult = drivers.ProxyResult

func NewOpenAIErrorResult(statusCode int, message, errorType, code string) *ProxyResult {
	body, _ := json.Marshal(kellmtype.OpenAIErrorEnvelope{
		Error: kellmtype.OpenAIError{
			Message: message,
			Type:    errorType,
			Code:    code,
		},
	})
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	return &ProxyResult{
		StatusCode: statusCode,
		Header:     header,
		Body:       body,
	}
}
