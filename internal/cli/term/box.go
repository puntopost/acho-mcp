package term

import (
	"fmt"
	"strings"
)

// Box drawing characters
const (
	topLeft     = "╔"
	topRight    = "╗"
	bottomLeft  = "╚"
	bottomRight = "╝"
	horizontal  = "═"
	vertical    = "║"
	teeLeft     = "╠"
	teeRight    = "╣"
)

// Box renders content inside a colored box frame.
type Box struct {
	width int
	lines []string
}

func NewBox(width int) *Box {
	return &Box{width: width}
}

func (b *Box) Blank() {
	b.lines = append(b.lines, "")
}

func (b *Box) Title(text string) {
	b.lines = append(b.lines, fmt.Sprintf("%s%s%s%s", T.Bold(), T.Secondary(), centerPad(text, b.width-6), T.Reset()))
}

func (b *Box) Separator() {
	b.lines = append(b.lines, "---")
}

func (b *Box) Section(title string) {
	b.lines = append(b.lines, fmt.Sprintf("%s%s%s%s", T.Bold(), T.Accent(), title, T.Reset()))
}

// SplitBar renders a row with a two-tone bar: `active` segment in yellow,
// `deleted` segment in red, sized relative to `max` (= max total across all
// rows). The count column shows both numbers (yellow / red), right-aligned.
func (b *Box) SplitBar(label string, active, deleted, max, labelWidth, countWidth int) {
	inner := b.width - 6
	// counts column width: "A/D" with each field padded to countWidth.
	countsVisible := countWidth*2 + 1
	// indent(2) + gap(2) + gap(2) = 6 around label/bar/counts
	barMax := inner - labelWidth - countsVisible - 6
	if barMax < 1 {
		barMax = 1
	}

	total := active + deleted
	totalLen := 0
	activeLen := 0
	deletedLen := 0
	if max > 0 {
		totalLen = (total * barMax) / max
	}
	if totalLen < 1 && total > 0 {
		totalLen = 1
	}
	if totalLen > barMax {
		totalLen = barMax
	}
	if total > 0 {
		activeLen = (active * totalLen) / total
		if activeLen < 1 && active > 0 {
			activeLen = 1
		}
		if activeLen > totalLen {
			activeLen = totalLen
		}
		deletedLen = totalLen - activeLen
	}

	pad := strings.Repeat(" ", barMax-totalLen)
	activeSeg := strings.Repeat("█", activeLen)
	deletedSeg := strings.Repeat("█", deletedLen)

	counts := fmt.Sprintf("%s%s%*d%s",
		T.Bold(), T.Secondary(), countWidth, active, T.Reset(),
	)
	if deleted > 0 {
		counts += fmt.Sprintf("%s/%s%s%s%d%s",
			T.Muted(), T.Reset(),
			T.Bold(), T.Danger(), deleted, T.Reset(),
		)
	}

	b.lines = append(b.lines, fmt.Sprintf(
		"  %s%-*s%s  %s%s%s%s%s%s  %s",
		T.Muted(), labelWidth, label, T.Reset(),
		T.Secondary(), activeSeg, T.Reset(),
		T.Danger(), deletedSeg, T.Reset(),
		pad+counts,
	))
}

func (b *Box) String() string {
	inner := b.width - 6 // ║ (1) + "  " (2) + content (inner) + "  " (2) + ║ (1) = width
	var sb strings.Builder

	// Top border
	sb.WriteString(fmt.Sprintf("%s%s%s%s%s\n", T.Border(), topLeft, strings.Repeat(horizontal, b.width-2), topRight, T.Reset()))

	for _, line := range b.lines {
		if line == "---" {
			// Separator line
			sb.WriteString(fmt.Sprintf("%s%s%s%s%s\n", T.Border(), teeLeft, strings.Repeat(horizontal, b.width-2), teeRight, T.Reset()))
		} else if line == "" {
			// Blank line
			sb.WriteString(fmt.Sprintf("%s%s%s  %*s  %s%s\n", T.Border(), vertical, T.Reset(), inner, "", T.Border(), vertical+T.Reset()))
		} else {
			// Content line - pad to inner width accounting for ANSI codes
			visible := visibleLen(line)
			padding := inner - visible
			if padding < 0 {
				padding = 0
			}
			sb.WriteString(fmt.Sprintf("%s%s%s  %s%*s  %s%s\n", T.Border(), vertical, T.Reset(), line, padding, "", T.Border(), vertical+T.Reset()))
		}
	}

	// Bottom border
	sb.WriteString(fmt.Sprintf("%s%s%s%s%s\n", T.Border(), bottomLeft, strings.Repeat(horizontal, b.width-2), bottomRight, T.Reset()))

	return sb.String()
}

func centerPad(text string, width int) string {
	if len(text) >= width {
		return text
	}
	left := (width - len(text)) / 2
	right := width - len(text) - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

// visibleLen returns the length of a string excluding ANSI escape sequences.
func visibleLen(s string) int {
	n := 0
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		n++
	}
	return n
}
