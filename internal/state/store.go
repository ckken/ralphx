package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ckken/ralphx/internal/contracts"
)

func (p Paths) Ensure() error {
	if err := os.MkdirAll(p.Root, defaultDirMode); err != nil {
		return err
	}
	if err := os.MkdirAll(p.LogDir, defaultDirMode); err != nil {
		return err
	}
	return nil
}

func (p Paths) LastResultFile() string {
	return p.LastJSONFile
}

func WriteSummary(p Paths, summary string) error {
	return writeTextFile(p.SummaryFile, summary)
}

func WriteLastOutput(p Paths, output string) error {
	return writeTextFile(p.LastOutputFile, output)
}

func WriteLastResult(p Paths, result contracts.RoundResult) error {
	return writeJSONFile(p.LastJSONFile, result)
}

func WriteState(p Paths, iteration int, result contracts.RoundResult) error {
	return WriteStateAt(p, iteration, result, time.Now())
}

func WriteStateAt(p Paths, iteration int, result contracts.RoundResult, now time.Time) error {
	runState := RunState{
		Iteration: iteration,
		UpdatedAt: formatTimestamp(now),
		Result:    result,
	}
	return writeJSONFile(p.StateFile, runState)
}

func WriteStats(p Paths, stats Stats) error {
	return writeJSONFile(p.StatsFile, stats)
}

func WriteStatsAt(p Paths, stats Stats, now time.Time) error {
	if stats.UpdatedAt == "" {
		stats.UpdatedAt = formatTimestamp(now)
	}
	return WriteStats(p, stats)
}

func formatTimestamp(t time.Time) string {
	return t.Format(timestampLayout)
}

func writeTextFile(path, contents string) error {
	if err := os.MkdirAll(filepath.Dir(path), defaultDirMode); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(contents), defaultFileMode)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), defaultDirMode); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, defaultFileMode)
}
