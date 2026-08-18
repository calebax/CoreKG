package svcforestchat

import (
	"encoding/json"
	"strings"

	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
)

func filterVisibleAnswer(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	lines := strings.Split(trimmed, "\n")
	var builder strings.Builder
	parsedAny := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item llmchat.WriteResult
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		parsedAny = true
		if item.Flag == "" {
			builder.WriteString(item.Content)
		}
	}

	return builder.String(), parsedAny
}

func normalizeAnswer(raw string) string {
	if filtered, ok := filterVisibleAnswer(raw); ok {
		return filtered
	}
	return raw
}
