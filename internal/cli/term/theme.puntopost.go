package term

import "fmt"

var _ Theme = PuntoPostTheme{}

// PuntoPostTheme renders PuntoPost brand colors as ANSI escape codes.
type PuntoPostTheme struct{}

func (t PuntoPostTheme) Primary() string   { return rgb(19, 165, 144) }  // green
func (t PuntoPostTheme) Secondary() string { return rgb(255, 225, 53) }  // yellow
func (t PuntoPostTheme) Accent() string    { return rgb(208, 237, 233) } // light green
func (t PuntoPostTheme) Border() string    { return rgb(12, 110, 96) }   // dark green
func (t PuntoPostTheme) Muted() string     { return rgb(100, 100, 100) } // gray
func (t PuntoPostTheme) Danger() string    { return rgb(200, 40, 40) }   // red
func (t PuntoPostTheme) Bold() string      { return "\033[1m" }
func (t PuntoPostTheme) Reset() string     { return "\033[0m" }

func (t PuntoPostTheme) Success(msg string) string {
	return fmt.Sprintf("%s%s✓ %s%s", t.Bold(), t.Primary(), msg, t.Reset())
}

func (t PuntoPostTheme) Error(msg string) string {
	return fmt.Sprintf("%s%s✗ %s%s", t.Bold(), t.Danger(), msg, t.Reset())
}

func (t PuntoPostTheme) ID(id interface{}) string {
	return fmt.Sprintf("%s%s#%v%s", t.Bold(), t.Secondary(), id, t.Reset())
}
