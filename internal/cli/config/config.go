package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

type Config struct {
	DBPath                 string   `json:"db_path"`
	DBOptions              string   `json:"db_options"`
	DefaultPaginationLimit int      `json:"default_pagination_limit"`
	SnippetLength          int      `json:"snippet_length"`
	EnabledProjects        []string `json:"enabled_projects,omitempty"`
	Version                string   `json:"-"`
}

func DefaultDir() string {
	if dir := os.Getenv("ACHO_PATH"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".acho")
}

func DefaultPath() string {
	return filepath.Join(DefaultDir(), "config.json")
}

func configLockPath() string {
	return filepath.Join(DefaultDir(), "config.lock")
}

// WithConfigLock acquires an exclusive advisory lock on the config lockfile,
// runs fn, and releases. Serializes read-modify-write of config.json across
// concurrent acho processes so two enable/disable operations cannot clobber
// each other.
func WithConfigLock(fn func() error) error {
	p := configLockPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("config lock: %w", err)
	}
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("config lock: %w", err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("config lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()
	return fn()
}

func Load() (Config, error) {
	return LoadFrom(DefaultPath())
}

func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}
	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	cfg.EnabledProjects = NormalizeProjects(cfg.EnabledProjects)
	return cfg, nil
}

func Default() Config {
	return Config{
		DBPath:                 filepath.Join(DefaultDir(), "acho.db"),
		DBOptions:              "_journal_mode=wal&_busy_timeout=5000&_synchronous=normal&_foreign_keys=on",
		DefaultPaginationLimit: 10,
		SnippetLength:          1000,
	}
}

func (c Config) IsProjectEnabled(project string) bool {
	if project == "" {
		return false
	}
	for _, p := range c.EnabledProjects {
		if p == project {
			return true
		}
	}
	return false
}

func NormalizeProjects(projects []string) []string {
	if len(projects) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(projects))
	out := make([]string, 0, len(projects))
	for _, project := range projects {
		project = strings.TrimSpace(project)
		if project == "" {
			continue
		}
		if _, ok := seen[project]; ok {
			continue
		}
		seen[project] = struct{}{}
		out = append(out, project)
	}
	if len(out) == 0 {
		return nil
	}
	slices.Sort(out)
	return out
}
