package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectInfo struct {
	GodotVersion string   `json:"godot_version"`
	ProjectRoot  string   `json:"project_root"`
	Scenes       []string `json:"scenes"`
	Scripts      []string `json:"scripts"`
	Resources    []string `json:"resources"`
	Shaders      []string `json:"shaders"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gdt-ai <inspect|scenes|scripts>")
		os.Exit(1)
	}

	root := os.Getenv("GDT_PROJECT_ROOT")
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		root = cwd
	}

	switch os.Args[1] {
	case "inspect":
		cmdInspect(root)
	case "scenes":
		cmdScenes(root)
	case "scripts":
		cmdScripts(root)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func walkFiles(root string) (scenes, scripts, resources, shaders []string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".godot" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".tscn":
			scenes = append(scenes, rel)
		case ".gd", ".cs":
			scripts = append(scripts, rel)
		case ".tres":
			resources = append(resources, rel)
		case ".gdshader":
			shaders = append(shaders, rel)
		}
		return nil
	})
	return
}

func cmdInspect(root string) {
	scenes, scripts, resources, shaders := walkFiles(root)

	info := ProjectInfo{
		GodotVersion: os.Getenv("GDT_GODOT_VERSION"),
		ProjectRoot:  root,
		Scenes:       nonNil(scenes),
		Scripts:      nonNil(scripts),
		Resources:    nonNil(resources),
		Shaders:      nonNil(shaders),
	}

	enc, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}

func cmdScenes(root string) {
	scenes, _, _, _ := walkFiles(root)
	for _, s := range scenes {
		fmt.Println(s)
	}
}

func cmdScripts(root string) {
	_, scripts, _, _ := walkFiles(root)
	for _, s := range scripts {
		fmt.Println(s)
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
