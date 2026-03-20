package editor

func neovimSnippet() string {
	return `-- Add to your Neovim LSP configuration (e.g. init.lua)
local lspconfig = require('lspconfig')

lspconfig.gdscript.setup{
  cmd = { 'gdt', 'lsp' },
  filetypes = { 'gdscript', 'gd' },
  root_dir = lspconfig.util.root_pattern('project.godot'),
}`
}
