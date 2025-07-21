package cmd

import (
	"fmt"
	"jpkginspect/internal/exporter"
	"jpkginspect/internal/fs"
	"jpkginspect/internal/parser"
	"jpkginspect/internal/types"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	inspectCmd.Flags().StringP("output", "o", "output.json", "output file for the JSON dump")
}

var inspectCmd = &cobra.Command{
	Use:   "inspect [path]",
	Short: "Inspect a Java codebase and dump JSON representation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		files, err := fs.WalkJavaFiles(path)
		if err != nil {
			fmt.Printf("Error finding files: %v\n", err)
		}

		if len(files) == 0 {
			fmt.Println("No .java files found")
			return
		}

		parser, err := parser.NewParser()
		if err != nil {
			fmt.Printf("Error parsing files: %v\n", err)
			return
		}
		defer parser.Close()

		index := make(types.PackageIndex)

		for _, file := range files {
			raw, err := os.ReadFile(file)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", file, err)
				continue
			}

			parsed, err := parser.Parse(raw)
			if err != nil {
				fmt.Printf("Error parsing file %s: %v\n", file, err)
				continue
			}

			if _, exists := index[parsed.Package]; !exists {
				index[parsed.Package] = make(map[string]string)
			}

			for _, class := range parsed.Classes {
				index[parsed.Package][class] = file
			}
		}

		outputFile, _ := cmd.Flags().GetString("output")
		if err := exporter.WriteJSON(index, outputFile); err != nil {
			fmt.Printf("Error exporting JSON: %v\n", err)
			return
		}
		fmt.Printf("Inpsected %d files\n", len(files))
		fmt.Printf("Dumped JSON representation to %s\n", outputFile)
	},
}
