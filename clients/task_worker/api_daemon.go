package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	daemonRunnerCmd string
)

func withDaemonAPI(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&daemonRunnerCmd, "daemon_runner_cmd", "d", os.Getenv("DAEMON_RUNNER_CMD"), "Daemon runner command to execute")
}

func runDaemonAPI(ctx context.Context) (*DaemonAPI, error) {
	if daemonRunnerCmd == "" {
		return nil, fmt.Errorf("daemon runner command is not set, please use --daemon_runner_cmd to set it")
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", daemonRunnerCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start daemon process: %w", err)
	}

	return nil, nil
}

// DaemonAPI provides methods to interact with the daemon service.
type DaemonAPI struct {
	ctx     context.Context
	process *exec.Cmd
}

func (d *DaemonAPI) Stop() error {
	if d.process != nil && d.process.Process != nil {
		if err := d.process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop daemon process: %w", err)
		}
	}
	return nil
}

func (d *DaemonAPI) IsRunning() bool {
	if d.process == nil || d.process.Process == nil {
		return false
	}
	if err := d.process.Process.Signal(os.Interrupt); err != nil {
		return false // Process is not running
	}
	return true // Process is running
}
