package subagents

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ckken/ralphx/internal/assets"
)

type Spec struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
	SandboxMode     string `json:"sandbox_mode"`
	Source          string `json:"source"`
}

type Location struct {
	Scope string `json:"scope"`
	Path  string `json:"path"`
}

type Status struct {
	Spec      Spec       `json:"spec"`
	Installed bool       `json:"installed"`
	Locations []Location `json:"locations,omitempty"`
	Effective *Location  `json:"effective,omitempty"`
}

type Unknown struct {
	Name      string     `json:"name"`
	Locations []Location `json:"locations"`
}

type Discovery struct {
	Workdir    string    `json:"workdir"`
	GlobalDir  string    `json:"global_dir"`
	ProjectDir string    `json:"project_dir"`
	Catalog    []Status  `json:"catalog"`
	Unknown    []Unknown `json:"unknown,omitempty"`
}

type InstallResult struct {
	Scope     string     `json:"scope"`
	TargetDir string     `json:"target_dir"`
	Installed []Location `json:"installed"`
	Missing   []string   `json:"missing,omitempty"`
}

var defaultCatalog = []Spec{
	{
		Name:            "workflow-orchestrator",
		Category:        "meta-orchestration",
		Description:     "Use when the parent agent needs an explicit Codex subagent workflow for a complex task with multiple stages.",
		Model:           "gpt-5.4",
		ReasoningEffort: "high",
		SandboxMode:     "read-only",
		Source:          "awesome-codex-subagents/categories/09-meta-orchestration/workflow-orchestrator.toml",
	},
	{
		Name:            "task-distributor",
		Category:        "meta-orchestration",
		Description:     "Use when a broad task needs to be broken into concrete sub-tasks with clear boundaries for multiple agents or contributors.",
		Model:           "gpt-5.4",
		ReasoningEffort: "high",
		SandboxMode:     "read-only",
		Source:          "awesome-codex-subagents/categories/09-meta-orchestration/task-distributor.toml",
	},
	{
		Name:            "context-manager",
		Category:        "meta-orchestration",
		Description:     "Use when a task needs a compact project context summary that other subagents can rely on before deeper work begins.",
		Model:           "gpt-5.3-codex-spark",
		ReasoningEffort: "medium",
		SandboxMode:     "read-only",
		Source:          "awesome-codex-subagents/categories/09-meta-orchestration/context-manager.toml",
	},
	{
		Name:            "reviewer",
		Category:        "quality-security",
		Description:     "Use when a task needs PR-style review focused on correctness, security, behavior regressions, and missing tests.",
		Model:           "gpt-5.4",
		ReasoningEffort: "high",
		SandboxMode:     "read-only",
		Source:          "awesome-codex-subagents/categories/04-quality-security/reviewer.toml",
	},
	{
		Name:            "debugger",
		Category:        "quality-security",
		Description:     "Use when a task needs deep bug isolation across code paths, stack traces, runtime behavior, or failing tests.",
		Model:           "gpt-5.4",
		ReasoningEffort: "high",
		SandboxMode:     "read-only",
		Source:          "awesome-codex-subagents/categories/04-quality-security/debugger.toml",
	},
	{
		Name:            "test-automator",
		Category:        "quality-security",
		Description:     "Use when a task needs implementation of automated tests, test harness improvements, or targeted regression coverage.",
		Model:           "gpt-5.3-codex-spark",
		ReasoningEffort: "medium",
		SandboxMode:     "workspace-write",
		Source:          "awesome-codex-subagents/categories/04-quality-security/test-automator.toml",
	},
	{
		Name:            "backend-developer",
		Category:        "core-development",
		Description:     "Use when a task needs scoped backend implementation or backend bug fixes after the owning path is known.",
		Model:           "gpt-5.4",
		ReasoningEffort: "high",
		SandboxMode:     "workspace-write",
		Source:          "awesome-codex-subagents/categories/01-core-development/backend-developer.toml",
	},
	{
		Name:            "fullstack-developer",
		Category:        "core-development",
		Description:     "Use when one bounded feature or bug spans frontend and backend and a single worker should own the entire path.",
		Model:           "gpt-5.4",
		ReasoningEffort: "high",
		SandboxMode:     "workspace-write",
		Source:          "awesome-codex-subagents/categories/01-core-development/fullstack-developer.toml",
	},
}

func Catalog() []Spec {
	out := make([]Spec, len(defaultCatalog))
	copy(out, defaultCatalog)
	return out
}

func DefaultNames() []string {
	out := make([]string, 0, len(defaultCatalog))
	for _, spec := range defaultCatalog {
		out = append(out, spec.Name)
	}
	return out
}

func Install(workdir string, project bool, names []string) (InstallResult, error) {
	if strings.TrimSpace(workdir) == "" {
		wd, err := os.Getwd()
		if err != nil {
			return InstallResult{}, err
		}
		workdir = wd
	}
	if len(names) == 0 {
		names = DefaultNames()
	}

	scope := "global"
	targetDir := globalAgentsDir()
	if project {
		scope = "project"
		targetDir = projectAgentsDir(workdir)
	}

	unknown := missingSpecs(names)
	if len(unknown) > 0 {
		return InstallResult{
			Scope:     scope,
			TargetDir: targetDir,
			Missing:   unknown,
		}, fmt.Errorf("unknown subagents: %s", strings.Join(unknown, ", "))
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return InstallResult{}, err
	}

	result := InstallResult{
		Scope:     scope,
		TargetDir: targetDir,
	}
	for _, name := range names {
		data, ok := assets.SubagentBytes(name)
		if !ok {
			result.Missing = append(result.Missing, name)
			continue
		}
		path := filepath.Join(targetDir, name+".toml")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return result, err
		}
		result.Installed = append(result.Installed, Location{Scope: scope, Path: path})
	}
	if len(result.Missing) > 0 {
		return result, fmt.Errorf("missing embedded subagents: %s", strings.Join(result.Missing, ", "))
	}
	return result, nil
}

func Discover(workdir string) (Discovery, error) {
	if strings.TrimSpace(workdir) == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Discovery{}, err
		}
		workdir = wd
	}
	out := Discovery{
		Workdir:    workdir,
		GlobalDir:  globalAgentsDir(),
		ProjectDir: projectAgentsDir(workdir),
	}

	for _, spec := range defaultCatalog {
		status := Status{Spec: spec}
		for _, loc := range discoverLocations(spec.Name, out.ProjectDir, out.GlobalDir) {
			status.Locations = append(status.Locations, loc)
		}
		if len(status.Locations) > 0 {
			status.Installed = true
			effective := status.Locations[0]
			status.Effective = &effective
		}
		out.Catalog = append(out.Catalog, status)
	}

	extras := discoverExtras(out.ProjectDir, out.GlobalDir, catalogNamesSet())
	sort.Slice(extras, func(i, j int) bool {
		if extras[i].Name == extras[j].Name {
			return extras[i].Locations[0].Path < extras[j].Locations[0].Path
		}
		return extras[i].Name < extras[j].Name
	})
	out.Unknown = extras
	return out, nil
}

func discoverLocations(name, projectDir, globalDir string) []Location {
	locations := make([]Location, 0, 2)
	projectPath := filepath.Join(projectDir, name+".toml")
	if fileExists(projectPath) {
		locations = append(locations, Location{Scope: "project", Path: projectPath})
	}
	globalPath := filepath.Join(globalDir, name+".toml")
	if fileExists(globalPath) {
		locations = append(locations, Location{Scope: "global", Path: globalPath})
	}
	return locations
}

func discoverExtras(projectDir, globalDir string, known map[string]struct{}) []Unknown {
	index := map[string][]Location{}
	for _, item := range scanAgentsDir("project", projectDir) {
		if _, ok := known[item.Name]; ok {
			continue
		}
		index[item.Name] = append(index[item.Name], item.Location)
	}
	for _, item := range scanAgentsDir("global", globalDir) {
		if _, ok := known[item.Name]; ok {
			continue
		}
		index[item.Name] = append(index[item.Name], item.Location)
	}
	out := make([]Unknown, 0, len(index))
	for name, locations := range index {
		out = append(out, Unknown{Name: name, Locations: locations})
	}
	return out
}

type agentFile struct {
	Name     string
	Location Location
}

func scanAgentsDir(scope, dir string) []agentFile {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]agentFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".toml" {
			continue
		}
		out = append(out, agentFile{
			Name: strings.TrimSuffix(name, ".toml"),
			Location: Location{
				Scope: scope,
				Path:  filepath.Join(dir, name),
			},
		})
	}
	return out
}

func catalogNamesSet() map[string]struct{} {
	out := make(map[string]struct{}, len(defaultCatalog))
	for _, spec := range defaultCatalog {
		out[spec.Name] = struct{}{}
	}
	return out
}

func missingSpecs(names []string) []string {
	known := catalogNamesSet()
	missing := make([]string, 0)
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := known[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

func globalAgentsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".codex", "agents")
	}
	codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if codexHome == "" {
		codexHome = filepath.Join(home, ".codex")
	}
	return filepath.Join(codexHome, "agents")
}

func projectAgentsDir(root string) string {
	return filepath.Join(root, ".codex", "agents")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
