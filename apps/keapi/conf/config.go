package conf

import (
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

type mcpConfig struct {
	Addr string `yaml:"addr"`
}

type appConfig struct {
	MCP mcpConfig `yaml:"mcp"`
}

var MCPCfg = mcpConfig{
	Addr: "http://localhost:8086",
}

func InitMCPCfg(configFile string) {
	var ac appConfig
	if err := config.LoadYamlLocalFile(configFile, &ac); err != nil {
		logs.WarnContextf(nil, "[mcp] load mcp config failed, use default: %v", err)
		return
	}
	if ac.MCP.Addr == "" {
		logs.Warn("[mcp] mcp.addr is empty in config, use default")
		return
	}
	MCPCfg = ac.MCP
}
