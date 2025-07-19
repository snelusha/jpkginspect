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

type PackageParser struct {
	parser *tree_sitter.Parser
	query  *tree_sitter.Query
}

func NewPackageParser() (*PackageParser, error) {
	parser := tree_sitter.NewParser()

	language := tree_sitter.NewLanguage(tree_sitter_java.Language())
	parser.SetLanguage(language)

	const q = `
	  (package_declaration
	    (scoped_identifier) @package_full)
	`

	query, _ := tree_sitter.NewQuery(language, string(q))
	if query == nil {
		parser.Close()
		return nil, fmt.Errorf("failed to create query")
	}

	return &PackageParser{
		parser: parser,
		query:  query,
	}, nil
}

func (p *PackageParser) GetPackageName(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("no source to parse")
	}

	tree := p.parser.Parse(raw, nil)
	defer tree.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	captures := cursor.Captures(p.query, tree.RootNode(), raw)
	match, _ := captures.Next()
	if match == nil {
		return "", fmt.Errorf("no package declaration found")
	}

	node := match.Captures[0].Node
	return node.Utf8Text(raw), nil
}

func (p *PackageParser) Close() {
	p.parser.Close()
	p.query.Close()
}

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

		parser, err := NewPackageParser()
		if err != nil {
			fmt.Printf("error creating parser: %v\n", err)
			continue
		}
		packageName, err := parser.GetPackageName(raw)
		if err != nil {
			fmt.Printf("error getting package name from file %s: %v\n", file, err)
			continue
		}
		fmt.Printf("file: %s, Package: %s\n", file, packageName)
	}
}
