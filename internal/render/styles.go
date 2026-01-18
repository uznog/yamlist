package render

import "github.com/charmbracelet/lipgloss"

// Styles contains all the lipgloss styles for rendering
type Styles struct {
	// Row styles
	SelectedRow   lipgloss.Style
	NormalRow     lipgloss.Style

	// Key styles
	Key           lipgloss.Style
	SelectedKey   lipgloss.Style

	// Value styles by type
	StringValue   lipgloss.Style
	NumberValue   lipgloss.Style
	BoolValue     lipgloss.Style
	NullValue     lipgloss.Style
	TimestampValue lipgloss.Style

	// Structural styles
	ExpandIcon    lipgloss.Style
	TypeIcon      lipgloss.Style
	TreeLine      lipgloss.Style
	ChildCount    lipgloss.Style

	// Preview pane
	PreviewTitle  lipgloss.Style
	PreviewPath   lipgloss.Style
	PreviewBorder lipgloss.Style

	// Search styles
	SearchPrompt  lipgloss.Style
	SearchInput   lipgloss.Style
	MatchCount    lipgloss.Style
	MatchHighlight lipgloss.Style

	// Status bar
	StatusBar     lipgloss.Style
	StatusMode    lipgloss.Style
	StatusInfo    lipgloss.Style
}

// DefaultStyles returns the default color scheme
func DefaultStyles() *Styles {
	return &Styles{
		// Row styles
		SelectedRow: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")),
		NormalRow: lipgloss.NewStyle(),

		// Key styles
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")), // Light blue
		SelectedKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Bold(true),

		// Value styles
		StringValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")), // Green
		NumberValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("209")), // Orange
		BoolValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")), // Pink
		NullValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")). // Gray
			Italic(true),
		TimestampValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color("180")), // Tan

		// Structural styles
		ExpandIcon: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		TypeIcon: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		TreeLine: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		ChildCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),

		// Preview pane
		PreviewTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("117")),
		PreviewPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),
		PreviewBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),

		// Search styles
		SearchPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")),
		SearchInput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")),
		MatchCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		MatchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("227")).
			Foreground(lipgloss.Color("0")),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		StatusMode: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1),
		StatusInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
	}
}

// GetValueStyle returns the appropriate style for a scalar value
func (s *Styles) GetValueStyle(scalarType int) lipgloss.Style {
	switch scalarType {
	case 0: // ScalarString
		return s.StringValue
	case 1, 2: // ScalarInt, ScalarFloat
		return s.NumberValue
	case 3: // ScalarBool
		return s.BoolValue
	case 4: // ScalarNull
		return s.NullValue
	case 5: // ScalarTimestamp
		return s.TimestampValue
	default:
		return s.StringValue
	}
}
