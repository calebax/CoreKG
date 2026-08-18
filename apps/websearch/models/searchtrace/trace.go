// Package searchtrace emits request-scoped routing observations through the roc logger.
package searchtrace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/ygpkg/yg-go/logs"
)

// Event records one request-scoped search routing observation.
type Event struct {
	TraceID           string         `json:"trace_id"`
	RequestID         string         `json:"request_id"`
	Sequence          uint64         `json:"event_sequence"`
	Type              string         `json:"event_type"`
	OccurredAt        time.Time      `json:"occurred_at"`
	RequestedProvider string         `json:"requested_provider,omitempty"`
	Provider          string         `json:"provider,omitempty"`
	ProfileID         string         `json:"profile_id,omitempty"`
	LeaseID           string         `json:"lease_id,omitempty"`
	Classification    string         `json:"classification,omitempty"`
	QueryHash         string         `json:"query_hash,omitempty"`
	QueryLength       int            `json:"query_length,omitempty"`
	QueryPreview      string         `json:"query_preview,omitempty"`
	Query             string         `json:"query,omitempty"`
	Fields            map[string]any `json:"fields,omitempty"`
}

// Config controls diagnostic detail and query redaction.
type Config struct {
	Diagnostics  bool
	StoreQuery   bool
	PreviewChars int
	Now          func() time.Time
}

// Manager sequences and emits request-scoped search observations.
type Manager struct {
	config    Config
	mu        sync.Mutex
	sequences map[string]uint64
}

// New creates a search trace manager.
func New(config Config) *Manager {
	if config.PreviewChars <= 0 {
		config.PreviewChars = 32
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	return &Manager{config: config, sequences: make(map[string]uint64)}
}

// Append sequences and writes one trace event through the roc logger.
func (manager *Manager) Append(ctx context.Context, event Event) error {
	manager.mu.Lock()
	manager.sequences[event.TraceID]++
	if event.Sequence == 0 {
		event.Sequence = manager.sequences[event.TraceID]
	}
	manager.mu.Unlock()
	if event.OccurredAt.IsZero() {
		event.OccurredAt = manager.config.Now().UTC()
	}
	isSummary := event.Type == "request_finished" || event.Type == "request_failed"
	if !manager.config.Diagnostics && !isSummary {
		return nil
	}
	attributes := []any{"event", event.Type, "request_id", event.RequestID, "trace_id", event.TraceID, "sequence", event.Sequence, "requested_provider", event.RequestedProvider, "provider", event.Provider, "profile_id", event.ProfileID, "lease_id", event.LeaseID, "classification", event.Classification, "query_hash", event.QueryHash, "query_length", event.QueryLength, "query_preview", event.QueryPreview, "fields", event.Fields}
	if manager.config.StoreQuery && event.Query != "" {
		attributes = append(attributes, "query", event.Query)
	}
	if isSummary {
		logs.InfoContextw(ctx, "search trace", attributes...)
	} else {
		logs.DebugContextw(ctx, "search trace", attributes...)
	}
	return nil
}

// QueryMetadata returns the redacted metadata allowed by the configured policy.
func (manager *Manager) QueryMetadata(query string) (hash string, length int, preview string, stored string) {
	digest := sha256.Sum256([]byte(query))
	hash = hex.EncodeToString(digest[:])
	runes := []rune(strings.TrimSpace(query))
	length = len(runes)
	if len(runes) > manager.config.PreviewChars {
		runes = runes[:manager.config.PreviewChars]
	}
	for index, value := range runes {
		if value >= '0' && value <= '9' {
			runes[index] = '*'
		}
	}
	preview = string(runes)
	if manager.config.StoreQuery {
		stored = query
	}
	return
}
