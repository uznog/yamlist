package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/uznog/yamlist/internal/nvim"
	"github.com/uznog/yamlist/internal/tui"
	"github.com/uznog/yamlist/internal/yamlparse"
)

var (
	version = "dev"
)

func main() {
	// Command line flags
	noIcons := flag.Bool("no-icons", false, "Use ASCII characters instead of Nerd Font icons")
	maxPreviewLines := flag.Int("max-preview-lines", 200, "Maximum lines to show in preview pane")
	theme := flag.String("theme", "auto", "Color theme: auto, dark, mono")
	nvimSocket := flag.String("nvim-socket", "", "Unix socket path for Neovim cursor sync")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("yamlist %s\n", version)
		os.Exit(0)
	}

	// Validate theme
	validThemes := map[string]bool{"auto": true, "dark": true, "mono": true}
	if !validThemes[*theme] {
		fmt.Fprintf(os.Stderr, "Error: invalid theme %q (use: auto, dark, mono)\n", *theme)
		os.Exit(1)
	}

	// Get file path
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: yamlist [options] <file.yaml>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	filePath := args[0]

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	// Parse YAML file
	doc, err := yamlparse.ParseFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Create config
	config := &tui.Config{
		UseIcons:        !*noIcons,
		MaxPreviewLines: *maxPreviewLines,
		Theme:           *theme,
	}

	// Create Neovim client if socket path provided
	var nvimClient *nvim.Client
	if *nvimSocket != "" {
		client, err := nvim.NewClient(*nvimSocket)
		if err != nil {
			// Log warning but continue - standalone mode
			fmt.Fprintf(os.Stderr, "Warning: could not connect to Neovim socket: %v\n", err)
		} else {
			nvimClient = client
			defer nvimClient.Close()
		}
	}

	// Create and run TUI
	model := tui.NewModel(doc, config, nvimClient)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
