package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectRootInCurrentDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "project.godot"), []byte("[application]\n"), 0644)

	root, err := DetectRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Errorf("root = %q, want %q", root, dir)
	}
}

func TestDetectRootInParent(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "scenes", "level1")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, "project.godot"), []byte("[application]\n"), 0644)

	root, err := DetectRoot(sub)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Errorf("root = %q, want %q", root, dir)
	}
}

func TestDetectRootNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := DetectRoot(dir)
	if err == nil {
		t.Error("should error when no project.godot found")
	}
}

func TestHasCSharp(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "project.godot"), []byte("[application]\n"), 0644)
	os.WriteFile(filepath.Join(dir, "Player.cs"), []byte("class Player {}"), 0644)

	has, err := HasCSharp(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("should detect C# files")
	}
}

func TestHasCSharpNone(t *testing.T) {
	dir := t.TempDir()

	has, err := HasCSharp(dir)
	if err != nil {
		t.Fatal(err)
	}
	if has {
		t.Error("should not detect C# when no .cs files")
	}
}

func TestHasCSharpCsproj(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Game.csproj"), []byte("<Project/>"), 0644)

	has, err := HasCSharp(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Error("should detect .csproj")
	}
}
