package commands

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/puntopost/acho-mcp/internal/cli/assets"
	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

func init() {
	Register(&juan{})
}

var _ Command = (*juan)(nil)

type juan struct{}

const juanCaption = "El espiritu de Juan ha sido imbuido dentro de Acho"
const juanRuleTitle = "Juan, your trusted Murcian companion"
const juanRuleText = "You are a brutally incisive senior developer from Murcia. You use dark humor, sarcasm, and sharp criticism to expose mistakes, inconsistencies, and sloppy work with cruel but useful precision. You always help, but never sugarcoat bad ideas, cheap hacks, or wasted effort. Keep the tone dry, sharp, understandable, and solution-oriented. You must speak only in Murcian Spanish (Castilian from Murcia), never in English or any other language."
const juanBorder = "\x1b[38;5;245m"
const juanCaptionColor = "\x1b[38;5;250m"
const juanReset = "\x1b[0m"

func (c *juan) Match(name string) bool { return name == "juan" }
func (c *juan) Usage() string          { return "acho juan" }
func (c *juan) Description() string    { return "Print Juan ASCII art" }
func (c *juan) Order() int             { return 999 }
func (c *juan) Help() string {
	return "acho juan — Print the ANSI-colored Juan ASCII art\n\nUsage:\n  acho juan\n"
}

func (c *juan) Run(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}

	art, err := readJuanArt()
	if err != nil {
		return err
	}
	if err := ensureJuanRule(); err != nil {
		return err
	}

	fmt.Print(string(renderJuanArt(art)))
	return nil
}

func ensureJuanRule() error {
	d, err := loadDeps(nil)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Rules.Upsert(rule.JuanRuleID, juanRuleTitle, juanRuleText, "", time.Unix(0, 0).UTC()); err != nil {
		return err
	}
	return nil
}

func readJuanArt() ([]byte, error) {
	art := assets.JuanANS()
	if len(art) == 0 {
		return nil, fmt.Errorf("embedded juan art is empty")
	}
	return art, nil
}

func stripANSIBackgrounds(src []byte) []byte {
	var out bytes.Buffer
	for i := 0; i < len(src); i++ {
		if src[i] != 0x1b || i+1 >= len(src) || src[i+1] != '[' {
			out.WriteByte(src[i])
			continue
		}

		j := i + 2
		for j < len(src) && src[j] != 'm' {
			j++
		}
		if j >= len(src) {
			out.Write(src[i:])
			break
		}

		params := strings.Split(string(src[i+2:j]), ";")
		filtered := filterANSIBackgroundParams(params)
		if len(filtered) > 0 {
			out.WriteString("\x1b[")
			out.WriteString(strings.Join(filtered, ";"))
			out.WriteByte('m')
		}
		i = j
	}
	return out.Bytes()
}

func renderJuanArt(src []byte) []byte {
	art := stripANSIBackgrounds(src)
	lines := visibleJuanLines(art)
	width := maxANSITextWidth(lines)
	if width == 0 {
		return art
	}

	var out bytes.Buffer
	out.WriteString(renderJuanBorder(width, '-'))
	out.WriteByte('\n')
	for _, line := range lines {
		out.WriteString(renderJuanBoxLine(line, width, juanReset))
		out.WriteByte('\n')
	}
	out.WriteString(renderJuanBorder(width, '-'))
	out.WriteByte('\n')
	out.WriteString(renderJuanCaption(width))
	out.WriteByte('\n')
	out.WriteString(renderJuanBorder(width, '-'))
	out.WriteByte('\n')
	return out.Bytes()
}

func renderJuanCaption(width int) string {
	text := juanCaption
	textWidth := len(text)
	if textWidth >= width {
		return renderJuanBoxString(juanCaptionColor+text+juanReset, textWidth, juanCaptionColor)
	}

	left := (width - textWidth) / 2
	right := width - textWidth - left
	line := juanCaptionColor + strings.Repeat(" ", left) + text + strings.Repeat(" ", right) + juanReset
	return renderJuanBoxString(line, width, juanCaptionColor)
}

func visibleJuanLines(src []byte) [][]byte {
	lines := bytes.Split(src, []byte{'\n'})
	visible := make([][]byte, 0, len(lines))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) > 0 {
			visible = append(visible, line)
		}
	}
	return visible
}

func maxANSITextWidth(lines [][]byte) int {
	width := 0
	for _, line := range lines {
		lineWidth := ansiTextWidth(line)
		if lineWidth > width {
			width = lineWidth
		}
	}
	return width
}

func renderJuanBorder(width int, fill rune) string {
	return juanBorder + "+" + strings.Repeat(string(fill), width+2) + "+" + juanReset
}

func renderJuanBoxLine(content []byte, width int, color string) string {
	padding := width - ansiTextWidth(content)
	if padding < 0 {
		padding = 0
	}
	return juanBorder + "| " + juanReset + string(content) + color + strings.Repeat(" ", padding) + juanReset + juanBorder + " |" + juanReset
}

func renderJuanBoxString(content string, width int, color string) string {
	padding := width - ansiTextWidth([]byte(content))
	if padding < 0 {
		padding = 0
	}
	return juanBorder + "| " + juanReset + content + color + strings.Repeat(" ", padding) + juanReset + juanBorder + " |" + juanReset
}

func ansiTextWidth(line []byte) int {
	width := 0
	for i := 0; i < len(line); i++ {
		if line[i] == 0x1b && i+1 < len(line) && line[i+1] == '[' {
			j := i + 2
			for j < len(line) && line[j] != 'm' {
				j++
			}
			if j >= len(line) {
				break
			}
			i = j
			continue
		}
		width++
	}
	return width
}

func filterANSIBackgroundParams(params []string) []string {
	filtered := make([]string, 0, len(params))
	for i := 0; i < len(params); i++ {
		param := params[i]
		code, err := strconv.Atoi(param)
		if err != nil {
			filtered = append(filtered, param)
			continue
		}

		if isANSIBackgroundCode(code) {
			continue
		}
		if code == 48 {
			if i+1 < len(params) {
				next := params[i+1]
				if next == "5" {
					i += 2
					continue
				}
				if next == "2" {
					i += 4
					continue
				}
			}
			continue
		}

		filtered = append(filtered, param)
	}
	return filtered
}

func isANSIBackgroundCode(code int) bool {
	return (code >= 40 && code <= 47) || (code >= 100 && code <= 107)
}
