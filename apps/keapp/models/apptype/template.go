package apptype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type AppCapabilities struct {
	AIAssistant bool `json:"ai_assistant"`
	Search      bool `json:"search"`
	FAQ         bool `json:"faq"`
	Widget      bool `json:"widget"`
}

type WebsiteConfig struct {
	URL           string          `json:"url"`
	SyncSchedule  string          `json:"sync_schedule"`
	MaxDepth      int             `json:"max_depth"`
	MaxPages      int             `json:"max_pages"`
	RespectRobots bool            `json:"respect_robots"`
	Capabilities  AppCapabilities `json:"capabilities"`
}

type ProductConfig struct {
	ProductName  string          `json:"product_name"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type AftersalesConfig struct {
	ServiceScope string          `json:"service_scope"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type TrainingConfig struct {
	Department   string          `json:"department"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type PolicyConfig struct {
	OrgName      string          `json:"org_name"`
	Capabilities AppCapabilities `json:"capabilities"`
}

type AppConfig struct {
	Type   AppTemplateType `json:"type"`
	Config json.RawMessage `json:"config"`
}

func (c *AppConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *AppConfig) Scan(value interface{}) error {
	if value == nil {
		c.Type = ""
		c.Config = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan AppConfig: %T", value)
	}
	return json.Unmarshal(bytes, c)
}

func (c *AppConfig) AsWebsite() (*WebsiteConfig, error) {
	var cfg WebsiteConfig
	err := json.Unmarshal(c.Config, &cfg)
	return &cfg, err
}
func (c *AppConfig) AsProduct() (*ProductConfig, error) {
	var cfg ProductConfig
	err := json.Unmarshal(c.Config, &cfg)
	return &cfg, err
}
func (c *AppConfig) AsAftersales() (*AftersalesConfig, error) {
	var cfg AftersalesConfig
	err := json.Unmarshal(c.Config, &cfg)
	return &cfg, err
}
func (c *AppConfig) AsTraining() (*TrainingConfig, error) {
	var cfg TrainingConfig
	err := json.Unmarshal(c.Config, &cfg)
	return &cfg, err
}
func (c *AppConfig) AsPolicy() (*PolicyConfig, error) {
	var cfg PolicyConfig
	err := json.Unmarshal(c.Config, &cfg)
	return &cfg, err
}
