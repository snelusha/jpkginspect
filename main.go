package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

func findFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".java") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("no directory provided!")
		return
	}

	dir := args[0]

	files, err := findFiles(dir)
	if err != nil {
		fmt.Printf("error finding files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("no .java files found")
		return
	}

	fmt.Printf("found %d .java files:\n", len(files))

	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("error reading file %s: %v\n", file, err)
			continue
		}

		parser := tree_sitter.NewParser()
		defer parser.Close()

		language := tree_sitter.NewLanguage(tree_sitter_java.Language())
		parser.SetLanguage(language)

		tree := parser.Parse(raw, nil)
		defer tree.Close()

		root := tree.RootNode()

		const q = `
		  (package_declaration
		    (scoped_identifier) @package_full)
		`

		query, _ := tree_sitter.NewQuery(language, string(q))
		defer query.Close()

		cursor := tree_sitter.NewQueryCursor()
		defer cursor.Close()

		captures := cursor.Captures(query, root, raw)
		for match, idx := captures.Next(); match != nil; match, idx = captures.Next() {
			node := match.Captures[idx].Node
			text := node.Utf8Text(raw)
			fmt.Printf("Found package: %s in file: %s\n", text, file)
		}
	}
}
