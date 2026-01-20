package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/uznog/yamlist/internal/model"
	"github.com/uznog/yamlist/internal/nvim"
	"github.com/uznog/yamlist/internal/render"
	"github.com/uznog/yamlist/internal/yamlparse"
)

// Mode represents the current UI mode
type Mode int

const (
	TreeMode Mode = iota
	SearchMode
)

// ViewMode represents the current view mode (tree or flat)
type ViewMode int

const (
	TreeView ViewMode = iota
	FlatView
)

// Config holds configuration options
type Config struct {
	UseIcons        bool
	MaxPreviewLines int
	Theme           string // "auto", "dark", "mono"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		UseIcons:        true,
		MaxPreviewLines: 200,
		Theme:           "auto",
	}
}

// Model is the main Bubble Tea model
type Model struct {
	// Document is the parsed YAML document
	Document *yamlparse.Document

	// TreeState holds the tree view state
	TreeState *model.TreeState

	// Mode is the current UI mode
	Mode Mode

	// ViewMode is the current view mode (tree or flat)
	ViewMode ViewMode

	// Search state
	SearchInput   textinput.Model
	SearchMatches []*model.PathEntry
	SearchIndex   int
	SearchActive  bool // True when search results should be highlighted/dimmed

	// Rendering
	RowRenderer     *render.RowRenderer
	PreviewRenderer *render.PreviewRenderer
	Icons           *render.IconSet
	Styles          *render.Styles

	// Layout
	Width         int
	Height        int
	TreeWidth     int
	PreviewWidth  int

	// Config
	Config *Config

	// Error message (if any)
	Error string

	// NvimClient for cursor sync with Neovim (nil if standalone)
	NvimClient *nvim.Client
}

// NewModel creates a new TUI model
func NewModel(doc *yamlparse.Document, config *Config, nvimClient *nvim.Client) *Model {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize icons
	var icons *render.IconSet
	if config.UseIcons {
		icons = render.NerdFontIcons()
	} else {
		icons = render.ASCIIIcons()
	}

	styles := render.StylesForTheme(render.Theme(config.Theme))

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 256

	// Create tree state
	treeState := model.NewTreeState(doc.Root)

	// Expand all nodes by default
	treeState.ExpandAll()

	m := &Model{
		Document:        doc,
		TreeState:       treeState,
		Mode:            TreeMode,
		ViewMode:        TreeView,
		SearchInput:     ti,
		SearchMatches:   make([]*model.PathEntry, 0),
		SearchIndex:     0,
		RowRenderer:     render.NewRowRenderer(icons, styles),
		PreviewRenderer: render.NewPreviewRenderer(styles, config.MaxPreviewLines),
		Icons:           icons,
		Styles:          styles,
		Config:          config,
		NvimClient:      nvimClient,
	}

	// Initialize visible rows
	m.computeVisibleRows()

	return m
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateLayout()
		return m, nil
	}

	// Update search input if in search mode
	if m.Mode == SearchMode {
		var cmd tea.Cmd
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		m.updateSearchMatches()
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model
func (m *Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	return m.renderLayout()
}

// handleKeyMsg handles keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "q", "ctrl+c":
		if m.Mode == TreeMode {
			return m, tea.Quit
		}
	}

	// Mode-specific handling
	if m.Mode == SearchMode {
		return m.handleSearchKey(msg)
	}

	return m.handleTreeKey(msg)
}

// SetError sets an error message
func (m *Model) SetError(err string) {
	m.Error = err
}

// ClearError clears any error message
func (m *Model) ClearError() {
	m.Error = ""
}
