package query

import (
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// Examples in https://github.com/tree-sitter/go-tree-sitter/blob/master/query_test.go

// TODO: Why don't use starlark parser?
// e.g. syntax.Parse(filename string, src interface{}, mode syntax.Mode)
// - We would need a parser that parses blocks of code
// - We would need in-memory parse
// - We won't have custom queries
// - Parsers generate AST not an CST
// - In CST I can look direct into nodes an query specified node in position offset

const functionQuery = `
(function_definition
  name: (identifier) @function.name
	parameters: (parameters
			[
				(identifier)
				(default_parameter (identifier))
				(typed_parameter (identifier))
				(typed_default_parameter (identifier))
				(list_splat_pattern (identifier))
				(dictionary_splat_pattern (identifier))
			] @function.param
	) @function.params
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

func QuerySymbols(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
	symbols := []*symbol.Symbol{}
	classSymbols, err := ExtractClass(tree, text, source)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, classSymbols...)

	funcSymbols, err := ExtractFunctions(tree, text, source)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, funcSymbols...)

	varSymbols, err := ExtractVariables(tree, text, source)
	if err != nil {
		return nil, err
	}
	symbols = append(symbols, varSymbols...)

	return symbols, nil
}

func ExtractClass(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
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
		currClass := &symbol.Symbol{Kind: symbol.ClassKind, Source: source}
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
				currClass.FullyQualifiedName = patternValue
				currClass.Signature = currClass.Name
			case "class.docstring":
				currClass.DocString = sanitizeComment(patternValue)
			case "method.name":
				p := patternValue
				currentMethod = &p

				methodMap[*currentMethod] = &symbol.Symbol{
					Name:               patternValue,
					Source:             source,
					FullyQualifiedName: currClass.Name + "." + patternValue,
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
					s.DocString = sanitizeComment(patternValue)
				}
			}
		}

		currClass.Symbols = slices.Collect(maps.Values(methodMap))
		classes = append(classes, currClass)
	}

	return classes, nil
}

func ExtractFunctions(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
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

	matches := cursor.Matches(query, tree.RootNode(), text)

	funs := map[string]*symbol.Symbol{}

	for match := matches.Next(); match != nil; match = matches.Next() {
		currSymbol := &symbol.Symbol{Kind: symbol.FunctionKind, Source: source}
		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			nodeRange := capture.Node.Range()
			patternValue := capture.Node.Utf8Text(text)

			switch patternName {
			case "function.name":
				// Params query repeats Captures. We use Function Name as id
				if ss, ok := funs[patternValue]; ok {
					ss.Parameters = append(ss.Parameters, currSymbol.Parameters...)
					currSymbol = ss
				}

				currSymbol.Position.RowStart = nodeRange.StartPoint.Row
				currSymbol.Position.ColumnStart = nodeRange.StartPoint.Column
				currSymbol.Name = patternValue
				currSymbol.FullyQualifiedName = patternValue

				funs[patternValue] = currSymbol
			case "function.params":
				currSymbol.Signature = currSymbol.Name + patternValue
			case "function.param":
				newParam := &symbol.Parameter{Name: patternValue}
				currSymbol.Parameters = append(currSymbol.Parameters, newParam)
			case "function.docstring":
				currSymbol.DocString = sanitizeComment(patternValue)
			}

		}

	}

	return slices.Collect(maps.Values(funs)), nil
}

func ExtractVariables(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
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

		currSymbol := &symbol.Symbol{Kind: symbol.VariableKind, Source: source}
		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			patternValue := capture.Node.Utf8Text(text)

			switch patternName {
			case "var.comment":
				currSymbol.DocString = sanitizeComment(patternValue)
			case "var.name":
				currSymbol.Name = patternValue
				currSymbol.FullyQualifiedName = patternValue
			case "var.value":
				currSymbol.Value = patternValue
			}

		}

		vars = append(vars, currSymbol)
	}

	return vars, nil
}

func sanitizeComment(cmmt string) string {
	if cmmt, ok := strings.CutPrefix(cmmt, "#"); ok {
		return processCommentLines(cmmt)
	}

	if cmmt, ok := strings.CutPrefix(cmmt, "\"\"\""); ok {
		cmmt, _ = strings.CutSuffix(cmmt, "\"\"\"")
		return processCommentLines(cmmt)
	}

	return processCommentLines(cmmt)
}

func processCommentLines(comment string) string {
	lines := strings.Split(comment, "\n")
	var processedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			processedLines = append(processedLines, trimmed)
		}
	}

	return strings.Join(processedLines, "\n")
}
