package query

import (
	"errors"

	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

const functionQuery = `
  (function_definition
	  name: (identifier) @function.name
	  parameters: (parameters) @function.params)
`

var ErrEmptyTreeError = errors.New("empty tree")

func ExtractFunctions(tree *tree_sitter.Tree, text []byte) ([]*symbol.Symbol, error) {
	// TODO: bsena; See how to use the global server
	lang := tree_sitter.NewLanguage(tree_sitter_python.Language())

	if tree.RootNode() == nil {
		return nil, ErrEmptyTreeError
	}

	query, err := tree_sitter.NewQuery(lang, functionQuery)
	defer query.Close()
	if err != nil {
		return nil, err
	}

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	matches := cursor.Captures(query, tree.RootNode(), text)
	match, _ := matches.Next()
	for match != nil {
		for _, capture := range match.Captures {
			node := capture.Node
			node.StartPosition()
			node.EndPosition()
			node.Kind()
		}

		match, _ = matches.Next()
	}

	// matches := cu.Matches(q, tree.RootNode(), rawText)
	// first := matches.Next()
	// captures := first.Captures
	// capture := captures[0]
	//
	// newCaps := cu.Captures(q, tree.RootNode(), rawText)
	// first2, _ := newCaps.Next()
	// caps := first2.Captures

	return nil, nil
}
