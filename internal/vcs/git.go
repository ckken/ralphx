package vcs

import (
	"bytes"
	"errors"
	"os/exec"
	"sort"
	"strings"
)

type Snapshot struct {
	InsideRepo bool
	Status     string
}

func CaptureStatusSnapshot(workdir string) (Snapshot, error) {
	if _, err := exec.LookPath("git"); err != nil {
		var notFound *exec.Error
		if errors.As(err, &notFound) {
			return Snapshot{}, nil
		}
		return Snapshot{}, err
	}

	insideCmd := exec.Command("git", "-C", workdir, "rev-parse", "--is-inside-work-tree")
	if err := insideCmd.Run(); err != nil {
		return Snapshot{}, nil
	}

	statusCmd := exec.Command("git", "-C", workdir, "status", "--short")
	output, err := statusCmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Snapshot{}, nil
		}
		return Snapshot{}, err
	}

	return Snapshot{
		InsideRepo: true,
		Status:     sortLines(output),
	}, nil
}

func sortLines(output []byte) string {
	trimmed := strings.TrimSpace(string(bytes.TrimSpace(output)))
	if trimmed == "" {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}
