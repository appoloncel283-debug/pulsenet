package core

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

func TraceRoute(ctx context.Context, host string, maxHops int, timeout time.Duration) (string, error) {
	if maxHops < 1 || maxHops > 64 {
		return "", fmt.Errorf("max hops must be between 1 and 64")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command = "tracert"
		args = []string{"-d", "-h", strconv.Itoa(maxHops), host}
	default:
		command = "traceroute"
		args = []string{"-n", "-m", strconv.Itoa(maxHops), host}
	}
	binary, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("%s is not installed or not available in PATH", command)
	}
	output, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
	if ctx.Err() != nil {
		return string(output), fmt.Errorf("trace timed out: %w", ctx.Err())
	}
	if err != nil {
		return string(output), fmt.Errorf("trace command failed: %w", err)
	}
	return string(output), nil
}
