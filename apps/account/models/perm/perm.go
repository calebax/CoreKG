package perm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/insmtx/corekg/apps/kechat/models/agentperm"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
)

// Set is a set include forestPermSet and chatPermSet
type Set struct {
	ForestPs []*perm.Set          `json:"forestPs"`
	ChatPs   []*agentperm.PermSet `json:"chatPs"`
}

// Value 实现 driver.Valuer 接口，用于将 PermSet 转换为数据库可存储的值
func (p *Set) Value() (driver.Value, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PermSet to JSON: %w", err)
	}
	return string(jsonData), nil
}

func (p *Set) Scan(src any) error {
	if src == nil {
		return nil
	}
	var source []byte
	switch s := src.(type) {
	case string:
		source = []byte(s)
	case []byte:
		source = s
	case nil:
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, p)
	}
	if len(source) == 0 {
		return nil
	}
	err := json.Unmarshal(source, p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to PermSet: %w", err)
	}
	return nil
}
