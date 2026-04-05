package term

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Theme provides semantic color and style formatting.
type Theme interface {
	Primary() string
	Secondary() string
	Accent() string
	Border() string
	Muted() string
	Danger() string
	Bold() string
	Reset() string
	Success(msg string) string
	Error(msg string) string
	ID(id interface{}) string
}

// T is the active theme. Set by Init().
var T Theme = PlainTheme{}

// Init selects the given color theme if stdout is a TTY, or plain theme otherwise.
func Init(colorTheme Theme) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		T = colorTheme
	} else {
		T = PlainTheme{}
	}
}

func rgb(r, g, b int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}
