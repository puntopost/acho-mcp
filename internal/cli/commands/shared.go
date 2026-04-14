package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/puntopost/acho-mcp/internal/app"
	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence"
	"github.com/puntopost/acho-mcp/internal/persistence/store"
)

// ConfigPath is set by main.go when --config is passed globally.
var ConfigPath string

// deps is an alias so command files keep using *deps unchanged.
type deps = app.Deps

func loadDeps(args []string) (*deps, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	projectName := flagValue(args, "--project")
	if projectName == "" {
		cwd, _ := os.Getwd()
		projectName = detectProject(cwd)
	}

	return app.Build(&cfg, projectName)
}

func flagValue(args []string, name string) string {
	for i, arg := range args {
		if arg == name && i+1 < len(args) {
			return args[i+1]
		}
		if len(arg) > len(name)+1 && arg[:len(name)+1] == name+"=" {
			return arg[len(name)+1:]
		}
	}
	return ""
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

// projectFlag parses --project with three states:
//
//	absent: the flag is not present at all
//	bare:   the flag is present without value ("--project")
//	value:  the flag is present with a value ("--project=NAME" or "--project NAME")
//
// When state=="value" the second return holds the value.
func projectFlag(args []string) (state, value string) {
	for i, a := range args {
		if a == "--project" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return "value", args[i+1]
			}
			return "bare", ""
		}
		if strings.HasPrefix(a, "--project=") {
			v := strings.TrimPrefix(a, "--project=")
			if v == "" {
				return "bare", ""
			}
			return "value", v
		}
	}
	return "absent", ""
}

func positionalArg(args []string, index int) string {
	pos := 0
	for _, arg := range args {
		if len(arg) > 0 && arg[0] == '-' {
			continue
		}
		if pos == index {
			return arg
		}
		pos++
	}
	return ""
}

func colorizeTruncation(content string) string {
	colored := term.T.Primary() + persistence.TruncationIndicator + term.T.Reset()
	return strings.ReplaceAll(content, persistence.TruncationIndicator, colored)
}

func printItemPreview(item store.RegistryItem) {
	content := colorizeTruncation(item.Content)
	projectLabel := item.Project
	if projectLabel == "" {
		projectLabel = "(global)"
	}
	deletedTag := ""
	if item.Deleted {
		d := ""
		if item.DeletedDate != nil {
			d = " at " + item.DeletedDate.Format(cli.DateYMD_HM)
		}
		deletedTag = fmt.Sprintf(" %s[DELETED%s]%s", term.T.Danger(), d, term.T.Reset())
	}
	fmt.Printf("%s (%s%s%s) — %s%s%s%s\n  %s\n  %sproject: %s%s%s | date: %s%s\n\n",
		term.T.ID(item.ID), term.T.Primary(), item.Type, term.T.Reset(),
		term.T.Bold(), item.Title, term.T.Reset(), deletedTag,
		content,
		term.T.Muted(),
		term.T.Secondary(), projectLabel, term.T.Muted(),
		item.Date.Format(cli.DateYMD_HM), term.T.Reset())
}
