# yamlist

A terminal-based YAML navigator with tree view, fuzzy search, and Neovim integration.

## Features

- **Full-width tree view** - Navigate large YAML files with an expandable tree structure
- **Fuzzy search** - Find nodes quickly with live filtering and match highlighting
- **Neovim integration** - Cursor sync: navigate the tree and your editor follows
- **Syntax highlighting** - Color-coded values by type (strings, numbers, booleans, etc.)
- **Vim-style navigation** - Familiar keybindings for efficient browsing

## Installation

### Go Binary

**Option 1: go install (recommended)**
```bash
go install github.com/uznog/yamlist/cmd/yamlist@latest
```
Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`.

**Option 2: Build from source**
```bash
git clone https://github.com/uznog/yamlist.git
cd yamlist
go build -o yamlist ./cmd/yamlist
# Move to a directory in your PATH
mv yamlist ~/.local/bin/
```

**Verify installation:**
```bash
yamlist --version
```

### Neovim Plugin

**Prerequisites:**
- Neovim 0.7+ (for `vim.loop` / libuv support)
- `yamlist` binary installed and in `PATH`

**Using lazy.nvim:**
```lua
{
  "uznog/yamlist",
  config = function()
    require("yamlist").setup({
      -- Optional: path to yamlist binary (if not in PATH)
      -- executable = "/path/to/yamlist",
    })
  end,
  ft = { "yaml", "yml" },  -- Lazy load for YAML files
}
```

**Using packer.nvim:**
```lua
use {
  "uznog/yamlist",
  config = function()
    require("yamlist").setup()
  end,
  ft = { "yaml", "yml" },
}
```

**Using vim-plug:**
```vim
Plug 'uznog/yamlist'

" In your init.vim or init.lua, after plug#end():
lua require("yamlist").setup()
```

**Manual installation:**
```bash
# Clone to your Neovim packages directory
git clone https://github.com/uznog/yamlist.git \
  ~/.local/share/nvim/site/pack/plugins/start/yamlist

# Add to your init.lua:
# require("yamlist").setup()
```

## Usage

### Standalone

```bash
yamlist <file.yaml>
```

**Options:**
```
  --no-icons           Use ASCII characters instead of Nerd Font icons
  --theme <theme>      Color theme: auto, dark, mono (default: auto)
  --nvim-socket <path> Unix socket path for Neovim cursor sync
  --version            Show version and exit
```

### Neovim

```vim
:YAMList           " Open yamlist for current YAML buffer
:YAMList path.yaml " Open yamlist for specific file
```

**Recommended keymap (add to your config):**
```lua
vim.keymap.set("n", "<leader>fy", "<cmd>YAMList<cr>", { desc = "YAML navigator" })
```

## Keybindings

| Key | Mode | Action |
|-----|------|--------|
| `j` / `k` | Tree | Move down / up |
| `h` | Tree | Collapse node or go to parent |
| `l` | Tree | Expand node or go to first child |
| `space` / `enter` | Tree | Toggle expand/collapse |
| `z` | Tree | Collapse all |
| `Z` | Tree | Expand all |
| `g` / `G` | Tree | Go to top / bottom |
| `Ctrl+d` / `Ctrl+u` | Tree | Page down / up |
| `/` | Tree | Enter search mode |
| `n` / `N` | Tree/Search | Next / previous match |
| `esc` | Tree | Clear search highlighting |
| `q` | Tree | Quit |
| (typing) | Search | Update search query, grey out non-matches |
| `enter` | Search | Confirm search, return to tree mode |
| `esc` | Search | Clear search and highlighting |

## Neovim Integration

When opened from Neovim using `:YAMList`:

1. A floating terminal window opens with yamlist
2. Navigating the tree automatically moves the cursor in your edit buffer to the corresponding YAML line
3. Search and use `n`/`N` to jump between matches - your editor cursor follows
4. Press `q` to close - the floating window and temporary files are cleaned up automatically

This provides a powerful way to navigate complex YAML files while keeping your place in the editor.

## Themes

- `auto` (default) - Colorful theme optimized for dark terminals
- `dark` - Same as auto, explicitly dark-optimized
- `mono` - Minimal monochrome for reduced visual noise

## Configuration

### Neovim Plugin

```lua
require("yamlist").setup({
  executable = "yamlist",  -- Path to yamlist binary
  no_icons = false,        -- Use ASCII instead of Nerd Font icons
})
```

## License

MIT
