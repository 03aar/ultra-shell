package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Padding(0, 1)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Bold(true)
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	riskLow      = lipgloss.NewStyle().Background(lipgloss.Color("42")).Foreground(lipgloss.Color("0")).Padding(0, 1).Bold(true)
	riskMedium   = lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("0")).Padding(0, 1).Bold(true)
	riskHigh     = lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("255")).Padding(0, 1).Bold(true)
	riskCritical = lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("255")).Padding(0, 1).Bold(true).Blink(true)

	NoColor bool
)

func StyleHeader(text string) string {
	if NoColor {
		return "=== " + text + " ==="
	}
	return headerStyle.Render(text)
}

func StyleSuccess(text string) string {
	if NoColor {
		return "[OK] " + text
	}
	return successStyle.Render("✓ " + text)
}

func StyleError(text string) string {
	if NoColor {
		return "[ERROR] " + text
	}
	return errorStyle.Render("✗ " + text)
}

func StyleWarning(text string) string {
	if NoColor {
		return "[WARN] " + text
	}
	return warningStyle.Render("⚠ " + text)
}

func StyleMuted(text string) string {
	if NoColor {
		return text
	}
	return mutedStyle.Render(text)
}

func StyleLabel(text string) string {
	if NoColor {
		return text
	}
	return labelStyle.Render(text)
}

func StyleValue(text string) string {
	if NoColor {
		return text
	}
	return valueStyle.Render(text)
}

func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Header row
	headerParts := make([]string, len(headers))
	for i, h := range headers {
		padded := h + strings.Repeat(" ", colWidths[i]-len(h))
		if NoColor {
			headerParts[i] = padded
		} else {
			headerParts[i] = labelStyle.Render(padded)
		}
	}
	sb.WriteString(strings.Join(headerParts, "  "))
	sb.WriteString("\n")

	// Separator
	sepParts := make([]string, len(headers))
	for i := range headers {
		sep := strings.Repeat("─", colWidths[i])
		if NoColor {
			sepParts[i] = sep
		} else {
			sepParts[i] = borderStyle.Render(sep)
		}
	}
	sb.WriteString(strings.Join(sepParts, "  "))
	sb.WriteString("\n")

	// Data rows
	for _, row := range rows {
		parts := make([]string, len(headers))
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			padded := cell + strings.Repeat(" ", colWidths[i]-len(cell))
			parts[i] = padded
		}
		sb.WriteString(strings.Join(parts, "  "))
		sb.WriteString("\n")
	}

	return sb.String()
}

func FormatRiskBadge(level string) string {
	if NoColor {
		return "[" + strings.ToUpper(level) + "]"
	}
	switch strings.ToLower(level) {
	case "low":
		return riskLow.Render("LOW")
	case "medium", "med":
		return riskMedium.Render("MED")
	case "high":
		return riskHigh.Render("HIGH")
	case "critical":
		return riskCritical.Render("CRIT")
	default:
		return mutedStyle.Render(strings.ToUpper(level))
	}
}

func FormatExitCode(code int) string {
	s := fmt.Sprintf("%d", code)
	if NoColor {
		return s
	}
	if code == 0 {
		return successStyle.Render(s)
	}
	return errorStyle.Render(s)
}

func FormatKeyValue(key, value string) string {
	return StyleLabel(key+": ") + StyleValue(value)
}

func FormatBool(val bool) string {
	if val {
		if NoColor {
			return "yes"
		}
		return successStyle.Render("yes")
	}
	if NoColor {
		return "no"
	}
	return errorStyle.Render("no")
}
