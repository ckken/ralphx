package legacy

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func RunScript(scriptName string, args []string) error {
	scriptPath, err := FindRepoFile(scriptName)
	if err != nil {
		return err
	}

	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	return cmd.Run()
}

func FindRepoFile(name string) (string, error) {
	candidates := []string{}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, walkUpCandidates(cwd, name)...)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, walkUpCandidates(filepath.Dir(exe), name)...)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not locate %s from current directory or executable path", name)
}

func walkUpCandidates(start, name string) []string {
	out := []string{}
	current := start
	for {
		out = append(out, filepath.Join(current, name))
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return out
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}
