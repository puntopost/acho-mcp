package term

import (
	"strings"
)

// FormatHelp colorizes a plain-text help string automatically.
func FormatHelp(text string) string {
	var b strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// First line: "acho xxx — description" → command primary, description white
		if i == 0 && strings.HasPrefix(trimmed, "acho ") {
			parts := strings.SplitN(trimmed, " — ", 2)
			if len(parts) == 2 {
				b.WriteString(T.Bold() + T.Primary() + parts[0] + T.Reset() + " — " + parts[1] + "\n")
			} else {
				b.WriteString(T.Bold() + T.Primary() + trimmed + T.Reset() + "\n")
			}
			continue
		}

		// Section headers: "Usage:", "Flags:", "Arguments:", "Examples:", etc.
		if isSection(trimmed) {
			b.WriteString(T.Bold() + T.Primary() + trimmed + T.Reset() + "\n")
			continue
		}

		// Lines containing --flag → colorize the flag part
		if strings.Contains(line, "--") {
			b.WriteString(colorizeFlags(line) + "\n")
			continue
		}

		// Example lines starting with "acho "
		if strings.HasPrefix(trimmed, "acho ") {
			b.WriteString(T.Muted() + line + T.Reset() + "\n")
			continue
		}

		b.WriteString(line + "\n")
	}

	return b.String()
}

func isSection(line string) bool {
	sections := []string{"Usage:", "Flags:", "Arguments:", "Examples:", "Options:"}
	for _, s := range sections {
		if strings.HasPrefix(line, s) {
			return true
		}
	}
	return false
}

func colorizeFlags(line string) string {
	var b strings.Builder
	words := strings.Fields(line)
	indent := leadingSpaces(line)
	b.WriteString(indent)

	for i, word := range words {
		if i > 0 {
			b.WriteString(" ")
		}
		if strings.HasPrefix(word, "--") {
			b.WriteString(T.Secondary() + word + T.Reset())
		} else {
			b.WriteString(word)
		}
	}
	return b.String()
}

func leadingSpaces(s string) string {
	for i, c := range s {
		if c != ' ' && c != '\t' {
			return s[:i]
		}
	}
	return s
}
