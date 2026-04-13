package state

import (
	"path/filepath"
)

const (
	DirName         = ".ralphx"
	LogsDirName     = "logs"
	StateFileName   = "state.json"
	LastOutputName  = "last-output.txt"
	LastResultName  = "last-result.json"
	SummaryFileName = "summary.txt"
	StatsFileName   = "stats.json"
	SessionFileName = "session.json"
	defaultFileMode = 0o644
	defaultDirMode  = 0o755
	timestampLayout = "2006-01-02 15:04:05"
)

func DerivePaths(workdir, stateDir string) Paths {
	root := stateDir
	if root == "" {
		root = filepath.Join(workdir, DirName)
	}

	return Paths{
		Root:           root,
		LogDir:         filepath.Join(root, LogsDirName),
		StateFile:      filepath.Join(root, StateFileName),
		LastOutputFile: filepath.Join(root, LastOutputName),
		LastJSONFile:   filepath.Join(root, LastResultName),
		SummaryFile:    filepath.Join(root, SummaryFileName),
		StatsFile:      filepath.Join(root, StatsFileName),
		SessionFile:    filepath.Join(root, SessionFileName),
	}
}
