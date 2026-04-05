package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/internal/persistence/rtype"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
	"github.com/puntopost/acho-mcp/internal/persistence/store"
)

func init() {
	Register(&export{})
}

var _ Command = (*export)(nil)

type export struct{}

func (c *export) Match(name string) bool { return name == "export" }
func (c *export) Usage() string          { return "acho export [file]" }
func (c *export) Description() string    { return "Export rules, types and registries to JSON" }
func (c *export) Order() int             { return 38 }
func (c *export) Help() string {
	return `acho export — Export rules, types and registries to JSON

Usage:
  acho export [file]

Arguments:
  file                   Output file path (default: acho-export.json)

Exports every rule, type and registry to a single JSON file with the acho
version and timestamp. Import with acho import on the same version.

Examples:
  acho export
  acho export backup.json
`
}

type exportData struct {
	Version    string           `json:"version"`
	ExportedAt string           `json:"exported_at"`
	Rules      []rule.Rule      `json:"rules"`
	Types      []rtype.RType    `json:"types"`
	Registries []store.Registry `json:"registries"`
}

func (c *export) Run(args []string) error {
	file := positionalArg(args, 0)
	if file == "" {
		file = "acho-export.json"
	}

	d, err := loadDeps(args)
	if err != nil {
		return err
	}
	defer d.Close()

	types, err := d.Types.List("", false, rtype.ListQuery{})
	if err != nil {
		return fmt.Errorf("export types: %w", err)
	}
	rules, err := d.Rules.List("", false, rule.ListQuery{})
	if err != nil {
		return fmt.Errorf("export rules: %w", err)
	}
	registries, err := d.Service.ExportAll()
	if err != nil {
		return fmt.Errorf("export registries: %w", err)
	}

	data := exportData{
		Version:    Version,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Types:      types,
		Rules:      rules,
		Registries: registries,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.WriteFile(file, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("%s %s%d%s rules, %s%d%s types, %s%d%s registries to %s%s%s\n",
		term.T.Success("Exported"),
		term.T.Bold(), len(rules), term.T.Reset(),
		term.T.Bold(), len(types), term.T.Reset(),
		term.T.Bold(), len(registries), term.T.Reset(),
		term.T.Bold(), file, term.T.Reset())
	return nil
}
