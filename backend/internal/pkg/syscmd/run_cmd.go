package syscmd

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// RunCmd runs a command with args and streams stdout/stderr to terminal.
// Supports interactive commands.
func RunCmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunShell runs a shell command string (supports pipes, redirects, variables, &&, ||)
// and streams stdout/stderr live.
func RunShell(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunOutput runs a command and returns its combined stdout and stderr as string.
func RunOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunShellOutput runs a shell command and returns its combined stdout and stderr as string.
func RunShellOutput(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunCmdWithTimeout runs a command with a timeout duration. If the command exceeds the timeout,
// it will be automatically killed.
func RunCmdWithTimeout(timeout time.Duration, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunCmd(ctx, name, args...)
}

// RunShellWithTimeout runs a shell command with a timeout. Supports pipes, redirects, etc.
func RunShellWithTimeout(timeout time.Duration, command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunShell(ctx, command)
}
