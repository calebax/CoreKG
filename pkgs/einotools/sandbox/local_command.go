package sandbox

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

// / TODO 目前为本地命令执行，调用docker容器进程
type localCommandSandbox struct {
	timeout time.Duration
}

func NewLocalCommandSandbox(cfg *Config) (Sandbox, error) {
	return &localCommandSandbox{
		timeout: time.Duration(cfg.Timeout) * time.Second,
	}, nil
}

func (s *localCommandSandbox) Exec(ctx context.Context, lang string, code string) (*ExecResult, error) {
	// TODO 当前仅支持 python
	if lang == "" {
		lang = "python"
	}

	// docker exec -i py-sandbox python -
	cctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, "docker", "exec", "-i", "py-sandbox", "python", "-")
	cmd.Stdin = bytes.NewBufferString(code)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	exit := exitCodeFromError(err)

	res := &ExecResult{
		Stdout:   out.String(),
		Stderr:   stderr.String(),
		ExitCode: exit,
	}
	if err != nil {
		return res, err
	}
	return res, nil
}

func (s *localCommandSandbox) CheckSyntax(ctx context.Context, lang string, code string) (*SyntaxCheckResult, error) {
	if lang == "" {
		lang = "python"
	}

	// docker exec -i py-sandbox python check_syntax.py
	cctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, "docker", "exec", "-i", "py-sandbox", "python", "check_syntax.py")
	cmd.Stdin = bytes.NewBufferString(code)

	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	exit := exitCodeFromError(err)

	res := &SyntaxCheckResult{
		Valid:    exit == 0,
		Stdout:   out.String(),
		Stderr:   stderr.String(),
		ExitCode: exit,
	}
	if err != nil {
		return res, err
	}
	return res, nil
}

func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}
