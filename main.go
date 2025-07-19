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

	parser, err := parser.NewParser()
	if err != nil {
		fmt.Printf("error creating parser: %v\n", err)
		return
	}
	defer parser.Close()

	packageIndex := make(map[string]map[string]string)

	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("error reading file %s: %v\n", file, err)
			continue
		}

		parsed, err := parser.Parse(raw)
		if err != nil {
			fmt.Printf("error parsing file %s: %v\n", file, err)
			continue
		}

		if _, exists := packageIndex[parsed.Package]; !exists {
			packageIndex[parsed.Package] = make(map[string]string)
		}

		for _, cls := range parsed.Classes {
			packageIndex[parsed.Package][cls] = file
		}
	}

	fmt.Println(packageIndex)
}
