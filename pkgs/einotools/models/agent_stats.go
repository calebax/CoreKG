package models

import "time"

type AgentStats struct {
	ModelUsages    []ModelUsage `json:"model_usages"`
	TotalUsage     Usage        `json:"total_usage"`
	DurationMs     int64        `json:"duration_ms"`
	StartTimestamp int64        `json:"start_timestamp"`
	EndTimestamp   int64        `json:"end_timestamp"`
}

func (a *AgentStats) Start() {
	a.StartTimestamp = time.Now().UnixMilli()
}

func (a *AgentStats) Stop() {
	a.EndTimestamp = time.Now().UnixMilli()
	a.DurationMs = a.EndTimestamp - a.StartTimestamp
}

func (a *AgentStats) AddModelUsage(modelUsage ModelUsage) {
	a.ModelUsages = append(a.ModelUsages, modelUsage)
}

func (a *AgentStats) AddTotalUsage(usage *Usage) {
	a.TotalUsage.PromptTokens += usage.PromptTokens
	a.TotalUsage.CompletionTokens += usage.CompletionTokens
	a.TotalUsage.TotalTokens += usage.TotalTokens
	a.TotalUsage.PromptCacheHitTokens += usage.PromptCacheHitTokens
	a.TotalUsage.PromptCacheMissTokens += usage.PromptCacheMissTokens
}

type ModelUsage struct {
	ModelName   string `json:"model_name"`
	ModelVendor string `json:"model_vendor"`
	Usage       Usage  `json:"usage"`
}

type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}
