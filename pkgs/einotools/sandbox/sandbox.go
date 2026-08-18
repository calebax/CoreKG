package sandbox

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
)

// ExecResult 执行结果
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// SyntaxCheckResult 语法校验结果
type SyntaxCheckResult struct {
	Valid    bool
	Stdout   string
	Stderr   string
	ExitCode int
}

type Sandbox interface {
	Exec(ctx context.Context, lang string, code string) (*ExecResult, error)
	CheckSyntax(ctx context.Context, lang string, code string) (*SyntaxCheckResult, error)
}

// 执行策略
type SandboxMode string

const (
	SandboxModeLocalCommand SandboxMode = "local_command"
	SandboxModeRemoteHTTP   SandboxMode = "remote_http"
	SandboxModeAuto         SandboxMode = "auto"
)

type Config struct {
	Mode    SandboxMode `json:"mode"`
	Timeout int         `json:"timeout"`
	// Remote 相关
	HttpBaseURL string `json:"httpBaseUrl" yaml:"http_base_url"` // 请求地址
	HttpToken   string `json:"httpToken" yaml:"http_token"`      // 可选鉴权
}

func NewSandbox(cfg *Config) (Sandbox, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg = mergeWithEnv(cfg)

	mode := cfg.Mode
	if mode == "" || mode == SandboxModeAuto {
		mode = autoDetectMode()
	}

	switch mode {
	case SandboxModeLocalCommand:
		return NewLocalCommandSandbox(cfg)
	case SandboxModeRemoteHTTP:
		return NewRemoteHTTPSandbox(cfg)
	default:
		return nil, errors.New("unknown sandbox mode: " + string(mode))
	}
}

// autoDetectMode 自动判断执行策略
func autoDetectMode() SandboxMode {
	defaultSandboxMode := os.Getenv("SANDBOX_MODE")
	v := strings.ToLower(strings.TrimSpace(defaultSandboxMode))

	switch v {
	case string(SandboxModeLocalCommand):
		return SandboxModeLocalCommand
	case string(SandboxModeRemoteHTTP):
		return SandboxModeRemoteHTTP
	// case string(SandboxModeAuto):
	// TODO 默认策略或是根据环境变量
	default:
		return SandboxModeLocalCommand
	}
}

var (
	defaultOnce       sync.Once
	defaultSandbox    Sandbox
	defaultSandboxErr error
)

// DefaultSandbox 返回按环境自动选择的单例 Sandbox
func DefaultSandbox() (Sandbox, error) {
	defaultOnce.Do(func() {
		cfg := mergeWithEnv(&Config{Mode: SandboxModeAuto})
		defaultSandbox, defaultSandboxErr = NewSandbox(cfg)
	})
	return defaultSandbox, defaultSandboxErr
}

// 动态获取配置
func mergeWithEnv(in *Config) *Config {
	out := *in
	// Mode
	if out.Mode == "" || out.Mode == SandboxModeAuto {
		out.Mode = autoDetectMode()
	}
	// Timeout
	if out.Timeout == 0 {
		out.Timeout = 120
	}

	// TODO Remote
	if out.HttpBaseURL == "" {
	}
	if out.HttpToken == "" {
	}
	return &out
}
