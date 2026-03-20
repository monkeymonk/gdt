package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var assetExts = map[string]bool{
	".png": true, ".jpg": true, ".svg": true,
	".wav": true, ".ogg": true, ".mp3": true,
	".glb": true, ".gltf": true, ".obj": true, ".fbx": true,
	".tres": true, ".tscn": true,
}

type fileEntry struct {
	path string
	size int64
}

func humanSize(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func severity(size int64) string {
	switch {
	case size >= 10<<20:
		return "FAIL"
	case size >= 1<<20:
		return "WARN"
	default:
		return "OK"
	}
}

func projectRoot() string {
	root := os.Getenv("GDT_PROJECT_ROOT")
	if root != "" {
		return root
	}
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return dir
}

func collectAssets(root string) ([]fileEntry, error) {
	var entries []fileEntry
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".godot" || base == ".git" || base == ".import" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if assetExts[ext] {
			entries = append(entries, fileEntry{path: path, size: info.Size()})
		}
		return nil
	})
	return entries, err
}

func cmdAudit() {
	root := projectRoot()
	entries, err := collectAssets(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("No asset files found.")
		return
	}

	var totalSize int64
	var warns, fails int

	for _, e := range entries {
		rel, _ := filepath.Rel(root, e.path)
		if rel == "" {
			rel = e.path
		}
		sev := severity(e.size)
		fmt.Printf("  %-4s  %s (%s)\n", sev, rel, humanSize(e.size))
		totalSize += e.size
		switch sev {
		case "WARN":
			warns++
		case "FAIL":
			fails++
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d files, %s\n", len(entries), humanSize(totalSize))
	if warns > 0 || fails > 0 {
		fmt.Printf("Warnings: %d  Errors: %d\n", warns, fails)
	}

	if fails > 0 {
		os.Exit(1)
	}
}

func cmdOptimize() {
	root := projectRoot()
	entries, err := collectAssets(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Optimization requires external tools. Install: pngquant, oxipng")
	fmt.Println()

	var count int
	for _, e := range entries {
		if e.size < 1<<20 {
			continue
		}
		rel, _ := filepath.Rel(root, e.path)
		if rel == "" {
			rel = e.path
		}
		fmt.Printf("  %s (%s)\n", rel, humanSize(e.size))
		count++
	}

	if count == 0 {
		fmt.Println("No files need optimization.")
	} else {
		fmt.Printf("\n%d file(s) would benefit from optimization.\n", count)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `gdt-assets - Asset optimization and auditing for Godot projects

Usage:
  gdt-assets <command>

Commands:
  audit      Scan project assets and report sizes
  optimize   Show files that would benefit from optimization
`)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "audit":
		cmdAudit()
	case "optimize":
		cmdOptimize()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
	}
}
