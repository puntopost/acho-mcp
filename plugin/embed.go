package plugin

import (
	"embed"
	"io/fs"
)

//go:embed all:claude-code all:opencode
var embedded embed.FS

var ClaudeFS fs.FS = mustSub(embedded, "claude-code")
var OpenCodeFS fs.FS = mustSub(embedded, "opencode")

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic("plugin embed: " + err.Error())
	}
	return sub
}
