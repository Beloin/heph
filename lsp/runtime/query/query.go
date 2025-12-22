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

const variablesQuery = `
(
 ((comment) @var.comment)? .
 (expression_statement
	(assignment
		left: (identifier) @var.name
		right: (
			[
				(integer)
				(string)
				(float)
				(list)
				(dictionary)
				(call)
				(identifier)
				(binary_operator)
			] @var.value)
		))
)
`

var ErrEmptyTreeError = errors.New("empty tree")

// TODO: bsena; See how to use the global server
var lang = tree_sitter.NewLanguage(tree_sitter_python.Language())

func ExtractFunctions(tree *tree_sitter.Tree, text []byte) ([]*symbol.Symbol, error) {
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

func ExtractVariables(tree *tree_sitter.Tree, text []byte) ([]*symbol.Symbol, error) {
	root := tree.RootNode()
	if root == nil {
		return nil, ErrEmptyTreeError
	}

	query, err := tree_sitter.NewQuery(lang, variablesQuery)
	if err != nil {
		return nil, err
	}

	defer query.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	vars := []*symbol.Symbol{}
	matches := cursor.Matches(query, root, text)
	for match := matches.Next(); match != nil; match = matches.Next() {
		// TODO: how to work with repeated symbols?

		currSymbol := &symbol.Symbol{Kind: symbol.VariableKind}
		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			patternValue := capture.Node.Utf8Text(text)

			switch patternName {
			case "var.comment":
				currSymbol.DocString = patternValue
			case "var.name":
				currSymbol.Name = patternValue
			case "var.value":
				currSymbol.Value = patternValue
			}

			fmt.Printf(
				"Match %d, Capture %d (%s): %s\n",
				match.PatternIndex,
				capture.Index,
				patternName,
				patternValue,
			)
		}

		vars = append(vars, currSymbol)
	}

	return vars, nil
}
