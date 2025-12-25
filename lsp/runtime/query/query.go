package query

import (
	"errors"
	"maps"
	"slices"

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

const classQuery = `
(class_definition
  name: (identifier) @class.name
  body: (block .
		((expression_statement
			(string (string_content) )) @class.docstring)?
		(function_definition
			name: (identifier) @method.name
			parameters: (parameters) @method.params
			body: (block .
				(expression_statement
					(string (string_content) @method.docstring))?))*))
`

var ErrEmptyTreeError = errors.New("empty tree")

// TODO: bsena; See how to use the global server
var lang = tree_sitter.NewLanguage(tree_sitter_python.Language())

func QuerySymbols(tree *tree_sitter.Tree, text []byte) ([]*symbol.Symbol, error) {
	symbols := []*symbol.Symbol{}
	classSymbols, err := ExtractClass(tree, text)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, classSymbols...)

	funcSymbols, err := ExtractFunctions(tree, text)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, funcSymbols...)

	varSymbols, err := ExtractVariables(tree, text)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, varSymbols...)

	return symbols, nil
}

func ExtractClass(tree *tree_sitter.Tree, text []byte) ([]*symbol.Symbol, error) {
	if tree.RootNode() == nil {
		return nil, ErrEmptyTreeError
	}

	query, err := tree_sitter.NewQuery(lang, classQuery)
	if err != nil {
		return nil, err
	}

	defer query.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	classes := []*symbol.Symbol{}
	matches := cursor.Matches(query, tree.RootNode(), text)
	for match := matches.Next(); match != nil; match = matches.Next() {
		currClass := &symbol.Symbol{Kind: symbol.ClassKind}
		methodMap := map[string]*symbol.Symbol{}
		var currentMethod *string
		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			nodeRange := capture.Node.Range()
			patternValue := capture.Node.Utf8Text(text)

			switch patternName {
			case "class.name":
				currClass.Position.RowStart = nodeRange.StartPoint.Row
				currClass.Position.ColumnStart = nodeRange.StartPoint.Column
				currClass.Name = patternValue
				currClass.Signature = currClass.Name + patternValue
			case "class.docstring":
				currClass.DocString = patternValue
			case "method.name":
				p := patternValue
				currentMethod = &p

				methodMap[*currentMethod] = &symbol.Symbol{
					Name: patternValue,
					Position: symbol.Position{
						RowStart:    nodeRange.StartPoint.Row,
						ColumnStart: nodeRange.StartPoint.Column,
					},
				}
			case "method.params":
				if currentMethod != nil {
					s := methodMap[*currentMethod]
					s.Signature = s.Name + patternValue
				}
			case "method.docstring":
				if currentMethod != nil {
					s := methodMap[*currentMethod]
					s.DocString = patternValue
				}
			}
		}

		currClass.Symbols = slices.Collect(maps.Values(methodMap))
		classes = append(classes, currClass)
	}

	return classes, nil
}

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
			case "function.docstring":
				currSymbol.DocString = patternValue
			}

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

		}

		vars = append(vars, currSymbol)
	}

	return vars, nil
}
