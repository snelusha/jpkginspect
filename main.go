package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	parser "jpkginspect/internal/parser"
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

		parser, err := parser.NewParser()
		if err != nil {
			fmt.Printf("error creating parser: %v\n", err)
			continue
		}

		const q = `
		  (package_declaration
		    (scoped_identifier) @package_full)
		`

		results, err := parser.ExecuteQuery(string(q), raw)
		if err != nil {
			fmt.Printf("error executing query on file %s: %v\n", file, err)
			continue
		}
		if len(results) == 0 {
			fmt.Printf("no package declaration found in file %s\n", file)
			continue
		}
		fmt.Printf("file: %s, package: %s\n", file, results[0])
	}
}
