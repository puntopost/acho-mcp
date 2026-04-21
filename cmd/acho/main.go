package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/puntopost/acho-mcp/internal/cli"
	"github.com/puntopost/acho-mcp/internal/cli/agent"
	"github.com/puntopost/acho-mcp/internal/cli/commands"
	"github.com/puntopost/acho-mcp/internal/cli/term"
	"github.com/puntopost/acho-mcp/plugin"
)

func main() {
	term.Init(term.PuntoPostTheme{})
	cli.InitLogger()
	commands.Version = "0.5.0"

	// Make embedded plugin files available to agent setup
	agent.PluginFS = plugin.ClaudeFS
	agent.OpenCodePluginFS = plugin.OpenCodeFS
	agent.PluginVersion = commands.Version
	agent.OpenCodePluginVersion = commands.Version

	args := os.Args[1:]
	slog.Info("cli invocation", "args", truncateArgs(args), "arg_count", len(args))

	// Extract global --config flag before dispatching
	args = extractConfigFlag(args)

	if len(args) < 1 {
		slog.Info("cli invocation completed", "outcome", "usage")
		commands.PrintUsage()
		return
	}

	// Try two-word subcommand first (e.g. "model download"), then single word
	for _, depth := range []int{2, 1} {
		if len(args) < depth {
			continue
		}
		name := strings.Join(args[:depth], " ")
		for _, cmd := range commands.All() {
			if cmd.Match(name) {
				rest := args[depth:]
				start := time.Now()
				slog.Info("cli command started", "command", name, "args", truncateArgs(rest), "arg_count", len(rest))
				if len(rest) > 0 && (rest[0] == "--help" || rest[0] == "-h") {
					slog.Info("cli command completed", "command", name, "outcome", "help", "duration_ms", time.Since(start).Milliseconds())
					fmt.Print(term.FormatHelp(cmd.Help()))
					return
				}
				if err := cmd.Run(rest); err != nil {
					if errors.Is(err, cli.ErrCancelled) {
						slog.Info("cli command cancelled", "command", name, "duration_ms", time.Since(start).Milliseconds())
						fmt.Println("Cancelled.")
						return
					}
					slog.Error("cli command failed", "command", name, "args", truncateArgs(rest), "duration_ms", time.Since(start).Milliseconds(), "error", err)
					fmt.Fprintf(os.Stderr, "%s\n", term.T.Error(err.Error()))
					os.Exit(1)
				}
				slog.Info("cli command completed", "command", name, "duration_ms", time.Since(start).Milliseconds())
				return
			}
		}
	}

	slog.Error("cli command failed", "command", strings.Join(args, " "), "error", "unknown command")
	fmt.Fprintf(os.Stderr, "%s\n", term.T.Error("unknown command: "+strings.Join(args, " ")))
	os.Exit(1)
}

func truncateArgs(args []string) []string {
	truncated := make([]string, 0, len(args))
	for _, arg := range args {
		if len(arg) > 200 {
			truncated = append(truncated, arg[:197]+"...")
			continue
		}
		truncated = append(truncated, arg)
	}
	return truncated
}

func extractConfigFlag(args []string) []string {
	var remaining []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--config" && i+1 < len(args) {
			commands.ConfigPath = args[i+1]
			i++ // skip value
		} else if len(args[i]) > 9 && args[i][:9] == "--config=" {
			commands.ConfigPath = args[i][9:]
		} else {
			remaining = append(remaining, args[i])
		}
	}
	return remaining
}
