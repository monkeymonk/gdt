package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		cmdList()
	case "tree":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gdt-scenes tree <scene.tscn>")
			os.Exit(1)
		}
		cmdTree(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: gdt-scenes <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  list              List all .tscn files in the project")
	fmt.Fprintln(os.Stderr, "  tree <file.tscn>  Display scene node hierarchy")
}

func projectRoot() string {
	root := os.Getenv("GDT_PROJECT_ROOT")
	if root == "" {
		fmt.Fprintln(os.Stderr, "Error: GDT_PROJECT_ROOT is not set")
		os.Exit(1)
	}
	return root
}

func cmdList() {
	root := projectRoot()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".godot" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tscn") {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				rel = path
			}
			fmt.Println(rel)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking project: %v\n", err)
		os.Exit(1)
	}
}

type node struct {
	name     string
	typ      string
	children []*node
}

func cmdTree(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var root *node
	nodes := make(map[string]*node) // path -> node

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "[node ") {
			continue
		}

		name := extractAttr(line, "name")
		typ := extractAttr(line, "type")
		parent := extractAttr(line, "parent")

		n := &node{name: name, typ: typ}

		if parent == "" {
			// Root node
			root = n
			nodes["."] = n
		} else {
			// Compute full path for this node
			var fullPath string
			if parent == "." {
				fullPath = name
			} else {
				fullPath = parent + "/" + name
			}
			nodes[fullPath] = n

			// Find parent node
			var parentNode *node
			if parent == "." {
				parentNode = root
			} else {
				parentNode = nodes[parent]
			}
			if parentNode != nil {
				parentNode.children = append(parentNode.children, n)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if root == nil {
		fmt.Fprintln(os.Stderr, "No root node found in scene file")
		os.Exit(1)
	}

	printTree(root, 0)
}

func extractAttr(line, attr string) string {
	key := attr + "=\""
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	end := strings.Index(line[start:], "\"")
	if end < 0 {
		return ""
	}
	return line[start : start+end]
}

func printTree(n *node, depth int) {
	indent := strings.Repeat("  ", depth)
	if n.typ != "" {
		fmt.Printf("%s%s (%s)\n", indent, n.name, n.typ)
	} else {
		fmt.Printf("%s%s\n", indent, n.name)
	}
	for _, child := range n.children {
		printTree(child, depth+1)
	}
}
