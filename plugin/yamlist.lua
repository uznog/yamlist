if vim.g.loaded_yamlist then
  return
end
vim.g.loaded_yamlist = true

require("yamlist").setup()
