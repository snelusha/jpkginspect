package cmd

import (
	"encoding/json"
	"fmt"
	"jpkginspect/internal/parser"
	"jpkginspect/internal/types"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	findCmd.Flags().StringP("input", "i", "output.json", "input file containing the package index")
}

var findCmd = &cobra.Command{
	Use:   "find [path]",
	Short: "Find the all classes that used in a Java file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Flag("input").Value.String()

		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			return
		}

		var index types.PackageIndex
		if err := json.Unmarshal(raw, &index); err != nil {
			fmt.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}

		parser, err := parser.NewParser()
		if err != nil {
			fmt.Printf("Error parsing files: %v\n", err)
			return
		}
		defer parser.Close()

		filePath := args[0]

		var queue []string
		visited := make(map[string]bool)

		queue = append(queue, filePath)

		graph := NewGraph()

		for len(queue) > 0 {
			path, queue = queue[0], queue[1:]

			if visited[path] {
				continue
			}

			visited[path] = true

			source, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				continue
			}
			parsed, err := parser.Parse(source)
			if err != nil {
				fmt.Printf("Error parsing file %s: %v\n", path, err)
				continue
			}

			for _, cls := range parsed.Classes {
				graph.AddNode(cls)
			}

			if len(parsed.Imports) == 0 {
				continue
			}

			normalizedImports := normalizeImports(parsed.Imports, index)

			for imp, path := range normalizedImports {
				for _, cls := range parsed.Classes {
					graph.AddEdge(cls, imp)
				}

				queue = append(queue, path)
			}
		}

		ordered, err := graph.TopologicalSort()
		if err != nil {
			fmt.Printf("Error during topological sort: %v\n", err)
			return
		}

		for _, cls := range ordered {
			fmt.Println(cls)
		}
	},
}

func normalizeImports(imports []string, index types.PackageIndex) (normalized map[string]string) {
	normalized = make(map[string]string)

	for _, imp := range imports {
		if strings.HasSuffix(imp, ".*") {
			pkg := strings.TrimSuffix(imp, ".*")

			for class := range index[pkg] {
				if _, exists := normalized[class]; !exists {
					normalized[class] = index[pkg][class]
				}
			}
		} else {
			parts := strings.Split(imp, ".")

			pkg := strings.Join(parts[:len(parts)-1], ".")
			class := parts[len(parts)-1]

			if _, exists := index[pkg][class]; exists {
				normalized[class] = index[pkg][class]
			}
		}
	}

	return normalized
}

type Graph struct {
	adj map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		adj: make(map[string][]string),
	}
}

func (g *Graph) AddNode(node string) {
	if _, exists := g.adj[node]; !exists {
		g.adj[node] = nil
	}
}

func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)

	g.adj[from] = append(g.adj[from], to)
}

func (g *Graph) TopologicalSort() ([]string, error) {
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var result []string

	var visit func(string) error
	visit = func(node string) error {
		if temp[node] {
			return nil
		}

		if !visited[node] {
			temp[node] = true
			for _, m := range g.adj[node] {
				if err := visit(m); err != nil {
					return err
				}
			}
			temp[node] = false
			visited[node] = true
			result = append(result, node)
		}
		return nil
	}

	for node := range g.adj {
		if !visited[node] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
