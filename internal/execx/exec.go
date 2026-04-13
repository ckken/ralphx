package execx

import (
	"bytes"
	"context"
	"os/exec"
)

type Result struct {
	Output   []byte
	ExitCode int
}

func Run(ctx context.Context, name string, args []string, stdin []byte, dir string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	output, err := cmd.CombinedOutput()
	result := Result{Output: output, ExitCode: 0}
	if err == nil {
		return result, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, err
	}
	result.ExitCode = 1
	return result, err
}
