package hooks

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/ckken/ralphx/internal/contracts"
	"github.com/ckken/ralphx/internal/state"
	"github.com/ckken/ralphx/internal/task"
)

var ErrNoTaskContext = errors.New("no task context available")

func LoadStopGuardInput(taskPath, checklistPath, summaryPath, statePath, lastResultPath string, testsRequired, testsPassedNow bool) (GuardInput, error) {
	taskPath, checklistPath = inferTaskPaths(taskPath, checklistPath, statePath)
	if strings.TrimSpace(taskPath) == "" {
		return GuardInput{}, ErrNoTaskContext
	}
	bundle, err := task.Load(taskPath, task.LoadOptions{
		ChecklistPath: checklistPath,
		SummaryPath:   summaryPath,
		StatePath:     statePath,
	})
	if err != nil {
		return GuardInput{}, err
	}

	result, err := loadResult(statePath, lastResultPath)
	if err != nil {
		return GuardInput{}, err
	}

	return GuardInput{
		Event:          EventStop,
		Result:         result,
		ChecklistOpen:  bundle.Checklist.OpenItems,
		TestsRequired:  testsRequired,
		TestsPassedNow: testsPassedNow,
	}, nil
}

func inferTaskPaths(taskPath, checklistPath, statePath string) (string, string) {
	if strings.TrimSpace(taskPath) != "" {
		return taskPath, checklistPath
	}
	if statePath == "" {
		return taskPath, checklistPath
	}
	if data, err := os.ReadFile(statePath); err == nil {
		var runState state.RunState
		if err := json.Unmarshal(data, &runState); err == nil {
			if strings.TrimSpace(taskPath) == "" {
				taskPath = runState.TaskFile
			}
			if strings.TrimSpace(checklistPath) == "" {
				checklistPath = runState.ChecklistFile
			}
		}
	}
	return taskPath, checklistPath
}

func loadResult(statePath, lastResultPath string) (contracts.RoundResult, error) {
	if lastResultPath != "" {
		if data, err := os.ReadFile(lastResultPath); err == nil {
			var result contracts.RoundResult
			if err := json.Unmarshal(data, &result); err == nil {
				return result, nil
			}
		}
	}

	if statePath != "" {
		if data, err := os.ReadFile(statePath); err == nil {
			var runState state.RunState
			if err := json.Unmarshal(data, &runState); err == nil {
				return runState.Result, nil
			}
		}
	}

	return contracts.RoundResult{}, nil
}
