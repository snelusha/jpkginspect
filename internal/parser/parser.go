package parser

import (
	"fmt"
	"slices"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
)

type ParsedFile struct {
	Package string
	Classes []string
	Imports []string
}

type Parser struct {
	parser *tree_sitter.Parser
	query  *tree_sitter.Query

	pkgIdx, clsIdx, impIdx, asteriskIdx uint32
}

func NewParser() (*Parser, error) {
	parser := tree_sitter.NewParser()

	language := tree_sitter.NewLanguage(tree_sitter_java.Language())
	parser.SetLanguage(language)

	const q = `
	  (package_declaration (scoped_identifier) @pkg)
	  (class_declaration name: (identifier) @cls)
	  (import_declaration (scoped_identifier) @imp)
	  (import_declaration ((scoped_identifier) (asterisk)) @asterisk)
	  (interface_declaration name: (identifier) @cls)
	  (enum_declaration name: (identifier) @cls)
	`

	query, _ := tree_sitter.NewQuery(language, q)
	if query == nil {
		parser.Close()
		return nil, fmt.Errorf("failed to parse query")
	}

	pkgIdx, _ := query.CaptureIndexForName("pkg")
	cls, _ := query.CaptureIndexForName("cls")
	imp, _ := query.CaptureIndexForName("imp")
	asterisk, _ := query.CaptureIndexForName("asterisk")

	return &Parser{
		parser:      parser,
		query:       query,
		pkgIdx:      uint32(pkgIdx),
		clsIdx:      uint32(cls),
		impIdx:      uint32(imp),
		asteriskIdx: uint32(asterisk),
	}, nil
}

func (p *Parser) Parse(src []byte) (*ParsedFile, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("no source to parse")
	}

	tree := p.parser.Parse(src, nil)
	if tree == nil {
		return nil, fmt.Errorf("failed to build tree")
	}
	defer tree.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	captures := cursor.Captures(p.query, tree.RootNode(), src)
	parsed := &ParsedFile{}

	var wildcardImports []string

	for match, _ := captures.Next(); match != nil; match, _ = captures.Next() {
		for _, cap := range match.Captures {
			text := cap.Node.Utf8Text(src)
			switch cap.Index {
			case p.pkgIdx:
				parsed.Package = text
			case p.clsIdx:
				parsed.Classes = append(parsed.Classes, text)
			case p.impIdx:
				parsed.Imports = append(parsed.Imports, text)
			case p.asteriskIdx:
				wildcardImports = append(wildcardImports, text)
			}
		}
	}

	if len(wildcardImports) > 0 {
		for i, imp := range parsed.Imports {
			if slices.Contains(wildcardImports, imp) {
				parsed.Imports[i] = fmt.Sprintf("%s.*", imp)
			}
		}
	}

	return parsed, nil
}

func (p *Parser) Close() {
	p.parser.Close()
}
