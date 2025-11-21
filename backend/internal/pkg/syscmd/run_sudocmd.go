package syscmd

import (
	"context"
	"os"
	"os/exec"
	"time"
)

// RunSudoCmd runs a command with sudo and streams stdout/stderr to terminal.
// Supports interactive commands (prompts for password if needed).
func RunSudoCmd(ctx context.Context, name string, args ...string) error {
	allArgs := append([]string{name}, args...)
	cmd := exec.CommandContext(ctx, "sudo", allArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // allows interactive password input
	return cmd.Run()
}

// RunSudoShell runs a shell command string with sudo (supports pipes, redirects, variables, &&, ||)
// and streams stdout/stderr live.
func RunSudoShell(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sudo", "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunSudoOutput runs a command with sudo and returns combined stdout/stderr as string.
func RunSudoOutput(ctx context.Context, name string, args ...string) (string, error) {
	allArgs := append([]string{name}, args...)
	cmd := exec.CommandContext(ctx, "sudo", allArgs...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunSudoShellOutput runs a shell command string with sudo and returns combined stdout/stderr.
func RunSudoShellOutput(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "sudo", "sh", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// RunSudoCmdWithTimeout runs a sudo command with timeout
func RunSudoCmdWithTimeout(timeout time.Duration, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunSudoCmd(ctx, name, args...)
}

// RunSudoShellWithTimeout runs a sudo shell command with timeout
func RunSudoShellWithTimeout(timeout time.Duration, command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunSudoShell(ctx, command)
}
