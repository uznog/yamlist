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

	// Reset index if out of bounds
	if m.SearchIndex >= len(m.SearchMatches) {
		m.SearchIndex = 0
	}

	// Jump to first match if available
	if len(m.SearchMatches) > 0 {
		m.previewMatch(m.SearchIndex)
	}

	// Update row dimming AFTER previewMatch (which may call computeVisibleRows)
	m.updateRowDimming()
}

// updateRowDimming updates the IsDimmed and IsSearchMatch flags on all visible rows
func (m *Model) updateRowDimming() {
	// Build a set of matching node paths for quick lookup
	matchPaths := make(map[string]bool)
	for _, match := range m.SearchMatches {
		if match.Node != nil && match.Node.Path != nil {
			matchPaths[match.Node.Path.String()] = true
		}
	}

	// Update each visible row
	for _, row := range m.TreeState.VisibleRows {
		if m.SearchActive && len(m.SearchMatches) > 0 {
			// Search is active with matches - dim non-matching rows
			pathStr := ""
			if row.Node.Path != nil {
				pathStr = row.Node.Path.String()
			}
			row.IsSearchMatch = matchPaths[pathStr]
			row.IsDimmed = !row.IsSearchMatch
		} else {
			// No active search - no dimming
			row.IsSearchMatch = false
			row.IsDimmed = false
		}
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

	// Expand ancestors and select the node for preview
	// but don't change the actual tree selection yet
	m.TreeState.ExpandToNode(match.Node)
	m.computeVisibleRows()
	m.TreeState.SelectNode(match.Node)
	m.centerSelected()

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
