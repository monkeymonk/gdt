package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const desktopFileName = "gdt-godot.desktop"
const iconFileName = "gdt-godot.svg"

const godotIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
  <path fill="#478cbf" d="M63.8 17.3c-18.6 0-34.8 10-43.6 24.9v30.1c0 2.3 1.3 4.4 3.3 5.5l36.7 21.2c2.1 1.2 4.6 1.2 6.7 0l36.7-21.2c2-1.2 3.3-3.3 3.3-5.5V42.2c-8.9-14.9-25.1-24.9-43.7-24.9z"/>
  <circle fill="#fff" cx="46" cy="52" r="8"/>
  <circle fill="#414042" cx="46" cy="52" r="4"/>
  <circle fill="#fff" cx="82" cy="52" r="8"/>
  <circle fill="#414042" cx="82" cy="52" r="4"/>
  <path fill="#fff" d="M54 72h20v6H54z" rx="3"/>
</svg>`

// installDesktop creates the .desktop file and icon for Godot managed by gdt.
// Best-effort: errors are silently ignored.
// Linux only — no-op on other platforms.
func (s *Service) installDesktop() {
	if runtime.GOOS != "linux" {
		return
	}

	gdtBin := gdtBinaryPath()

	// Install icon
	iconPath := s.installIcon()

	// Create .desktop file
	dir := applicationsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	content := "[Desktop Entry]\n" +
		"Name=Godot Engine (gdt)\n" +
		"Comment=Godot game engine managed by gdt\n" +
		"Exec=" + gdtBin + " run %f\n" +
		"Icon=" + iconPath + "\n" +
		"Type=Application\n" +
		"Categories=Development;IDE;Game;\n" +
		"MimeType=application/x-godot-project;\n" +
		"Keywords=godot;game;engine;gamedev;\n" +
		"StartupWMClass=Godot\n"

	path := filepath.Join(dir, desktopFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return
	}

	// Update desktop database if available
	if dbPath, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(dbPath, dir).Run()
	}
}

// removeDesktop removes the .desktop file and icon.
// Best-effort: errors are silently ignored.
// Linux only — no-op on other platforms.
func (s *Service) removeDesktop() {
	if runtime.GOOS != "linux" {
		return
	}

	// Remove .desktop file
	path := filepath.Join(applicationsDir(), desktopFileName)
	_ = os.Remove(path)

	// Remove icon
	iconPath := filepath.Join(iconShareDir(), iconFileName)
	_ = os.Remove(iconPath)

	// Update desktop database if available
	dir := applicationsDir()
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		if dbPath, err := exec.LookPath("update-desktop-database"); err == nil {
			_ = exec.Command(dbPath, dir).Run()
		}
	}
}

// isDesktopInstalled reports whether the .desktop file exists.
func (s *Service) isDesktopInstalled() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err := os.Stat(filepath.Join(applicationsDir(), desktopFileName))
	return err == nil
}

// installIcon writes the Godot icon to the appropriate location.
// Returns the icon path. Best-effort: falls back to gdt home dir.
func (s *Service) installIcon() string {
	iconDir := iconShareDir()
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		iconDir = s.Home
	}

	iconPath := filepath.Join(iconDir, iconFileName)
	if err := os.WriteFile(iconPath, []byte(godotIconSVG), 0644); err != nil {
		return ""
	}

	return iconPath
}

func gdtBinaryPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "gdt"
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe
	}
	return resolved
}

func applicationsDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "applications")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "applications")
}

func iconShareDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "icons", "hicolor", "scalable", "apps")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "icons", "hicolor", "scalable", "apps")
}
