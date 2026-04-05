package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/puntopost/acho-mcp/internal/cli/term"
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
their type to exist so they come last). Items that already exist (same id /
same type name) are skipped.

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

	rulesImported, rulesSkipped := 0, 0
	for _, r := range export.Rules {
		if err := d.Rules.Upsert(r.ID, r.Title, r.Text, r.Project, r.Date); err != nil {
			rulesSkipped++
			continue
		}
		rulesImported++
	}

	typesImported, typesSkipped := 0, 0
	for _, t := range export.Types {
		if err := d.Types.Create(t.Name, t.Schema, t.Project, t.Date); err != nil {
			typesSkipped++
			continue
		}
		typesImported++
	}

	regsImported, regsSkipped := 0, 0
	for _, r := range export.Registries {
		if err := d.Service.SaveAs(r.ID, r.Title, r.Content, r.Type, r.Project, r.Date); err != nil {
			regsSkipped++
			continue
		}
		regsImported++
	}

	fmt.Printf("%s rules: %d imported, %d skipped | types: %d / %d | registries: %d / %d (from %s)\n",
		term.T.Success("Imported"),
		rulesImported, rulesSkipped,
		typesImported, typesSkipped,
		regsImported, regsSkipped,
		file)
	return nil
}
