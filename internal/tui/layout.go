package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// TreeWidthPercent is the percentage of width for the tree pane
	TreeWidthPercent = 60

	// MinTreeWidth is the minimum tree pane width
	MinTreeWidth = 30

	// MinPreviewWidth is the minimum preview pane width
	MinPreviewWidth = 20

	// StatusBarHeight is the height of the status bar
	StatusBarHeight = 1

	// SearchBarHeight is the height of the search bar when visible
	SearchBarHeight = 1
)

// updateLayout recalculates pane dimensions
func (m *Model) updateLayout() {
	// Calculate tree and preview widths
	m.TreeWidth = (m.Width * TreeWidthPercent) / 100
	if m.TreeWidth < MinTreeWidth {
		m.TreeWidth = MinTreeWidth
	}

	m.PreviewWidth = m.Width - m.TreeWidth - 3 // 3 for separator
	if m.PreviewWidth < MinPreviewWidth {
		m.PreviewWidth = MinPreviewWidth
		m.TreeWidth = m.Width - m.PreviewWidth - 3
	}
}

// renderLayout renders the complete layout
func (m *Model) renderLayout() string {
	// Calculate content height
	contentHeight := m.Height - StatusBarHeight
	if m.Mode == SearchMode {
		contentHeight -= SearchBarHeight
	}

	// Render tree pane
	treeContent := m.renderTreePane(contentHeight)

	// Render preview pane
	previewContent := m.renderPreviewPane(contentHeight)

	// Combine panes side by side
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		treeContent,
		m.renderSeparator(contentHeight),
		previewContent,
	)

	// Build final layout
	var b strings.Builder
	b.WriteString(mainContent)
	b.WriteString("\n")

	// Search bar (if in search mode)
	if m.Mode == SearchMode {
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

	// Info section
	var info string
	if m.Error != "" {
		info = m.Styles.MatchHighlight.Render(m.Error)
	} else {
		// Show current position
		total := len(m.TreeState.VisibleRows)
		current := m.TreeState.SelectedIndex + 1
		if total > 0 {
			info = m.Styles.StatusInfo.Render(
				strings.Join([]string{
					string(rune('0' + current%10)),
					"/",
					string(rune('0' + total%10)),
				}, ""),
			)
			// Proper formatting
			info = m.Styles.StatusInfo.Render(
				formatPosition(current, total),
			)
		}
	}

	// Help hint
	help := m.Styles.StatusInfo.Render("j/k:nav h/l:fold /:search q:quit")

	// Combine
	leftPart := mode + " " + info
	rightPart := help

	// Calculate padding
	padding := m.Width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
	if padding < 0 {
		padding = 0
	}

	return m.Styles.StatusBar.Render(
		leftPart + strings.Repeat(" ", padding) + rightPart,
	)
}

// renderSearchBar renders the search input bar
func (m *Model) renderSearchBar() string {
	prompt := m.Styles.SearchPrompt.Render("/")
	input := m.SearchInput.View()

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
