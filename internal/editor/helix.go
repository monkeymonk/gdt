package editor

func helixSnippet() string {
	return `# Add to ~/.config/helix/languages.toml

[[language]]
name = "gdscript"
language-servers = ["gdscript"]

[language-server.gdscript]
command = "gdt"
args = ["lsp"]`
}
