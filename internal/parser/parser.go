package parser

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

type Parser struct {
	parser *tree_sitter.Parser
}

func NewParser() (*Parser, error) {
	parser := tree_sitter.NewParser()

	language := tree_sitter.NewLanguage(tree_sitter_java.Language())
	parser.SetLanguage(language)

	return &Parser{parser: parser}, nil
}

func (p *Parser) ExecuteQuery(q string, src []byte) ([]string, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("no source to parse")
	}

	tree := p.parser.Parse(src, nil)
	defer tree.Close()

	query, _ := tree_sitter.NewQuery(p.parser.Language(), q)

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	captures := cursor.Captures(query, tree.RootNode(), src)

	var results []string
	for match, _ := captures.Next(); match != nil; match, _ = captures.Next() {
		for _, capture := range match.Captures {
			results = append(results, capture.Node.Utf8Text(src))
		}
	}

	return results, nil
}

func (p *Parser) Close() {
	p.parser.Close()
}
