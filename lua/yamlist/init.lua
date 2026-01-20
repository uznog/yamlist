local M = {}

M.config = {
  executable = "yamlist",
  no_icons = false,
}

function M.setup(opts)
  M.config = vim.tbl_deep_extend("force", M.config, opts or {})

  vim.api.nvim_create_user_command("YAMList", function(args)
    M.open(args.args ~= "" and args.args or nil)
  end, { nargs = "?", complete = "file" })
end

-- Create a centered floating window
local function create_float_win()
  local width = math.floor(vim.o.columns * 0.8)
  local height = math.floor(vim.o.lines * 0.8)
  local row = math.floor((vim.o.lines - height) / 2)
  local col = math.floor((vim.o.columns - width) / 2)

  local buf = vim.api.nvim_create_buf(false, true)
  local win = vim.api.nvim_open_win(buf, true, {
    relative = "editor",
    width = width,
    height = height,
    row = row,
    col = col,
    style = "minimal",
    border = "rounded",
  })

  return buf, win
end

-- Write buffer content to a temporary file
local function write_buffer_to_temp(bufnr)
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local tmpfile = vim.fn.tempname() .. ".yaml"
  vim.fn.writefile(lines, tmpfile)
  return tmpfile
end

-- Start a Unix socket server for cursor sync
local function start_socket_server(edit_win)
  local uv = vim.loop
  local socket_path = vim.fn.tempname() .. ".sock"
  local server = uv.new_pipe(false)

  server:bind(socket_path)
  server:listen(128, function(err)
    if err then
      return
    end
    local client = uv.new_pipe(false)
    server:accept(client)

    local buffer = ""
    client:read_start(function(read_err, data)
      if read_err or not data then
        client:close()
        return
      end
      buffer = buffer .. data

      -- Parse JSONL messages
      while true do
        local newline_pos = buffer:find("\n")
        if not newline_pos then
          break
        end

        local line = buffer:sub(1, newline_pos - 1)
        buffer = buffer:sub(newline_pos + 1)

        -- Parse JSON and handle cursor message
        local ok, msg = pcall(vim.json.decode, line)
        if ok and msg.op == "cursor" and msg.line then
          vim.schedule(function()
            if vim.api.nvim_win_is_valid(edit_win) then
              -- Clamp line number to valid range
              local line_count = vim.api.nvim_buf_line_count(
                vim.api.nvim_win_get_buf(edit_win)
              )
              local target_line = math.min(msg.line, line_count)
              target_line = math.max(1, target_line)
              pcall(vim.api.nvim_win_set_cursor, edit_win, { target_line, 0 })
            end
          end)
        end
      end
    end)
  end)

  return server, socket_path
end

function M.open(file)
  local edit_win = vim.api.nvim_get_current_win()
  local edit_buf = vim.api.nvim_get_current_buf()

  -- Determine source file
  local source_file
  local tmpfile = nil

  if file then
    -- Explicit file argument
    if vim.fn.filereadable(file) == 0 then
      vim.notify("YAMList: File not found: " .. file, vim.log.levels.ERROR)
      return
    end
    source_file = file
  else
    -- Use current buffer content via temp file
    local bufname = vim.api.nvim_buf_get_name(edit_buf)
    local ext = vim.fn.fnamemodify(bufname, ":e"):lower()
    if ext ~= "yaml" and ext ~= "yml" and bufname ~= "" then
      vim.notify("YAMList: Not a YAML file: " .. bufname, vim.log.levels.ERROR)
      return
    end

    -- Write buffer content to temp file
    tmpfile = write_buffer_to_temp(edit_buf)
    source_file = tmpfile
  end

  -- Start socket server for cursor sync
  local server, socket_path = start_socket_server(edit_win)

  -- Create floating terminal window
  local term_buf, term_win = create_float_win()

  -- Build command
  local cmd = { M.config.executable }
  if M.config.no_icons then
    table.insert(cmd, "--no-icons")
  end
  table.insert(cmd, "--nvim-socket=" .. socket_path)
  table.insert(cmd, source_file)

  -- Run yamlist in terminal
  vim.fn.termopen(cmd, {
    on_exit = function()
      -- Cleanup
      if server then
        server:close()
      end
      vim.fn.delete(socket_path)
      if tmpfile then
        vim.fn.delete(tmpfile)
      end
      if vim.api.nvim_win_is_valid(term_win) then
        vim.api.nvim_win_close(term_win, true)
      end
      if vim.api.nvim_buf_is_valid(term_buf) then
        vim.api.nvim_buf_delete(term_buf, { force = true })
      end
    end,
  })

  -- Enter insert mode to interact with terminal
  vim.cmd("startinsert")
end

return M
