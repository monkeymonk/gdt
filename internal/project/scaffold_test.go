package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateProject(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "mygame")

	opts := ScaffoldOptions{
		Name:     "mygame",
		Version:  "4.3",
		Renderer: "forward_plus",
		Dir:      projectDir,
	}

	err := Generate(opts)
	if err != nil {
		t.Fatal(err)
	}

	// Check project.godot exists and has correct content
	data, err := os.ReadFile(filepath.Join(projectDir, "project.godot"))
	if err != nil {
		t.Fatal("project.godot should exist")
	}
	content := string(data)
	if !strings.Contains(content, "config/name=\"mygame\"") {
		t.Error("project.godot should contain project name")
	}
	if !strings.Contains(content, "forward_plus") {
		t.Error("project.godot should contain renderer")
	}

	// Check .godot-version
	verData, err := os.ReadFile(filepath.Join(projectDir, ".godot-version"))
	if err != nil {
		t.Fatal(".godot-version should exist")
	}
	if strings.TrimSpace(string(verData)) != "4.3" {
		t.Errorf(".godot-version = %q, want %q", strings.TrimSpace(string(verData)), "4.3")
	}

	// Check .gitignore
	if _, err := os.Stat(filepath.Join(projectDir, ".gitignore")); err != nil {
		t.Error(".gitignore should exist")
	}

	// Check .editorconfig
	if _, err := os.Stat(filepath.Join(projectDir, ".editorconfig")); err != nil {
		t.Error(".editorconfig should exist")
	}
}

func TestGenerateProjectAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "existing")
	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "project.godot"), []byte("exists"), 0644)

	opts := ScaffoldOptions{
		Name:     "existing",
		Version:  "4.3",
		Renderer: "forward_plus",
		Dir:      projectDir,
	}

	err := Generate(opts)
	if err == nil {
		t.Error("should error when project.godot already exists")
	}
}

func TestGenerateRenderers(t *testing.T) {
	renderers := []string{"forward_plus", "mobile", "gl_compatibility"}
	for _, r := range renderers {
		t.Run(r, func(t *testing.T) {
			dir := t.TempDir()
			projectDir := filepath.Join(dir, "game")

			err := Generate(ScaffoldOptions{
				Name:     "game",
				Version:  "4.3",
				Renderer: r,
				Dir:      projectDir,
			})
			if err != nil {
				t.Fatal(err)
			}

			data, _ := os.ReadFile(filepath.Join(projectDir, "project.godot"))
			if !strings.Contains(string(data), r) {
				t.Errorf("project.godot should contain renderer %q", r)
			}
		})
	}
}

func TestGitIgnoreContent(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "game")

	Generate(ScaffoldOptions{Name: "game", Version: "4.3", Renderer: "forward_plus", Dir: projectDir})

	data, _ := os.ReadFile(filepath.Join(projectDir, ".gitignore"))
	content := string(data)
	if !strings.Contains(content, ".godot/") {
		t.Error(".gitignore should contain .godot/")
	}
	if !strings.Contains(content, "*.import") {
		t.Error(".gitignore should contain *.import")
	}
}

func TestGenerateCSharpProject(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "mygame")

	opts := ScaffoldOptions{
		Name:     "mygame",
		Version:  "4.3",
		Renderer: "forward_plus",
		Dir:      projectDir,
		CSharp:   true,
	}

	err := Generate(opts)
	if err != nil {
		t.Fatal(err)
	}

	// .godot-version should have -mono suffix
	verData, _ := os.ReadFile(filepath.Join(projectDir, ".godot-version"))
	if strings.TrimSpace(string(verData)) != "4.3-mono" {
		t.Errorf(".godot-version = %q, want %q", strings.TrimSpace(string(verData)), "4.3-mono")
	}

	// .csproj should exist
	csprojPath := filepath.Join(projectDir, "mygame.csproj")
	csprojData, err := os.ReadFile(csprojPath)
	if err != nil {
		t.Fatal("mygame.csproj should exist")
	}
	if !strings.Contains(string(csprojData), "Godot.NET.Sdk/4.3") {
		t.Error(".csproj should reference Godot.NET.Sdk with version")
	}

	// .sln should exist
	slnPath := filepath.Join(projectDir, "mygame.sln")
	slnData, err := os.ReadFile(slnPath)
	if err != nil {
		t.Fatal("mygame.sln should exist")
	}
	if !strings.Contains(string(slnData), "mygame.csproj") {
		t.Error(".sln should reference .csproj")
	}

	// .gitignore should have C# entries
	gitignoreData, _ := os.ReadFile(filepath.Join(projectDir, ".gitignore"))
	if !strings.Contains(string(gitignoreData), "bin/") {
		t.Error(".gitignore should contain bin/ for C# project")
	}
	if !strings.Contains(string(gitignoreData), "obj/") {
		t.Error(".gitignore should contain obj/ for C# project")
	}
}

func TestGenerateCSharpAlreadyMono(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "mygame")

	err := Generate(ScaffoldOptions{
		Name:     "mygame",
		Version:  "4.3-mono",
		Renderer: "forward_plus",
		Dir:      projectDir,
		CSharp:   true,
	})
	if err != nil {
		t.Fatal(err)
	}

	verData, _ := os.ReadFile(filepath.Join(projectDir, ".godot-version"))
	if strings.TrimSpace(string(verData)) != "4.3-mono" {
		t.Errorf(".godot-version = %q, want %q (should not double-suffix)", strings.TrimSpace(string(verData)), "4.3-mono")
	}
}

func TestGenerateFromTemplate2D(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	err := GenerateFromTemplate("2d", dir, "MyGame", "4.3")
	if err != nil {
		t.Fatal(err)
	}

	// project.godot should exist with name substituted
	data, err := os.ReadFile(filepath.Join(dir, "project.godot"))
	if err != nil {
		t.Fatal("project.godot should exist")
	}
	content := string(data)
	if !strings.Contains(content, `config/name="MyGame"`) {
		t.Error("project.godot should contain substituted project name")
	}
	if !strings.Contains(content, "run/main_scene") {
		t.Error("project.godot should contain main_scene")
	}

	// main.tscn should exist with Node2D
	sceneData, err := os.ReadFile(filepath.Join(dir, "main.tscn"))
	if err != nil {
		t.Fatal("main.tscn should exist")
	}
	if !strings.Contains(string(sceneData), "Node2D") {
		t.Error("main.tscn should contain Node2D")
	}

	// .godot-version should exist
	verData, err := os.ReadFile(filepath.Join(dir, ".godot-version"))
	if err != nil {
		t.Fatal(".godot-version should exist")
	}
	if strings.TrimSpace(string(verData)) != "4.3" {
		t.Errorf(".godot-version = %q, want %q", strings.TrimSpace(string(verData)), "4.3")
	}
}

func TestGenerateFromTemplate3D(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	err := GenerateFromTemplate("3d", dir, "My3DGame", "4.4")
	if err != nil {
		t.Fatal(err)
	}

	sceneData, err := os.ReadFile(filepath.Join(dir, "main.tscn"))
	if err != nil {
		t.Fatal("main.tscn should exist")
	}
	if !strings.Contains(string(sceneData), "Node3D") {
		t.Error("main.tscn should contain Node3D")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "project.godot"))
	if !strings.Contains(string(data), `config/name="My3DGame"`) {
		t.Error("project.godot should contain substituted project name")
	}
}

func TestGenerateFromTemplateUnknown(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	err := GenerateFromTemplate("fps", dir, "game", "4.3")
	if err == nil {
		t.Error("should error for unknown template")
	}
}

func TestGenerateFromTemplateAlreadyExists(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "project.godot"), []byte("exists"), 0644)

	err := GenerateFromTemplate("2d", dir, "game", "4.3")
	if err == nil {
		t.Error("should error when project.godot already exists")
	}
}

func TestCloneTemplate(t *testing.T) {
	// Create a fake "remote" repo
	repoDir := t.TempDir()
	exec.Command("git", "init", repoDir).Run()
	exec.Command("git", "-C", repoDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", repoDir, "config", "user.name", "test").Run()
	os.WriteFile(filepath.Join(repoDir, "project.godot"), []byte("[application]\n"), 0644)
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Template\n"), 0644)
	exec.Command("git", "-C", repoDir, "add", "-A").Run()
	exec.Command("git", "-C", repoDir, "commit", "-m", "init").Run()

	destDir := filepath.Join(t.TempDir(), "mygame")

	err := CloneTemplate(repoDir, destDir, "4.3")
	if err != nil {
		t.Fatal(err)
	}

	// Template files should exist
	if _, err := os.Stat(filepath.Join(destDir, "README.md")); err != nil {
		t.Error("README.md should exist from template")
	}
	if _, err := os.Stat(filepath.Join(destDir, "project.godot")); err != nil {
		t.Error("project.godot should exist from template")
	}

	// .godot-version should be overwritten
	data, _ := os.ReadFile(filepath.Join(destDir, ".godot-version"))
	if strings.TrimSpace(string(data)) != "4.3" {
		t.Errorf(".godot-version = %q, want %q", strings.TrimSpace(string(data)), "4.3")
	}
}

func TestCloneTemplateDestExists(t *testing.T) {
	destDir := t.TempDir()
	os.WriteFile(filepath.Join(destDir, "file.txt"), []byte("exists"), 0644)

	entries, _ := os.ReadDir(destDir)
	if len(entries) == 0 {
		t.Skip("dir should not be empty")
	}

	err := CloneTemplate("https://example.com/repo", destDir, "4.3")
	if err == nil {
		t.Error("should error when destination is not empty")
	}
}
