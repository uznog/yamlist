package render

import (
	"fmt"
	"strings"

	"github.com/vznog/yamlist/internal/model"
)

// PreviewRenderer handles rendering of the preview pane
type PreviewRenderer struct {
	Styles       *Styles
	MaxLines     int
	ShowLineNums bool
}

// NewPreviewRenderer creates a new preview renderer
func NewPreviewRenderer(styles *Styles, maxLines int) *PreviewRenderer {
	return &PreviewRenderer{
		Styles:       styles,
		MaxLines:     maxLines,
		ShowLineNums: false,
	}
}

// RenderPreview renders the preview for a node
func (p *PreviewRenderer) RenderPreview(node *model.Node, width, height int) string {
	if node == nil {
		return p.Styles.NullValue.Render("(no selection)")
	}

	var b strings.Builder

	// Header: path
	path := "(root)"
	if node.Path != nil {
		path = node.Path.String()
	}
	b.WriteString(p.Styles.PreviewPath.Render(path))
	b.WriteString("\n")

	// Type info
	typeInfo := p.formatTypeInfo(node)
	b.WriteString(p.Styles.ChildCount.Render(typeInfo))
	b.WriteString("\n\n")

	// Content
	content := p.renderContent(node, width, height-4) // Reserve lines for header
	b.WriteString(content)

	return b.String()
}

// formatTypeInfo returns a type description string
func (p *PreviewRenderer) formatTypeInfo(node *model.Node) string {
	switch node.Kind {
	case model.KindMap:
		return fmt.Sprintf("map (%d keys)", len(node.Children))
	case model.KindList:
		return fmt.Sprintf("list (%d items)", len(node.Children))
	case model.KindScalar:
		typeStr := fmt.Sprintf("scalar (%s)", node.ScalarType.String())
		// Add line count for string values
		if node.ScalarType == model.ScalarString && strings.Contains(node.ScalarValue, "\n") {
			lineCount := strings.Count(node.ScalarValue, "\n") + 1
			typeStr += fmt.Sprintf(" · %d lines", lineCount)
		}
		return typeStr
	default:
		return "unknown"
	}
}

// renderContent renders the main content based on node type
func (p *PreviewRenderer) renderContent(node *model.Node, width, maxLines int) string {
	switch node.Kind {
	case model.KindScalar:
		return p.renderScalar(node, width, maxLines)
	case model.KindMap:
		return p.renderMap(node, width, maxLines)
	case model.KindList:
		return p.renderList(node, width, maxLines)
	default:
		return ""
	}
}

// renderScalar renders a scalar value preview
func (p *PreviewRenderer) renderScalar(node *model.Node, width, maxLines int) string {
	value := node.ScalarValue

	// Format based on type
	switch node.ScalarType {
	case model.ScalarNull:
		return p.Styles.NullValue.Render("null")
	case model.ScalarBool:
		return p.Styles.BoolValue.Render(value)
	case model.ScalarInt, model.ScalarFloat:
		return p.Styles.NumberValue.Render(value)
	case model.ScalarTimestamp:
		return p.Styles.TimestampValue.Render(value)
	case model.ScalarString:
		// Handle multiline strings
		if strings.Contains(value, "\n") {
			return p.renderMultilineString(value, width, maxLines)
		}
		return p.Styles.StringValue.Render(value)
	}

	return value
}

// renderMultilineString renders a multiline string with line numbers
func (p *PreviewRenderer) renderMultilineString(value string, width, maxLines int) string {
	lines := strings.Split(value, "\n")
	var b strings.Builder

	lineNumWidth := len(fmt.Sprintf("%d", len(lines)))
	displayLines := lines
	if len(displayLines) > maxLines {
		displayLines = displayLines[:maxLines]
	}

	for i, line := range displayLines {
		if p.ShowLineNums {
			lineNum := fmt.Sprintf("%*d ", lineNumWidth, i+1)
			b.WriteString(p.Styles.ChildCount.Render(lineNum))
		}

		// Truncate long lines
		if len(line) > width-lineNumWidth-2 {
			line = line[:width-lineNumWidth-5] + "..."
		}

		b.WriteString(p.Styles.StringValue.Render(line))
		if i < len(displayLines)-1 {
			b.WriteString("\n")
		}
	}

	if len(lines) > maxLines {
		b.WriteString("\n")
		b.WriteString(p.Styles.ChildCount.Render(fmt.Sprintf("... (%d more lines)", len(lines)-maxLines)))
	}

	return b.String()
}

// renderMap renders a map preview
func (p *PreviewRenderer) renderMap(node *model.Node, width, maxLines int) string {
	var b strings.Builder
	displayCount := len(node.Children)
	if displayCount > maxLines {
		displayCount = maxLines
	}

	for i := 0; i < displayCount; i++ {
		child := node.Children[i]
		key := p.Styles.Key.Render(child.Key)
		b.WriteString(key)
		b.WriteString(": ")

		if child.Kind == model.KindScalar {
			value := p.formatPreviewValue(child)
			b.WriteString(value)
		} else {
			info := p.formatTypeInfo(child)
			b.WriteString(p.Styles.ChildCount.Render(info))
		}

		if i < displayCount-1 {
			b.WriteString("\n")
		}
	}

	if len(node.Children) > maxLines {
		b.WriteString("\n")
		b.WriteString(p.Styles.ChildCount.Render(fmt.Sprintf("... (%d more keys)", len(node.Children)-maxLines)))
	}

	return b.String()
}

// renderList renders a list preview
func (p *PreviewRenderer) renderList(node *model.Node, width, maxLines int) string {
	var b strings.Builder
	displayCount := len(node.Children)
	if displayCount > maxLines {
		displayCount = maxLines
	}

	for i := 0; i < displayCount; i++ {
		child := node.Children[i]
		index := p.Styles.ChildCount.Render(fmt.Sprintf("[%d] ", i))
		b.WriteString(index)

		if child.Kind == model.KindScalar {
			value := p.formatPreviewValue(child)
			b.WriteString(value)
		} else {
			info := p.formatTypeInfo(child)
			b.WriteString(p.Styles.ChildCount.Render(info))
		}

		if i < displayCount-1 {
			b.WriteString("\n")
		}
	}

	if len(node.Children) > maxLines {
		b.WriteString("\n")
		b.WriteString(p.Styles.ChildCount.Render(fmt.Sprintf("... (%d more items)", len(node.Children)-maxLines)))
	}

	return b.String()
}

// formatPreviewValue formats a scalar value for preview
func (p *PreviewRenderer) formatPreviewValue(node *model.Node) string {
	value := node.ScalarValue
	maxLen := 40

	// Truncate
	if len(value) > maxLen {
		value = value[:maxLen-3] + "..."
	}

	// Replace newlines for inline display
	value = strings.ReplaceAll(value, "\n", "\\n")

	switch node.ScalarType {
	case model.ScalarNull:
		return p.Styles.NullValue.Render("null")
	case model.ScalarBool:
		return p.Styles.BoolValue.Render(value)
	case model.ScalarInt, model.ScalarFloat:
		return p.Styles.NumberValue.Render(value)
	case model.ScalarTimestamp:
		return p.Styles.TimestampValue.Render(value)
	default:
		return p.Styles.StringValue.Render(value)
	}
}
