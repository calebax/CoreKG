package global

import (
	"os"
	"strings"
)

const (
	// EnvEnableNebulaGraph 是否启用Nebula Graph，默认启用，值为 true
	EnvEnableNebulaGraph = "ENABLE_NEBULA_GRAPH"
	// EnvEnableLicenseCheck 是否启用license校验，默认不启用，值为 true 时启用
	EnvEnableLicenseCheck = "ENABLE_LICENSE_CHECK"
)

func GetEnableNebulaGraphBool() bool {
	val := os.Getenv(EnvEnableNebulaGraph)
	if val == "" {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "false", "0", "no", "off":
		return false
	default:
		return true
	}
}

// GetEnableLicenseCheckBool 是否启用license校验，默认不启用。
// 只有当环境变量 ENABLE_LICENSE_CHECK 被显式设置为 true 时才启用。
func GetEnableLicenseCheckBool() bool {
	val := os.Getenv(EnvEnableLicenseCheck)
	if val == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}
