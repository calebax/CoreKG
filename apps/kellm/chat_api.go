package kellm

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kellm/services/svckellm"
	"github.com/ygpkg/yg-go/logs"
)

// ChatCompletions handles OpenAI-compatible chat/completions requests.
func ChatCompletions(ctx *gin.Context) {
	reqBodyData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		writeProxyResult(ctx, svckellm.NewOpenAIErrorResult(
			http.StatusBadRequest,
			"failed to read request body",
			"invalid_request_error",
			"read_request_body_failed",
		))
		return
	}

	result, err := svckellm.ProxyChatCompletions(ctx.Request.Context(), ctx.Request.Header, reqBodyData)
	if err != nil {
		logs.ErrorContextf(ctx, "kellm request model api failed: %v", err)
		switch {
		case errors.Is(err, svckellm.ErrMissingAuthorizationHeader), errors.Is(err, svckellm.ErrEmptyAuthorizationToken):
			result = svckellm.NewOpenAIErrorResult(http.StatusUnauthorized, "authorization token is required", "invalid_request_error", "authorization_required")
		case errors.Is(err, svckellm.ErrInvalidRequestBody):
			result = svckellm.NewOpenAIErrorResult(http.StatusBadRequest, "invalid request body", "invalid_request_error", "invalid_request_body")
		case errors.Is(err, svckellm.ErrModelRequired):
			result = svckellm.NewOpenAIErrorResult(http.StatusBadRequest, "model is required", "invalid_request_error", "model_required")
		case errors.Is(err, svckellm.ErrModelNotFound):
			result = svckellm.NewOpenAIErrorResult(http.StatusBadRequest, "model config not found", "invalid_request_error", "model_not_found")
		case errors.Is(err, svckellm.ErrModelURLRequired):
			result = svckellm.NewOpenAIErrorResult(http.StatusBadRequest, "model base_url is required", "invalid_request_error", "model_url_required")
		case errors.Is(err, svckellm.ErrUnsupportedModelType):
			result = svckellm.NewOpenAIErrorResult(http.StatusBadRequest, "unsupported model type", "invalid_request_error", "unsupported_model_type")
		case errors.Is(err, svckellm.ErrStreamNotSupported):
			result = svckellm.NewOpenAIErrorResult(http.StatusBadRequest, "stream is not supported by model config", "invalid_request_error", "stream_not_supported")
		default:
			result = svckellm.NewOpenAIErrorResult(http.StatusBadGateway, "upstream model request failed", "server_error", "upstream_request_failed")
		}
	}

	writeProxyResult(ctx, result)
}

func writeProxyResult(ctx *gin.Context, result *svckellm.ProxyResult) {
	if result == nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	for key, values := range result.Header {
		for _, value := range values {
			ctx.Writer.Header().Add(key, value)
		}
	}
	ctx.Status(result.StatusCode)
	if result.BodyReader != nil {
		defer result.BodyReader.Close()
		if err := copyStreamBody(ctx.Writer, result.BodyReader); err != nil {
			logs.ErrorContextf(ctx, "kellm stream response failed: %v", err)
			return
		}
		return
	}
	if len(result.Body) == 0 {
		return
	}
	if _, err := ctx.Writer.Write(result.Body); err != nil {
		logs.ErrorContextf(ctx, "kellm write response failed: %v", err)
	}
	if flusher, ok := ctx.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func copyStreamBody(w http.ResponseWriter, r io.Reader) error {
	reader := bufio.NewReader(r)
	flusher, _ := w.(http.Flusher)

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			if _, writeErr := w.Write(line); writeErr != nil {
				return writeErr
			}
			if flusher != nil && len(bytes.TrimSpace(line)) == 0 {
				flusher.Flush()
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				if flusher != nil {
					flusher.Flush()
				}
				return nil
			}
			return err
		}
	}
}
