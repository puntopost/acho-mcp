package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	goterm "golang.org/x/term"
)

// ErrCancelled is returned when the user presses Escape or Ctrl+C during a prompt.
var ErrCancelled = errors.New("cancelled")

// PromptLine reads a single line from stdin with raw terminal mode,
// allowing detection of Escape (cancel) and proper character editing.
// Falls back to standard line reading when stdin is not a terminal.
func PromptLine() (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := goterm.MakeRaw(fd)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		return strings.TrimSpace(line), nil
	}
	defer goterm.Restore(fd, oldState)

	var buf []byte
	var b [1]byte
	for {
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			fmt.Print("\r\n")
			return "", err
		}
		c := b[0]
		switch {
		case c == 0x1b:
			goterm.Restore(fd, oldState)
			fmt.Print("\r\n")
			return "", ErrCancelled
		case c == 0x03:
			fmt.Print("\r\n")
			return "", ErrCancelled
		case c == 0x0d || c == 0x0a:
			fmt.Print("\r\n")
			return string(buf), nil
		case c == 0x7f || c == 0x08:
			if len(buf) > 0 {
				_, size := utf8.DecodeLastRune(buf)
				buf = buf[:len(buf)-size]
				fmt.Print("\b \b")
			}
		case c >= 0xC0:
			width := 2
			if c >= 0xF0 {
				width = 4
			} else if c >= 0xE0 {
				width = 3
			}
			mb := []byte{c}
			for i := 1; i < width; i++ {
				if _, err := os.Stdin.Read(b[:]); err != nil {
					break
				}
				mb = append(mb, b[0])
			}
			buf = append(buf, mb...)
			fmt.Print(string(mb))
		case c >= 0x20 && c < 0x7f:
			buf = append(buf, c)
			fmt.Print(string(c))
		}
	}
}
