package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/uznog/yamlist/internal/model"
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
func (r *RowRenderer) FormatRow(row *model.VisibleRow, width int, isFlatMode bool) string {
	var b strings.Builder

	// Determine if row should be dimmed (non-match during active search)
	isDimmed := row.IsDimmed

	if isFlatMode {
		// In flat mode, show full path instead of indentation
		pathStr := row.PathString()
		if row.IsSelected {
			b.WriteString(r.Styles.SelectedKey.Render(pathStr))
		} else if isDimmed {
			b.WriteString(r.Styles.DimmedKey.Render(pathStr))
		} else {
			b.WriteString(r.Styles.Key.Render(pathStr))
		}

		// Add value for scalars
		if row.Kind() == model.KindScalar {
			b.WriteString(": ")
			value := r.formatScalarValue(row.ScalarValue(), row.ScalarType(), row.IsSelected, isDimmed)
			b.WriteString(value)
		}
	} else {
		// Tree mode rendering (existing behavior)
		// Indentation
		indent := strings.Repeat(" ", row.Depth*r.Indent)
		b.WriteString(indent)

		// Expand/collapse icon
		expandIcon := r.Icons.GetExpandIcon(row.IsExpanded, row.IsExpandable)
		if row.IsSelected {
			b.WriteString(expandIcon)
		} else if isDimmed {
			b.WriteString(r.Styles.DimmedRow.Render(expandIcon))
		} else {
			b.WriteString(r.Styles.ExpandIcon.Render(expandIcon))
		}
		b.WriteString(" ")

		// Type icon
		typeIcon := r.Icons.GetTypeIcon(row.Kind(), row.ScalarType())
		if row.IsSelected {
			b.WriteString(typeIcon)
		} else if isDimmed {
			b.WriteString(r.Styles.DimmedRow.Render(typeIcon))
		} else {
			b.WriteString(r.Styles.TypeIcon.Render(typeIcon))
		}
		b.WriteString(" ")

		// Key
		key := row.DisplayKey()
		if row.IsSelected {
			b.WriteString(r.Styles.SelectedKey.Render(key))
		} else if isDimmed {
			b.WriteString(r.Styles.DimmedKey.Render(key))
		} else {
			b.WriteString(r.Styles.Key.Render(key))
		}

		// Value or child count
		if row.Kind() == model.KindScalar {
			b.WriteString(": ")
			value := r.formatScalarValue(row.ScalarValue(), row.ScalarType(), row.IsSelected, isDimmed)
			b.WriteString(value)
		} else if row.HasChildren {
			countStr := fmt.Sprintf(" (%d)", row.ChildCount)
			if row.IsSelected {
				b.WriteString(countStr)
			} else if isDimmed {
				b.WriteString(r.Styles.DimmedRow.Render(countStr))
			} else {
				b.WriteString(r.Styles.ChildCount.Render(countStr))
			}
		}
	}

	content := b.String()

	// Apply row-level styling
	if row.IsSelected {
		if isFlatMode {
			// In flat mode, don't replace first char - just highlight the full row
			// Pad to full width for selection highlight
			contentWidth := lipglossWidth(content)
			if contentWidth < width {
				content = content + strings.Repeat(" ", width-contentWidth)
			}
			return r.Styles.SelectedRow.Render(content)
		} else {
			// In tree mode, add accent marker at start (replace first char with accent)
			accent := r.Styles.SelectionAccent.Render("▌")
			if len(content) > 0 {
				// Replace first character (usually a space) with accent marker
				runes := []rune(content)
				content = accent + string(runes[1:])
			} else {
				content = accent
			}

			// Pad to full width for selection highlight
			contentWidth := lipglossWidth(content)
			if contentWidth < width {
				content = content + strings.Repeat(" ", width-contentWidth)
			}
			return r.Styles.SelectedRow.Render(content)
		}
	}

	return content
}

// formatScalarValue formats a scalar value with appropriate styling
func (r *RowRenderer) formatScalarValue(value string, scalarType model.ScalarType, isSelected bool, isDimmed bool) string {
	displayValue := value

	// Handle null type first
	if scalarType == model.ScalarNull {
		displayValue = "null"
	} else {
		// Handle multiline FIRST (before truncation)
		if strings.Contains(displayValue, "\n") {
			lines := strings.Split(displayValue, "\n")
			if len(lines) > 1 {
				// Show line count summary for multiline values
				displayValue = fmt.Sprintf("[%d lines]", len(lines))
			} else {
				displayValue = strings.TrimSpace(lines[0])
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

	if isDimmed {
		return r.Styles.DimmedRow.Render(displayValue)
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
