package query

import (
	"errors"
	"fmt"

	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// Examples in https://github.com/tree-sitter/go-tree-sitter/blob/master/query_test.go

const functionQuery = `
(function_definition
  name: (identifier) @function.name
  parameters: (parameters) @function.params
  body: (block .
     (expression_statement
      (string (string_content) )) @function.docstring)?)
`

var ErrEmptyTreeError = errors.New("empty tree")

func ExtractFunctions(tree *tree_sitter.Tree, text []byte) ([]*symbol.Symbol, error) {
	// TODO: bsena; See how to use the global server
	lang := tree_sitter.NewLanguage(tree_sitter_python.Language())

	if tree.RootNode() == nil {
		return nil, ErrEmptyTreeError
	}

	query, err := tree_sitter.NewQuery(lang, functionQuery)
	if err != nil {
		return nil, err
	}

	defer query.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	functions := []*symbol.Symbol{}
	matches := cursor.Matches(query, tree.RootNode(), text)
	for match := matches.Next(); match != nil; match = matches.Next() {
		currSymbol := &symbol.Symbol{Kind: symbol.FunctionKind}
		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			nodeRange := capture.Node.Range()
			patternValue := capture.Node.Utf8Text(text)

			switch patternName {
			case "function.name":
				currSymbol.Position.RowStart = nodeRange.StartPoint.Row
				currSymbol.Position.ColumnStart = nodeRange.StartPoint.Column
				currSymbol.Name = patternValue
			case "function.params":
				currSymbol.Signature = currSymbol.Name + patternValue
			case "function.docString":
				currSymbol.DocString = patternValue
			}

			fmt.Printf(
				"Match %d, Capture %d (%s): %s\n",
				match.PatternIndex,
				capture.Index,
				patternName,
				patternValue,
			)
		}

		functions = append(functions, currSymbol)
	}

	return functions, nil
}
