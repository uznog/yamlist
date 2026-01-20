package tui

import (
	"strings"

	"github.com/uznog/yamlist/internal/model"
)

// updateSearchMatches updates the search matches based on current input
func (m *Model) updateSearchMatches() {
	query := m.SearchInput.Value()
	if query == "" {
		m.SearchMatches = nil
		m.SearchIndex = 0
		m.SearchActive = false
		m.computeVisibleRows() // Reset to full view
		m.updateRowDimming()
		return
	}

	// Use case-insensitive substring search on node key name only
	queryLower := strings.ToLower(query)
	m.SearchMatches = make([]*model.PathEntry, 0)

	for i := 0; i < m.Document.Index.Len(); i++ {
		entry := m.Document.Index.EntryAt(i)
		// Search only on the key name, not the full path
		if entry.Node != nil && strings.Contains(strings.ToLower(entry.Node.Key), queryLower) {
			m.SearchMatches = append(m.SearchMatches, entry)
		}
	}

	m.SearchActive = true

	// Reset index if out of bounds
	if m.SearchIndex >= len(m.SearchMatches) {
		m.SearchIndex = 0
	}

	// Filter visible rows to show only matches
	m.filterVisibleRowsToMatches()

	// Jump to first match if available
	if len(m.SearchMatches) > 0 && m.TreeState.SelectedIndex >= len(m.TreeState.VisibleRows) {
		m.previewMatch(m.SearchIndex)
	}

	// Update row dimming AFTER previewMatch (which may call computeVisibleRows)
	m.updateRowDimming()
}

// updateRowDimming updates the IsDimmed and IsSearchMatch flags on all visible rows
func (m *Model) updateRowDimming() {
	// If search is active but no matches, dim all rows
	if m.SearchActive && len(m.SearchMatches) == 0 && m.SearchInput.Value() != "" {
		for _, row := range m.TreeState.VisibleRows {
			row.IsDimmed = true
			row.IsSearchMatch = false
		}
		return
	}

	// Otherwise, no dimming (rows are filtered, not dimmed)
	for _, row := range m.TreeState.VisibleRows {
		row.IsDimmed = false
		row.IsSearchMatch = false
	}
}

// filterVisibleRowsToMatches filters the visible rows to show only matches (and their ancestors in tree mode)
func (m *Model) filterVisibleRowsToMatches() {
	if len(m.SearchMatches) == 0 {
		// No matches - keep all rows but dim them
		m.updateRowDimming()
		return
	}

	// Build set of matching paths
	matchPaths := make(map[string]bool)
	for _, match := range m.SearchMatches {
		if match.Node != nil && match.Node.Path != nil {
			matchPaths[match.Node.Path.String()] = true
		}
	}

	// First, recompute the full visible rows list
	if m.ViewMode == FlatView {
		m.computeFlatRows()
	} else {
		m.TreeState.VisibleRows = make([]*model.VisibleRow, 0)
		m.computeVisibleRowsRecursive(m.Document.Root, 0)
	}

	// Filter VisibleRows
	if m.ViewMode == FlatView {
		// In flat mode: show only matching nodes
		filtered := make([]*model.VisibleRow, 0)
		for _, row := range m.TreeState.VisibleRows {
			if matchPaths[row.Node.Path.String()] {
				filtered = append(filtered, row)
			}
		}
		m.TreeState.VisibleRows = filtered
	} else {
		// In tree mode: show matches + their ancestors (for context)
		filtered := make([]*model.VisibleRow, 0)
		for _, row := range m.TreeState.VisibleRows {
			if matchPaths[row.Node.Path.String()] {
				filtered = append(filtered, row)
			} else {
				// Check if this row is an ancestor of any match
				for _, match := range m.SearchMatches {
					if match.Node != nil && row.Node.Path.IsAncestorOf(match.Node.Path) {
						filtered = append(filtered, row)
						break
					}
				}
			}
		}
		m.TreeState.VisibleRows = filtered
	}

	// Recalculate indices
	for i, row := range m.TreeState.VisibleRows {
		row.Index = i
	}

	// Ensure selection is valid
	if m.TreeState.SelectedIndex >= len(m.TreeState.VisibleRows) {
		m.TreeState.SelectedIndex = len(m.TreeState.VisibleRows) - 1
	}
	if m.TreeState.SelectedIndex < 0 {
		m.TreeState.SelectedIndex = 0
	}
	if len(m.TreeState.VisibleRows) > 0 && m.TreeState.SelectedIndex < len(m.TreeState.VisibleRows) {
		m.TreeState.SelectedNode = m.TreeState.VisibleRows[m.TreeState.SelectedIndex].Node
	}
}

// nextMatch moves to the next search match
func (m *Model) nextMatch() {
	if len(m.SearchMatches) == 0 {
		return
	}

	m.SearchIndex = (m.SearchIndex + 1) % len(m.SearchMatches)
	m.previewMatch(m.SearchIndex)
}

// prevMatch moves to the previous search match
func (m *Model) prevMatch() {
	if len(m.SearchMatches) == 0 {
		return
	}

	m.SearchIndex--
	if m.SearchIndex < 0 {
		m.SearchIndex = len(m.SearchMatches) - 1
	}
	m.previewMatch(m.SearchIndex)
}

// previewMatch shows a preview of the match at the given index
func (m *Model) previewMatch(index int) {
	if index < 0 || index >= len(m.SearchMatches) {
		return
	}

	match := m.SearchMatches[index]

	if m.ViewMode == FlatView {
		// In flat mode, just select the node in the filtered view
		m.TreeState.SelectNode(match.Node)
		m.centerSelected()
	} else {
		// In tree mode, expand ancestors and select the node for preview
		m.TreeState.ExpandToNode(match.Node)
		m.computeVisibleRows()
		m.TreeState.SelectNode(match.Node)
		m.centerSelected()
	}

	// Re-apply dimming after computeVisibleRows recreated rows
	m.updateRowDimming()
}

// getCurrentMatch returns the currently selected match
func (m *Model) getCurrentMatch() *model.PathEntry {
	if m.SearchIndex < 0 || m.SearchIndex >= len(m.SearchMatches) {
		return nil
	}
	return m.SearchMatches[m.SearchIndex]
}

// hasMatches returns true if there are search matches
func (m *Model) hasMatches() bool {
	return len(m.SearchMatches) > 0
}

// matchCount returns the number of search matches
func (m *Model) matchCount() int {
	return len(m.SearchMatches)
}
