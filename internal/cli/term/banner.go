package term

import "strings"

var bannerLines = []string{
	"   █████╗  ██████╗██╗  ██╗ ██████╗      ███╗   ███╗ ██████╗██████╗",
	"  ██╔══██╗██╔════╝██║  ██║██╔═══██╗     ████╗ ████║██╔════╝██╔══██╗",
	"  ███████║██║     ███████║██║   ██║     ██╔████╔██║██║     ██████╔╝",
	"  ██╔══██║██║     ██╔══██║██║   ██║     ██║╚██╔╝██║██║     ██╔═══╝",
	"  ██║  ██║╚██████╗██║  ██║╚██████╔╝     ██║ ╚═╝ ██║╚██████╗██║",
	"  ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝      ╚═╝     ╚═╝ ╚═════╝╚═╝",
}

func Banner() string {
	var b strings.Builder
	b.WriteString("\n")
	for _, line := range bannerLines {
		b.WriteString(colorizeBannerLine(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString("  " + T.Secondary() + "acho! don't make me repeat this again" + T.Reset() + "\n")
	b.WriteString("  " + T.Muted() + "Persistent memory for AI coding agents" + T.Reset() + "\n")
	b.WriteString("\n")
	return b.String()
}

func colorizeBannerLine(line string) string {
	var b strings.Builder
	b.WriteString(T.Primary())
	inShadow := false
	for _, r := range line {
		isShadowChar := r == '╗' || r == '╝' || r == '╚' || r == '╔' || r == '═' || r == '║'
		if isShadowChar && !inShadow {
			b.WriteString(T.Border())
			inShadow = true
		} else if !isShadowChar && inShadow {
			b.WriteString(T.Primary())
			inShadow = false
		}
		b.WriteRune(r)
	}
	b.WriteString(T.Reset())
	return b.String()
}
