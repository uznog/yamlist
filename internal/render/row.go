package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/vznog/yamlist/internal/model"
)

// RowRenderer handles rendering of tree rows
type RowRenderer struct {
	Icons  *IconSet
	Styles *Styles
	Indent int // Spaces per indent level
}

// NewRowRenderer creates a new row renderer
func NewRowRenderer(icons *IconSet, styles *Styles) *RowRenderer {
	return &RowRenderer{
		Icons:  icons,
		Styles: styles,
		Indent: 2,
	}
}

// FormatRow formats a visible row for display
func (r *RowRenderer) FormatRow(row *model.VisibleRow, width int) string {
	var b strings.Builder

	// Indentation
	indent := strings.Repeat(" ", row.Depth*r.Indent)
	b.WriteString(indent)

	// Expand/collapse icon
	expandIcon := r.Icons.GetExpandIcon(row.IsExpanded, row.IsExpandable)
	if row.IsSelected {
		b.WriteString(expandIcon)
	} else {
		b.WriteString(r.Styles.ExpandIcon.Render(expandIcon))
	}
	b.WriteString(" ")

	// Type icon
	typeIcon := r.Icons.GetTypeIcon(row.Kind(), row.ScalarType())
	if row.IsSelected {
		b.WriteString(typeIcon)
	} else {
		b.WriteString(r.Styles.TypeIcon.Render(typeIcon))
	}
	b.WriteString(" ")

	// Key
	key := row.DisplayKey()
	if row.IsSelected {
		b.WriteString(r.Styles.SelectedKey.Render(key))
	} else {
		b.WriteString(r.Styles.Key.Render(key))
	}

	// Value or child count
	if row.Kind() == model.KindScalar {
		b.WriteString(": ")
		value := r.formatScalarValue(row.ScalarValue(), row.ScalarType(), row.IsSelected)
		b.WriteString(value)
	} else if row.HasChildren {
		countStr := fmt.Sprintf(" (%d)", row.ChildCount)
		if row.IsSelected {
			b.WriteString(countStr)
		} else {
			b.WriteString(r.Styles.ChildCount.Render(countStr))
		}
	}

	content := b.String()

	// Apply row-level styling
	if row.IsSelected {
		// Pad to full width for selection highlight
		contentWidth := lipglossWidth(content)
		if contentWidth < width {
			content = content + strings.Repeat(" ", width-contentWidth)
		}
		return r.Styles.SelectedRow.Render(content)
	}

	return content
}

// formatScalarValue formats a scalar value with appropriate styling
func (r *RowRenderer) formatScalarValue(value string, scalarType model.ScalarType, isSelected bool) string {
	displayValue := value

	// Handle null type first
	if scalarType == model.ScalarNull {
		displayValue = "null"
	} else {
		// Handle multiline FIRST (before truncation)
		if strings.Contains(displayValue, "\n") {
			lines := strings.Split(displayValue, "\n")
			firstLine := strings.TrimSpace(lines[0])
			if len(lines) > 1 {
				// Truncate first line if needed, then add line count
				maxFirstLine := 35
				if runeCount(firstLine) > maxFirstLine {
					firstLine = truncateRunes(firstLine, maxFirstLine-3) + "..."
				}
				displayValue = firstLine + fmt.Sprintf(" (+%d lines)", len(lines)-1)
			} else {
				displayValue = firstLine
			}
		}

		// Use rune-based truncation for display width
		maxLen := 50
		if runeCount(displayValue) > maxLen {
			displayValue = truncateRunes(displayValue, maxLen-3) + "..."
		}

		// Escape special chars for inline display
		displayValue = strings.ReplaceAll(displayValue, "\n", "\\n")
		displayValue = strings.ReplaceAll(displayValue, "\t", "\\t")
	}

	if isSelected {
		return displayValue
	}

	style := r.Styles.GetValueStyle(int(scalarType))
	return style.Render(displayValue)
}

// runeCount returns the number of runes in a string
func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

// truncateRunes truncates a string to maxRunes runes
func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

// lipglossWidth calculates the visible width of a string (accounting for ANSI)
func lipglossWidth(s string) int {
	// Simple approximation - count non-ANSI characters
	width := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		width++
	}
	return width
}
