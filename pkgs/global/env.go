package global

import (
	"os"
	"strings"
)

const (
	// EnvEnableNebulaGraph 是否启用Nebula Graph，默认启用，值为 true
	EnvEnableNebulaGraph = "ENABLE_NEBULA_GRAPH"
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
