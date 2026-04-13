package task

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var openChecklistItemPattern = regexp.MustCompile(`(?m)^[\t ]*[-*][\t ]+\[ \]`)

type Document struct {
	Path    string
	Content string
}

type Checklist struct {
	Path           string
	Content        string
	AutoDiscovered bool
	OpenItems      int
}

type StateTexts struct {
	SummaryPath string
	Summary     string
	StatePath   string
	State       string
}

type Bundle struct {
	Task      Document
	Checklist Checklist
	State     StateTexts
}

type LoadOptions struct {
	ChecklistPath string
	SummaryPath   string
	StatePath     string
}

func Load(taskPath string, opts LoadOptions) (Bundle, error) {
	taskDoc, err := ReadTask(taskPath)
	if err != nil {
		return Bundle{}, err
	}

	checklist, err := LoadChecklist(taskDoc.Path, opts.ChecklistPath)
	if err != nil {
		return Bundle{}, err
	}

	state, err := LoadStateTexts(opts.SummaryPath, opts.StatePath)
	if err != nil {
		return Bundle{}, err
	}

	return Bundle{
		Task:      taskDoc,
		Checklist: checklist,
		State:     state,
	}, nil
}

func ReadTask(path string) (Document, error) {
	if strings.TrimSpace(path) == "" {
		return Document{}, errors.New("task file path is required")
	}
	content, resolved, err := readExistingFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("read task: %w", err)
	}
	return Document{Path: resolved, Content: content}, nil
}

func LoadChecklist(taskPath, explicitPath string) (Checklist, error) {
	resolved, auto, err := ResolveChecklistPath(taskPath, explicitPath)
	if err != nil {
		return Checklist{}, err
	}
	if resolved == "" {
		return Checklist{}, nil
	}

	content, fullPath, err := readExistingFile(resolved)
	if err != nil {
		return Checklist{}, fmt.Errorf("read checklist: %w", err)
	}

	return Checklist{
		Path:           fullPath,
		Content:        content,
		AutoDiscovered: auto,
		OpenItems:      CountOpenChecklistItems(content),
	}, nil
}

func ResolveChecklistPath(taskPath, explicitPath string) (path string, autoDiscovered bool, err error) {
	if strings.TrimSpace(explicitPath) != "" {
		fullPath, resolveErr := filepath.Abs(explicitPath)
		if resolveErr != nil {
			return "", false, fmt.Errorf("resolve checklist path: %w", resolveErr)
		}
		if _, statErr := os.Stat(fullPath); statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) {
				return "", false, fmt.Errorf("checklist file not found: %s", explicitPath)
			}
			return "", false, fmt.Errorf("checklist file not found: %w", statErr)
		}
		return fullPath, false, nil
	}

	taskAbs, absErr := filepath.Abs(taskPath)
	if absErr != nil {
		return "", false, fmt.Errorf("resolve task path: %w", absErr)
	}

	autoPath := autoChecklistPath(taskAbs)
	if autoPath == "" {
		return "", false, nil
	}
	if _, statErr := os.Stat(autoPath); statErr == nil {
		return autoPath, true, nil
	} else if errors.Is(statErr, os.ErrNotExist) {
		return "", false, nil
	} else {
		return "", false, fmt.Errorf("stat checklist file: %w", statErr)
	}
}

func CountOpenChecklistItems(content string) int {
	return len(openChecklistItemPattern.FindAllStringIndex(content, -1))
}

func LoadStateTexts(summaryPath, statePath string) (StateTexts, error) {
	summary, resolvedSummary, err := ReadOptionalTextFile(summaryPath)
	if err != nil {
		return StateTexts{}, fmt.Errorf("read summary: %w", err)
	}
	state, resolvedState, err := ReadOptionalTextFile(statePath)
	if err != nil {
		return StateTexts{}, fmt.Errorf("read state: %w", err)
	}

	return StateTexts{
		SummaryPath: resolvedSummary,
		Summary:     summary,
		StatePath:   resolvedState,
		State:       state,
	}, nil
}

func ReadOptionalTextFile(path string) (content string, resolvedPath string, err error) {
	if strings.TrimSpace(path) == "" {
		return "", "", nil
	}

	fullPath, absErr := filepath.Abs(path)
	if absErr != nil {
		return "", "", fmt.Errorf("resolve path: %w", absErr)
	}

	data, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return "", fullPath, nil
		}
		return "", fullPath, readErr
	}
	return string(data), fullPath, nil
}

func autoChecklistPath(taskPath string) string {
	ext := filepath.Ext(taskPath)
	if ext != ".md" {
		return ""
	}
	return strings.TrimSuffix(taskPath, ext) + ".checklist" + ext
}

func readExistingFile(path string) (content string, resolvedPath string, err error) {
	fullPath, absErr := filepath.Abs(path)
	if absErr != nil {
		return "", "", fmt.Errorf("resolve path: %w", absErr)
	}
	data, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		if errors.Is(readErr, os.ErrNotExist) {
			return "", fullPath, fmt.Errorf("file not found: %s", path)
		}
		return "", fullPath, readErr
	}
	return string(data), fullPath, nil
}
