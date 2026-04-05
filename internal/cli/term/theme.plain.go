package term

import "fmt"

var _ Theme = PlainTheme{}

// PlainTheme returns empty strings — no colors, no formatting.
type PlainTheme struct{}

func (p PlainTheme) Primary() string           { return "" }
func (p PlainTheme) Secondary() string         { return "" }
func (p PlainTheme) Accent() string            { return "" }
func (p PlainTheme) Border() string            { return "" }
func (p PlainTheme) Muted() string             { return "" }
func (p PlainTheme) Danger() string            { return "" }
func (p PlainTheme) Bold() string              { return "" }
func (p PlainTheme) Reset() string             { return "" }
func (p PlainTheme) Success(msg string) string { return "✓ " + msg }
func (p PlainTheme) Error(msg string) string   { return "✗ " + msg }
func (p PlainTheme) ID(id interface{}) string  { return fmt.Sprintf("#%v", id) }
