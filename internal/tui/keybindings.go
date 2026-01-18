package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleTreeKey handles key input in tree mode
func (m *Model) handleTreeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	// Navigation
	case "j", "down":
		m.moveDown(1)
	case "k", "up":
		m.moveUp(1)

	// Expand/collapse
	case "h", "left":
		m.collapseSelected()
	case "l", "right":
		m.expandSelected()
	case "enter", " ":
		m.toggleExpand()

	// Collapse/expand all
	case "z":
		m.collapseAll()
	case "Z":
		m.expandAll()

	// Page navigation
	case "ctrl+d":
		m.pageDown()
	case "ctrl+u":
		m.pageUp()
	case "g":
		m.goToTop()
	case "G":
		m.goToBottom()

	// Search
	case "/":
		return m.enterSearchMode()

	// Quit
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// handleSearchKey handles key input in search mode
func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.exitSearchMode(false)

	case "enter":
		return m.exitSearchMode(true)

	case "ctrl+n", "down":
		m.nextMatch()
		return m, nil

	case "ctrl+p", "up":
		m.prevMatch()
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	default:
		// Pass to text input
		var cmd tea.Cmd
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		m.updateSearchMatches()
		return m, cmd
	}
}

// enterSearchMode switches to search mode
func (m *Model) enterSearchMode() (tea.Model, tea.Cmd) {
	m.Mode = SearchMode
	m.SearchInput.Reset()
	m.SearchInput.Focus()
	m.SearchMatches = nil
	m.SearchIndex = 0
	return m, nil
}

// exitSearchMode exits search mode
// If jump is true, jump to the selected match
func (m *Model) exitSearchMode(jump bool) (tea.Model, tea.Cmd) {
	m.Mode = TreeMode
	m.SearchInput.Blur()

	if jump && len(m.SearchMatches) > 0 {
		// Jump to selected match
		match := m.SearchMatches[m.SearchIndex]
		m.jumpToNode(match.Node)
	}

	return m, nil
}
