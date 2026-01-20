package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/uznog/yamlist/internal/model"
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
		if m.ViewMode == TreeView {
			m.collapseSelected()
		}
	case "l", "right":
		if m.ViewMode == TreeView {
			m.expandSelected()
		}
	case "enter", " ":
		if m.ViewMode == TreeView {
			m.toggleExpand()
		}

	// Collapse/expand all
	case "z":
		if m.ViewMode == TreeView {
			m.collapseAll()
		}
	case "Z":
		if m.ViewMode == TreeView {
			m.expandAll()
		}

	// Page navigation
	case "ctrl+d":
		m.pageDown()
	case "ctrl+u":
		m.pageUp()
	case "g":
		m.goToTop()
	case "G":
		m.goToBottom()

	// Search match navigation (works even outside search mode)
	case "n":
		m.nextMatch()
	case "N":
		m.prevMatch()

	// Search
	case "/":
		return m.enterSearchMode()

	// Toggle view mode (tree <-> flat)
	case "tab":
		return m.toggleViewMode()

	// Clear search / Quit
	case "esc":
		m.clearSearch()
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// handleSearchKey handles key input in search mode
func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Esc clears search completely
		return m.exitSearchMode(false)

	case "enter":
		// Enter confirms search, keeps highlighting
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
	// If there's an existing search, keep it and allow editing
	// Only reset if starting fresh (no active search)
	if !m.SearchActive {
		m.SearchInput.Reset()
		m.SearchMatches = nil
		m.SearchIndex = 0
	}
	m.SearchInput.Focus()
	m.SearchActive = true
	m.updateRowDimming()
	return m, nil
}

// exitSearchMode exits search mode
// If keep is true, keep the search results highlighted (Enter)
// If keep is false, clear search completely (Esc)
func (m *Model) exitSearchMode(keep bool) (tea.Model, tea.Cmd) {
	m.Mode = TreeMode
	m.SearchInput.Blur()

	if keep {
		// Enter: keep highlighting, jump to current match
		m.SearchActive = len(m.SearchMatches) > 0
		if len(m.SearchMatches) > 0 {
			if m.ViewMode == FlatView {
				// In flat mode, re-apply filter to keep only matching rows
				m.filterVisibleRowsToMatches()
				// Select the current match in the filtered view
				match := m.SearchMatches[m.SearchIndex]
				m.TreeState.SelectNode(match.Node)
				m.ensureSelectedVisible()
			} else {
				// In tree mode, jump to node (which expands ancestors)
				match := m.SearchMatches[m.SearchIndex]
				m.jumpToNode(match.Node)
			}
		}
		// Re-apply dimming after jumpToNode (which may recreate rows)
		m.updateRowDimming()
	} else {
		// Esc: clear search completely
		m.clearSearch()
	}

	return m, nil
}

// clearSearch clears the search state and removes all highlighting
func (m *Model) clearSearch() {
	m.SearchInput.Reset()
	m.SearchMatches = nil
	m.SearchIndex = 0
	m.SearchActive = false
	m.updateRowDimming()
}

// toggleViewMode switches between tree and flat view
func (m *Model) toggleViewMode() (tea.Model, tea.Cmd) {
	// Save current selection path
	var selectedPath *model.Path
	if row := m.TreeState.GetSelectedRow(); row != nil {
		selectedPath = row.Node.Path
	}

	// Toggle view mode
	if m.ViewMode == TreeView {
		m.ViewMode = FlatView
	} else {
		m.ViewMode = TreeView
	}

	// Recompute rows for new view
	m.computeVisibleRows()

	// Restore selection by path
	if selectedPath != nil {
		if !m.TreeState.SelectByPath(selectedPath) {
			// Fallback to first row if selection lost
			m.TreeState.SelectedIndex = 0
			if len(m.TreeState.VisibleRows) > 0 {
				m.TreeState.SelectedNode = m.TreeState.VisibleRows[0].Node
			}
		}
	}

	m.ensureSelectedVisible()
	m.notifyLineChange()
	return m, nil
}
