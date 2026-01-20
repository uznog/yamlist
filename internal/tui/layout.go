package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// StatusBarHeight is the height of the status bar
	StatusBarHeight = 1

	// SearchBarHeight is the height of the search bar when visible
	SearchBarHeight = 1
)

// updateLayout recalculates pane dimensions
func (m *Model) updateLayout() {
	// Full-width tree (no preview pane)
	m.TreeWidth = m.Width
	m.PreviewWidth = 0
}

// renderLayout renders the complete layout
func (m *Model) renderLayout() string {
	// Calculate content height
	contentHeight := m.Height - StatusBarHeight
	// Show search bar when in search mode OR when search is active (confirmed with Enter)
	showSearchBar := m.Mode == SearchMode || m.SearchActive
	if showSearchBar {
		contentHeight -= SearchBarHeight
	}

	// Render tree pane (full width, no preview)
	mainContent := m.renderTreePane(contentHeight)

	// Build final layout
	var b strings.Builder
	b.WriteString(mainContent)
	b.WriteString("\n")

	// Search bar (if in search mode or search is active)
	if showSearchBar {
		b.WriteString(m.renderSearchBar())
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderTreePane renders the tree view pane
func (m *Model) renderTreePane(height int) string {
	var lines []string

	// Calculate visible range
	visibleStart := m.TreeState.ScrollOffset
	visibleEnd := visibleStart + height
	if visibleEnd > len(m.TreeState.VisibleRows) {
		visibleEnd = len(m.TreeState.VisibleRows)
	}

	// Render visible rows
	for i := visibleStart; i < visibleEnd; i++ {
		row := m.TreeState.VisibleRows[i]
		row.IsSelected = (i == m.TreeState.SelectedIndex)
		line := m.RowRenderer.FormatRow(row, m.TreeWidth)
		lines = append(lines, truncateOrPad(line, m.TreeWidth))
	}

	// Pad with empty lines if needed
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", m.TreeWidth))
	}

	return strings.Join(lines, "\n")
}

// renderPreviewPane renders the preview pane
func (m *Model) renderPreviewPane(height int) string {
	// Get selected node
	var selectedNode *Model
	row := m.TreeState.GetSelectedRow()
	var node = m.Document.Root
	if row != nil {
		node = row.Node
	}

	// Render preview
	content := m.PreviewRenderer.RenderPreview(node, m.PreviewWidth, height)

	// Split into lines and pad
	lines := strings.Split(content, "\n")
	var result []string
	for i := 0; i < height; i++ {
		if i < len(lines) {
			result = append(result, truncateOrPad(lines[i], m.PreviewWidth))
		} else {
			result = append(result, strings.Repeat(" ", m.PreviewWidth))
		}
	}

	_ = selectedNode // unused, keeping for potential future use
	return strings.Join(result, "\n")
}

// renderSeparator renders the vertical separator between panes
func (m *Model) renderSeparator(height int) string {
	sep := m.Styles.TreeLine.Render("│")
	lines := make([]string, height)
	for i := range lines {
		lines[i] = " " + sep + " "
	}
	return strings.Join(lines, "\n")
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	// Mode indicator
	modeStr := "TREE"
	if m.Mode == SearchMode {
		modeStr = "SEARCH"
	}
	mode := m.Styles.StatusMode.Render(modeStr)

	// Help hint
	help := m.Styles.StatusInfo.Render("j/k:nav h/l:fold n/N:match /:search q:quit")

	// Path section - show full path of selected node
	var pathStr string
	if m.Error != "" {
		pathStr = m.Styles.MatchHighlight.Render(m.Error)
	} else {
		row := m.TreeState.GetSelectedRow()
		if row != nil {
			pathStr = row.PathString()
		}
	}

	// Calculate available width for path
	modeWidth := lipgloss.Width(mode) + 1 // +1 for space after mode
	helpWidth := lipgloss.Width(help)
	availableWidth := m.Width - modeWidth - helpWidth - 4 // 4 for padding/spaces

	// Truncate path with middle-ellipsis if needed
	if availableWidth > 0 && len(pathStr) > availableWidth {
		pathStr = truncatePathMiddle(pathStr, availableWidth)
	}

	pathRendered := m.Styles.StatusInfo.Render(pathStr)

	// Combine
	leftPart := mode + " " + pathRendered

	// Calculate padding
	padding := m.Width - lipgloss.Width(leftPart) - helpWidth
	if padding < 0 {
		padding = 0
	}

	return m.Styles.StatusBar.Render(
		leftPart + strings.Repeat(" ", padding) + help,
	)
}

// renderSearchBar renders the search input bar
func (m *Model) renderSearchBar() string {
	prompt := m.Styles.SearchPrompt.Render("/")

	// In tree mode with active search, show the term without cursor (non-editable display)
	var input string
	if m.Mode == SearchMode {
		input = m.SearchInput.View()
	} else {
		// Just show the search term text (no cursor/editing)
		input = m.Styles.SearchPrompt.Render(m.SearchInput.Value())
	}

	// Match count
	matchInfo := ""
	if len(m.SearchMatches) > 0 {
		matchInfo = m.Styles.MatchCount.Render(
			formatMatchInfo(m.SearchIndex+1, len(m.SearchMatches)),
		)
	}

	return prompt + input + " " + matchInfo
}

// truncateOrPad ensures a string is exactly the given width
func truncateOrPad(s string, width int) string {
	visWidth := lipgloss.Width(s)
	if visWidth > width {
		// Truncate - this is approximate due to ANSI codes
		return s[:width]
	}
	if visWidth < width {
		return s + strings.Repeat(" ", width-visWidth)
	}
	return s
}

// formatPosition formats a position string like "5/42"
func formatPosition(current, total int) string {
	return strings.Join([]string{
		intToString(current),
		"/",
		intToString(total),
	}, "")
}

// formatMatchInfo formats match info like "[3/15]"
func formatMatchInfo(current, total int) string {
	return "[" + intToString(current) + "/" + intToString(total) + "]"
}

// intToString converts int to string without imports
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// truncatePathMiddle truncates a path string with middle-ellipsis
// Keeps the first segment and last 2 meaningful segments
// Example: "metadata.spec.containers[0].env[2].name" → "metadata...env[2].name"
func truncatePathMiddle(path string, maxWidth int) string {
	if len(path) <= maxWidth {
		return path
	}

	// Need at least space for "..."
	if maxWidth < 4 {
		return path[:maxWidth]
	}

	// Split path into segments (by . and [)
	segments := splitPathSegments(path)
	if len(segments) <= 3 {
		// Too few segments - just truncate from end
		return path[:maxWidth-3] + "..."
	}

	// Keep first segment + "..." + last 2 segments
	first := segments[0]
	last2 := strings.Join(segments[len(segments)-2:], "")

	result := first + "..." + last2
	if len(result) <= maxWidth {
		return result
	}

	// Still too long - truncate the end portion
	availableEnd := maxWidth - len(first) - 3 // 3 for "..."
	if availableEnd < 1 {
		return path[:maxWidth-3] + "..."
	}

	return first + "..." + last2[len(last2)-availableEnd:]
}

// splitPathSegments splits a path into segments, keeping delimiters attached
// e.g., "a.b[0].c" → ["a", ".b", "[0]", ".c"]
func splitPathSegments(path string) []string {
	var segments []string
	var current strings.Builder

	for i, r := range path {
		if r == '.' || r == '[' {
			if current.Len() > 0 {
				segments = append(segments, current.String())
				current.Reset()
			}
			if r == '.' && i > 0 {
				current.WriteRune(r)
			} else if r == '[' {
				current.WriteRune(r)
			}
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		segments = append(segments, current.String())
	}

	return segments
}
