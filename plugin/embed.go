package plugin

import (
	"embed"
	"io/fs"
)

//go:embed all:claude-code all:opencode
var embedded embed.FS

// ClaudeFS exposes the embedded Claude Code plugin files for extraction during
// agent setup. Sub strips the "claude-code" prefix so the root contains
// .claude-plugin/, hooks/, etc.
var ClaudeFS fs.FS = mustSub(embedded, "claude-code")

// OpenCodeFS exposes the embedded OpenCode plugin files for extraction during
// agent setup. Sub strips the "opencode" prefix so the root contains plugins/
// and instructions/.
var OpenCodeFS fs.FS = mustSub(embedded, "opencode")

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic("plugin embed: " + err.Error())
	}
	return sub
}
