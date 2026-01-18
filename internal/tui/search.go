package tui

import (
	"github.com/vznog/yamlist/internal/model"
	"github.com/sahilm/fuzzy"
)

// searchSource wraps PathIndex to implement fuzzy.Source
type searchSource struct {
	index *model.PathIndex
}

func (s searchSource) String(i int) string {
	return s.index.EntryAt(i).DisplayString
}

func (s searchSource) Len() int {
	return s.index.Len()
}

// updateSearchMatches updates the search matches based on current input
func (m *Model) updateSearchMatches() {
	query := m.SearchInput.Value()
	if query == "" {
		m.SearchMatches = nil
		m.SearchIndex = 0
		return
	}

	// Perform fuzzy search
	source := searchSource{index: m.Document.Index}
	results := fuzzy.FindFrom(query, source)

	// Convert results to PathEntries
	m.SearchMatches = make([]*model.PathEntry, len(results))
	for i, result := range results {
		m.SearchMatches[i] = m.Document.Index.EntryAt(result.Index)
	}

	// Reset index if out of bounds
	if m.SearchIndex >= len(m.SearchMatches) {
		m.SearchIndex = 0
	}

	// Jump to first match if available
	if len(m.SearchMatches) > 0 {
		m.previewMatch(m.SearchIndex)
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
	m.ensureSelectedVisible()
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
