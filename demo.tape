# VHS tape file - https://github.com/charmbracelet/vhs
# Run: vhs demo.tape

Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

Type "gdt --version"
Enter
Sleep 1s

Type "# Install a Godot version"
Enter
Sleep 500ms
Type "gdt install 4.3"
Enter
Sleep 3s

Type "# Set as global default"
Enter
Sleep 500ms
Type "gdt use 4.3"
Enter
Sleep 1s

Type "# Create a new project"
Enter
Sleep 500ms
Type "gdt new mygame --version 4.3 --renderer forward_plus"
Enter
Sleep 2s

Type "# Check what's inside"
Enter
Sleep 500ms
Type "ls -la mygame/"
Enter
Sleep 2s

Type "# List installed versions"
Enter
Sleep 500ms
Type "gdt list"
Enter
Sleep 1s

Type "# Run diagnostics"
Enter
Sleep 500ms
Type "gdt doctor"
Enter
Sleep 2s

Sleep 3s
