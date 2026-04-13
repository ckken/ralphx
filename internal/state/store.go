package state

import (
	"encoding/json"
	"errors"
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
	return WriteStateWithGuidanceAt(p, iteration, result, nil, time.Now())
}

func WriteStateAt(p Paths, iteration int, result contracts.RoundResult, now time.Time) error {
	return WriteStateWithGuidanceAt(p, iteration, result, nil, now)
}

func WriteStateWithGuidance(p Paths, iteration int, result contracts.RoundResult, guidance *Guidance) error {
	return WriteStateWithGuidanceAt(p, iteration, result, guidance, time.Now())
}

func WriteStateWithGuidanceAt(p Paths, iteration int, result contracts.RoundResult, guidance *Guidance, now time.Time) error {
	runState := RunState{
		Iteration: iteration,
		UpdatedAt: formatTimestamp(now),
		Result:    result,
		Guidance:  guidance,
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

func WriteSession(p Paths, session SessionMeta) error {
	return writeJSONFile(p.SessionFile, session)
}

func LoadSession(p Paths) (SessionMeta, error) {
	data, err := os.ReadFile(p.SessionFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SessionMeta{}, nil
		}
		return SessionMeta{}, err
	}
	var session SessionMeta
	if err := json.Unmarshal(data, &session); err != nil {
		return SessionMeta{}, err
	}
	return session, nil
}

func ClearSession(p Paths) error {
	if err := os.Remove(p.SessionFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func SessionFresh(session SessionMeta, expiry time.Duration, now time.Time) bool {
	if session.ThreadID == "" {
		return false
	}
	if expiry <= 0 {
		return true
	}
	updatedAt, err := time.Parse(timestampLayout, session.UpdatedAt)
	if err != nil {
		return false
	}
	return now.Sub(updatedAt) <= expiry
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
