package project

import (
	"crypto/rand"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates
var embeddedTemplates embed.FS

// AvailableTemplates returns the list of built-in scaffold template names.
func AvailableTemplates() []string {
	return []string{"2d", "3d"}
}

// GenerateFromTemplate creates a project from an embedded template (2d or 3d).
// Template files support Go text/template placeholders (e.g. {{.Name}}).
func GenerateFromTemplate(templateName string, destDir string, name string, version string) error {
	base := "templates/" + templateName
	entries, err := fs.ReadDir(embeddedTemplates, base)
	if err != nil {
		return fmt.Errorf("unknown template: %s (available: 2d, 3d)", templateName)
	}

	if _, err := os.Stat(filepath.Join(destDir, "project.godot")); err == nil {
		return fmt.Errorf("project.godot already exists in %s", destDir)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	data := map[string]string{
		"Name": name,
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, readErr := fs.ReadFile(embeddedTemplates, base+"/"+e.Name())
		if readErr != nil {
			return readErr
		}

		tmpl, parseErr := template.New(e.Name()).Parse(string(raw))
		if parseErr != nil {
			// Not a valid template; write raw content
			if writeErr := os.WriteFile(filepath.Join(destDir, e.Name()), raw, 0o644); writeErr != nil {
				return writeErr
			}
			continue
		}

		var buf strings.Builder
		if execErr := tmpl.Execute(&buf, data); execErr != nil {
			return execErr
		}

		if writeErr := os.WriteFile(filepath.Join(destDir, e.Name()), []byte(buf.String()), 0o644); writeErr != nil {
			return writeErr
		}
	}

	return os.WriteFile(filepath.Join(destDir, ".godot-version"), []byte(version+"\n"), 0o644)
}

type ScaffoldOptions struct {
	Name     string
	Version  string
	Renderer string
	Dir      string
	CSharp   bool
}

func Generate(opts ScaffoldOptions) error {
	if _, err := os.Stat(filepath.Join(opts.Dir, "project.godot")); err == nil {
		return fmt.Errorf("project.godot already exists in %s", opts.Dir)
	}

	if err := os.MkdirAll(opts.Dir, 0755); err != nil {
		return err
	}

	if err := writeProjectGodot(opts); err != nil {
		return err
	}
	if err := writeGodotVersion(opts); err != nil {
		return err
	}
	if err := writeGitIgnore(opts); err != nil {
		return err
	}
	if err := writeEditorConfig(opts); err != nil {
		return err
	}
	if opts.CSharp {
		if err := writeCSharpFiles(opts); err != nil {
			return err
		}
	}

	return nil
}

// CopyTemplate copies a template directory to destDir, processing Go template placeholders.
func CopyTemplate(srcDir string, destDir string, name string, version string) error {
	if _, err := os.Stat(filepath.Join(destDir, "project.godot")); err == nil {
		return fmt.Errorf("project.godot already exists in %s", destDir)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	data := map[string]string{
		"Name": name,
	}

	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(srcDir, path)
		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		tmpl, parseErr := template.New(d.Name()).Parse(string(raw))
		if parseErr != nil {
			return os.WriteFile(destPath, raw, 0o644)
		}

		var buf strings.Builder
		if execErr := tmpl.Execute(&buf, data); execErr != nil {
			return execErr
		}

		return os.WriteFile(destPath, []byte(buf.String()), 0o644)
	})
}

func CloneTemplate(repoURL string, destDir string, version string) error {
	if !strings.HasPrefix(repoURL, "http") && !strings.HasPrefix(repoURL, "file://") && !strings.HasPrefix(repoURL, "/") {
		repoURL = "https://github.com/" + repoURL
	}

	if entries, err := os.ReadDir(destDir); err == nil && len(entries) > 0 {
		return fmt.Errorf("destination %s is not empty", destDir)
	}

	gitCmd := exec.Command("git", "clone", "--depth", "1", repoURL, destDir)
	gitCmd.Stdout = os.Stderr
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}

	os.RemoveAll(filepath.Join(destDir, ".git"))

	return writeGodotVersion(ScaffoldOptions{Version: version, Dir: destDir})
}

func writeProjectGodot(opts ScaffoldOptions) error {
	content := fmt.Sprintf(`; Engine configuration file.
; It's best edited using the editor UI and not directly,
; since the parameters that go here are not all obvious.
;
; Format:
;   [section] ; section goes between []
;   param=value ; assign values to parameters

config_version=5

[application]

config/name="%s"
config/features=PackedStringArray("%s")

[rendering]

renderer/rendering_method="%s"
`, opts.Name, opts.Renderer, opts.Renderer)

	return os.WriteFile(filepath.Join(opts.Dir, "project.godot"), []byte(content), 0644)
}

func writeGodotVersion(opts ScaffoldOptions) error {
	version := opts.Version
	if opts.CSharp && !strings.HasSuffix(version, "-mono") {
		version += "-mono"
	}
	return os.WriteFile(filepath.Join(opts.Dir, ".godot-version"), []byte(version+"\n"), 0644)
}

func writeGitIgnore(opts ScaffoldOptions) error {
	content := `# Godot
.godot/
*.import
export_presets.cfg

# Mono
.mono/
data_*/
mono_crash.*.txt

# Builds
/dist/
/build/

# OS
.DS_Store
Thumbs.db
`
	if opts.CSharp {
		content += `
# .NET
bin/
obj/
*.user
*.suo
*.nupkg
*.snupkg
`
	}
	return os.WriteFile(filepath.Join(opts.Dir, ".gitignore"), []byte(content), 0644)
}

func writeEditorConfig(opts ScaffoldOptions) error {
	content := `root = true

[*]
indent_style = tab
indent_size = 4
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.{gd,tscn,tres,godot,cfg}]
indent_style = tab
indent_size = 4

[*.{json,yml,yaml}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
`
	return os.WriteFile(filepath.Join(opts.Dir, ".editorconfig"), []byte(content), 0644)
}

func writeCSharpFiles(opts ScaffoldOptions) error {
	projectGUID := newGUID()
	slnGUID := newGUID()

	csproj := fmt.Sprintf(`<Project Sdk="Godot.NET.Sdk/%s">
  <PropertyGroup>
    <TargetFramework>net6.0</TargetFramework>
    <EnableDynamicLoading>true</EnableDynamicLoading>
  </PropertyGroup>
</Project>
`, opts.Version)

	if err := os.WriteFile(filepath.Join(opts.Dir, opts.Name+".csproj"), []byte(csproj), 0644); err != nil {
		return err
	}

	sln := fmt.Sprintf(`Microsoft Visual Studio Solution File, Format Version 12.00
# Visual Studio 2012
Project("{%s}") = "%s", "%s.csproj", "{%s}"
EndProject
Global
	GlobalSection(SolutionConfigurationPlatforms) = preSolution
		Debug|Any CPU = Debug|Any CPU
		ExportDebug|Any CPU = ExportDebug|Any CPU
		ExportRelease|Any CPU = ExportRelease|Any CPU
	EndGlobalSection
	GlobalSection(ProjectConfigurationPlatforms) = postSolution
		{%s}.Debug|Any CPU.ActiveCfg = Debug|Any CPU
		{%s}.Debug|Any CPU.Build.0 = Debug|Any CPU
		{%s}.ExportDebug|Any CPU.ActiveCfg = ExportDebug|Any CPU
		{%s}.ExportDebug|Any CPU.Build.0 = ExportDebug|Any CPU
		{%s}.ExportRelease|Any CPU.ActiveCfg = ExportRelease|Any CPU
		{%s}.ExportRelease|Any CPU.Build.0 = ExportRelease|Any CPU
	EndGlobalSection
EndGlobal
`, slnGUID, opts.Name, opts.Name, projectGUID,
		projectGUID, projectGUID, projectGUID, projectGUID, projectGUID, projectGUID)

	return os.WriteFile(filepath.Join(opts.Dir, opts.Name+".sln"), []byte(sln), 0644)
}

func newGUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08X-%04X-%04X-%04X-%012X",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
