package middleware

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/kellm/drivers"
)

func readProxyResultBody(result *drivers.ProxyResult) ([]byte, error) {
	if result == nil {
		return nil, nil
	}
	if result.BodyReader == nil {
		return result.Body, nil
	}
	defer result.BodyReader.Close()
	return io.ReadAll(result.BodyReader)
}

func generateToolCallID() string {
	return fmt.Sprintf("call_%s", uuid.New().String()[:24])
}

func generateChatCompletionID() string {
	return fmt.Sprintf("chatcmpl-%s", uuid.New().String())
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func defaultInt64(v, fallback int64) int64 {
	if v == 0 {
		return fallback
	}
	return v
}
