package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence"
)

func init() {
	Register(&importCmd{})
}

var _ Command = (*importCmd)(nil)

type importCmd struct{}

func (c *importCmd) Match(name string) bool { return name == "import" }
func (c *importCmd) Usage() string          { return "acho import [file]" }
func (c *importCmd) Description() string    { return "Import rules, types and registries from JSON" }
func (c *importCmd) Order() int             { return 39 }
func (c *importCmd) Help() string {
	return `acho import — Import rules, types and registries from a JSON export

Usage:
  acho import [file]

Arguments:
  file                   JSON file from acho export (default: acho-export.json)

Imports in order: rules first, then types, then registries (registries need
their type to exist so they come last). Existing items (same id / same type
name) are skipped. Any other error is reported and makes the command fail.

The file must have been exported with the same acho version.

Examples:
  acho import acho-export.json
  acho import backup.json
`
}

func (c *importCmd) Run(args []string) error {
	file := positionalArg(args, 0)
	if file == "" {
		if _, err := os.Stat("acho-export.json"); err != nil {
			return fmt.Errorf("usage: acho import <file>")
		}
		file = "acho-export.json"
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var export exportData
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if export.Version != Version {
		return fmt.Errorf("version mismatch: file is %s, acho is %s", export.Version, Version)
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	rulesImported, rulesSkipped, rulesFailed := 0, 0, 0
	var failures []string
	for _, r := range export.Rules {
		exists, err := importRuleExists(d, r.ID)
		if err != nil {
			rulesFailed++
			failures = append(failures, fmt.Sprintf("rule %s: check existing: %v", r.ID, err))
			continue
		}
		if exists {
			rulesSkipped++
			continue
		}
		if err := d.Rules.Upsert(r.ID, r.Title, r.Text, r.Project, r.Date); err != nil {
			rulesFailed++
			failures = append(failures, fmt.Sprintf("rule %s: %v", r.ID, err))
			continue
		}
		rulesImported++
	}

	typesImported, typesSkipped, typesFailed := 0, 0, 0
	for _, t := range export.Types {
		exists, err := importTypeExists(d, t.Name)
		if err != nil {
			typesFailed++
			failures = append(failures, fmt.Sprintf("type %s: check existing: %v", t.Name, err))
			continue
		}
		if exists {
			typesSkipped++
			continue
		}
		if err := d.Types.Create(t.Name, t.Description, t.Schema, t.Project, t.Date); err != nil {
			typesFailed++
			failures = append(failures, fmt.Sprintf("type %s: %v", t.Name, err))
			continue
		}
		typesImported++
	}

	regsImported, regsSkipped, regsFailed := 0, 0, 0
	for _, r := range export.Registries {
		exists, err := importRegistryExists(d, r.ID)
		if err != nil {
			regsFailed++
			failures = append(failures, fmt.Sprintf("registry %s: check existing: %v", r.ID, err))
			continue
		}
		if exists {
			regsSkipped++
			continue
		}
		if err := d.Service.SaveAs(r.ID, r.Title, r.Content, r.Type, r.Project, r.Date); err != nil {
			regsFailed++
			failures = append(failures, fmt.Sprintf("registry %s: %v", r.ID, err))
			continue
		}
		regsImported++
	}

	fmt.Printf("%s rules: %d imported, %d skipped, %d failed | types: %d imported, %d skipped, %d failed | registries: %d imported, %d skipped, %d failed (from %s)\n",
		term.T.Success("Imported"),
		rulesImported, rulesSkipped, rulesFailed,
		typesImported, typesSkipped, typesFailed,
		regsImported, regsSkipped, regsFailed,
		file)
	if len(failures) > 0 {
		return fmt.Errorf("import failed:\n%s", strings.Join(failures, "\n"))
	}
	return nil
}

func importRuleExists(d *deps, id string) (bool, error) {
	_, err := d.Rules.GetAny(id)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, persistence.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func importTypeExists(d *deps, name string) (bool, error) {
	_, err := d.Types.GetAny(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, persistence.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func importRegistryExists(d *deps, id string) (bool, error) {
	_, err := d.Service.GetAny(id)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, persistence.ErrNotFound) {
		return false, nil
	}
	return false, err
}
