package tui

import (
	"github.com/uznog/yamlist/internal/model"
)

// notifyLineChange sends a cursor position update to Neovim if connected
func (m *Model) notifyLineChange() {
	if m.NvimClient == nil {
		return
	}
	row := m.TreeState.GetSelectedRow()
	if row != nil && row.Node.LineNumber > 0 {
		m.NvimClient.SendCursor(row.Node.LineNumber)
	}
}

// computeVisibleRows rebuilds the visible rows list based on current expansion state
func (m *Model) computeVisibleRows() {
	if m.ViewMode == FlatView {
		m.computeFlatRows()
	} else {
		m.TreeState.VisibleRows = make([]*model.VisibleRow, 0)
		m.computeVisibleRowsRecursive(m.Document.Root, 0)

		// Ensure selection is valid
		if m.TreeState.SelectedIndex >= len(m.TreeState.VisibleRows) {
			m.TreeState.SelectedIndex = len(m.TreeState.VisibleRows) - 1
		}
		if m.TreeState.SelectedIndex < 0 {
			m.TreeState.SelectedIndex = 0
		}

		// Update selected node reference
		if len(m.TreeState.VisibleRows) > 0 {
			m.TreeState.SelectedNode = m.TreeState.VisibleRows[m.TreeState.SelectedIndex].Node
		}
	}
}

// computeFlatRows computes all rows in flat mode (full path display)
func (m *Model) computeFlatRows() {
	m.TreeState.VisibleRows = make([]*model.VisibleRow, 0)

	for i := 0; i < m.Document.Index.Len(); i++ {
		entry := m.Document.Index.EntryAt(i)
		// Skip root node (no meaningful path)
		if entry.Node != nil && entry.Node.Path != nil && entry.Node.Path.Depth() > 0 {
			row := model.NewVisibleRow(entry.Node, false, len(m.TreeState.VisibleRows))
			row.Depth = 0 // No indentation in flat mode
			m.TreeState.VisibleRows = append(m.TreeState.VisibleRows, row)
		}
	}

	// Ensure selection is valid (same logic as existing computeVisibleRows)
	if m.TreeState.SelectedIndex >= len(m.TreeState.VisibleRows) {
		m.TreeState.SelectedIndex = len(m.TreeState.VisibleRows) - 1
	}
	if m.TreeState.SelectedIndex < 0 {
		m.TreeState.SelectedIndex = 0
	}
	if len(m.TreeState.VisibleRows) > 0 {
		m.TreeState.SelectedNode = m.TreeState.VisibleRows[m.TreeState.SelectedIndex].Node
	}
}

// computeVisibleRowsRecursive recursively adds visible rows
func (m *Model) computeVisibleRowsRecursive(node *model.Node, depth int) {
	if node == nil {
		return
	}

	// Add this node as a visible row
	isExpanded := m.TreeState.IsExpanded(node.Path)
	rowIndex := len(m.TreeState.VisibleRows)
	row := model.NewVisibleRow(node, isExpanded, rowIndex)
	m.TreeState.VisibleRows = append(m.TreeState.VisibleRows, row)

	// If expanded, add children
	if isExpanded && node.HasChildren() {
		for _, child := range node.Children {
			m.computeVisibleRowsRecursive(child, depth+1)
		}
	}
}

// moveUp moves selection up by n rows
func (m *Model) moveUp(n int) {
	m.TreeState.MoveSelection(-n)
	m.ensureSelectedVisible()
	m.notifyLineChange()
}

// moveDown moves selection down by n rows
func (m *Model) moveDown(n int) {
	m.TreeState.MoveSelection(n)
	m.ensureSelectedVisible()
	m.notifyLineChange()
}

// expandSelected expands the selected node
func (m *Model) expandSelected() bool {
	row := m.TreeState.GetSelectedRow()
	if row == nil || !row.IsExpandable {
		return false
	}

	if row.IsExpanded {
		// Already expanded - move to first child
		moved := m.moveToFirstChild()
		if moved {
			m.notifyLineChange()
		}
		return moved
	}

	// Expand
	m.TreeState.SetExpanded(row.Node.Path, true)
	m.computeVisibleRows()
	return true
}

// collapseSelected collapses the selected node or moves to parent
func (m *Model) collapseSelected() bool {
	row := m.TreeState.GetSelectedRow()
	if row == nil {
		return false
	}

	if row.IsExpandable && row.IsExpanded {
		// Collapse this node
		m.TreeState.SetExpanded(row.Node.Path, false)
		m.computeVisibleRows()
		return true
	}

	// Move to parent
	moved := m.moveToParent()
	if moved {
		m.notifyLineChange()
	}
	return moved
}

// moveToParent moves selection to the parent node
func (m *Model) moveToParent() bool {
	row := m.TreeState.GetSelectedRow()
	if row == nil || row.Node.Parent == nil {
		return false
	}

	// Find parent in visible rows
	for i, vr := range m.TreeState.VisibleRows {
		if vr.Node == row.Node.Parent {
			m.TreeState.SelectedIndex = i
			m.TreeState.SelectedNode = vr.Node
			m.ensureSelectedVisible()
			return true
		}
	}

	return false
}

// moveToFirstChild moves selection to the first child of the selected node
func (m *Model) moveToFirstChild() bool {
	row := m.TreeState.GetSelectedRow()
	if row == nil || !row.HasChildren || !row.IsExpanded {
		return false
	}

	// First child should be the next row
	nextIndex := m.TreeState.SelectedIndex + 1
	if nextIndex < len(m.TreeState.VisibleRows) {
		m.TreeState.SelectedIndex = nextIndex
		m.TreeState.SelectedNode = m.TreeState.VisibleRows[nextIndex].Node
		m.ensureSelectedVisible()
		return true
	}

	return false
}

// expandAll expands all nodes
func (m *Model) expandAll() {
	m.TreeState.ExpandAll()
	m.computeVisibleRows()
}

// collapseAll collapses all nodes
func (m *Model) collapseAll() {
	// Remember current selection
	selectedPath := m.TreeState.SelectedNode.Path

	m.TreeState.CollapseAll()
	m.computeVisibleRows()

	// Try to keep selection or move to nearest ancestor
	if !m.TreeState.SelectByPath(selectedPath) {
		// Find nearest visible ancestor
		for current := m.TreeState.SelectedNode; current != nil; current = current.Parent {
			if m.TreeState.SelectByPath(current.Path) {
				break
			}
		}
	}
}

// toggleExpand toggles expansion of the selected node
func (m *Model) toggleExpand() bool {
	row := m.TreeState.GetSelectedRow()
	if row == nil || !row.IsExpandable {
		return false
	}

	m.TreeState.ToggleExpanded(row.Node.Path)
	m.computeVisibleRows()
	return true
}

// ensureSelectedVisible adjusts scroll offset to keep selection visible
func (m *Model) ensureSelectedVisible() {
	if len(m.TreeState.VisibleRows) == 0 {
		return
	}

	// Calculate visible height (approximate)
	visibleHeight := m.Height - StatusBarHeight
	if m.Mode == SearchMode || m.SearchActive {
		visibleHeight -= SearchBarHeight
	}
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Adjust scroll offset
	if m.TreeState.SelectedIndex < m.TreeState.ScrollOffset {
		m.TreeState.ScrollOffset = m.TreeState.SelectedIndex
	}
	if m.TreeState.SelectedIndex >= m.TreeState.ScrollOffset+visibleHeight {
		m.TreeState.ScrollOffset = m.TreeState.SelectedIndex - visibleHeight + 1
	}
}

// centerSelected centers the selected row in the viewport
func (m *Model) centerSelected() {
	if len(m.TreeState.VisibleRows) == 0 {
		return
	}

	// Calculate visible height
	visibleHeight := m.Height - StatusBarHeight
	if m.Mode == SearchMode || m.SearchActive {
		visibleHeight -= SearchBarHeight
	}
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Center the selection
	m.TreeState.ScrollOffset = m.TreeState.SelectedIndex - visibleHeight/2
	if m.TreeState.ScrollOffset < 0 {
		m.TreeState.ScrollOffset = 0
	}

	// Don't scroll past the end
	maxOffset := len(m.TreeState.VisibleRows) - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.TreeState.ScrollOffset > maxOffset {
		m.TreeState.ScrollOffset = maxOffset
	}
}

// jumpToNode expands all ancestors and selects a specific node
func (m *Model) jumpToNode(node *model.Node) bool {
	if node == nil {
		return false
	}

	// Expand all ancestors
	m.TreeState.ExpandToNode(node)

	// Recompute visible rows
	m.computeVisibleRows()

	// Select the node
	if m.TreeState.SelectNode(node) {
		m.ensureSelectedVisible()
		m.notifyLineChange()
		return true
	}

	return false
}

// pageUp moves up by a page
func (m *Model) pageUp() {
	pageSize := m.Height - StatusBarHeight - 2
	if m.Mode == SearchMode {
		pageSize -= SearchBarHeight
	}
	if pageSize < 1 {
		pageSize = 1
	}
	m.moveUp(pageSize)
}

// pageDown moves down by a page
func (m *Model) pageDown() {
	pageSize := m.Height - StatusBarHeight - 2
	if m.Mode == SearchMode {
		pageSize -= SearchBarHeight
	}
	if pageSize < 1 {
		pageSize = 1
	}
	m.moveDown(pageSize)
}

// goToTop moves to the first row
func (m *Model) goToTop() {
	m.TreeState.SelectedIndex = 0
	if len(m.TreeState.VisibleRows) > 0 {
		m.TreeState.SelectedNode = m.TreeState.VisibleRows[0].Node
	}
	m.TreeState.ScrollOffset = 0
	m.notifyLineChange()
}

// goToBottom moves to the last row
func (m *Model) goToBottom() {
	if len(m.TreeState.VisibleRows) == 0 {
		return
	}
	m.TreeState.SelectedIndex = len(m.TreeState.VisibleRows) - 1
	m.TreeState.SelectedNode = m.TreeState.VisibleRows[m.TreeState.SelectedIndex].Node
	m.ensureSelectedVisible()
	m.notifyLineChange()
}
