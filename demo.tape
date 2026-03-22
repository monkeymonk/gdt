# ===============================
#  Professional CLI Showcase
#  Author: monkeymonk
#  Project: gdt
# ===============================

# -------- Output --------
Output demo.gif

# -------- Terminal Look --------
Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 750
Set Theme "Catppuccin Mocha"
Set Padding 20
Set CursorBlink true

# Optional: for slower, more polished feel
Set TypingSpeed 80ms
Set Framerate 60

# Hidden setup: clean environment + prompt
Hide
Type `export PS1='$ '`
Enter
Type "export XDG_CONFIG_HOME=/tmp/alx-demo-vhs"
Enter
Type "rm -rf /tmp/alx-demo-vhs"
Enter
Type "export PATH=$PWD/bin:/usr/local/bin:/usr/bin:/bin"
Enter
Type "clear"
Enter
Sleep 500ms
Show

# -------- Intro Section --------
Sleep 500ms

# Big clean intro
Type `# gdt - A cross-platform CLI to manage Godot Engine installations`
Enter
Sleep 800ms

# -------- Basic Usage --------
Type "gdt --help"
Enter
Sleep 4s

# List remote versions
Type "gdt ls-remote"
Enter
Sleep 3s

# Install a Godot version (hide download, show result)
Type "gdt install 4.3"
Enter
Sleep 1s
Hide
Sleep 120s
Show
Sleep 2s

# List installed versions
Type "gdt list"
Enter
Sleep 2s

# Create a new project
Type "gdt new /tmp/gdt-demo/mygame --version 4.3 --renderer forward_plus"
Enter
Sleep 5s

# Inspect project
Type "ls -la /tmp/gdt-demo/mygame/"
Enter
Sleep 2s

Type "cat /tmp/gdt-demo/mygame/.godot-version"
Enter
Sleep 2s

# Work from inside the project
Type "cd /tmp/gdt-demo/mygame"
Enter

# Pin a local version
Type "gdt local 4.3"
Enter
Sleep 2s

# Run diagnostics
Type "gdt doctor"
Enter
Sleep 3s

# Install export templates (hide download, show result)
Type "gdt templates install 4.3"
Enter
Sleep 1s
Hide
Sleep 120s
Show
Sleep 2s

# Generate CI pipeline
Type "gdt ci setup --provider github"
Enter
Sleep 3s

# Plugin management
Type "gdt plugin list"
Enter
Sleep 3s

# -------- Outro --------
Type `# Documentation: https://github.com/monkeymonk/gdt`
Enter
Sleep 1500ms
