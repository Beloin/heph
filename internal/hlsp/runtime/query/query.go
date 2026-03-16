package query

import (
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// Examples in https://github.com/tree-sitter/go-tree-sitter/blob/master/query_test.go

const functionQuery = `
(function_definition
  name: (identifier) @function.name
	parameters: (parameters
			[
				(identifier) @function.param
				(default_parameter ( (identifier) @function.param . (_) @function.param.value ))
				(typed_parameter (identifier) @function.param (type (_) @function.param.type))
				(typed_default_parameter ( ((identifier) @function.param) . (type (_) @function.param.type) . ((_) @function.param.value) ))
				(list_splat_pattern (identifier) @function.param)
				(dictionary_splat_pattern (identifier) @function.param)
			]
	)? @function.params
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
		right: (_) @var.value
		))
)
`

// TODO: bsena; Also extract struct like expressions so you can match them
// this will also be helpfull with heph.obj.otherfn
// And also wil be helpfull with new driver providers
// They are object:
//   (expression_statement ; [31, 0] - [31, 5]
// (attribute ; [31, 0] - [31, 5]
//   object: (attribute ; [31, 0] - [31, 3]
//     object: (identifier) ; [31, 0] - [31, 1]
//     attribute: (identifier)) ; [31, 2] - [31, 3]
//   attribute: (identifier)))) ; [31, 4] - [31, 5]
//

var ErrEmptyTreeError = errors.New("empty tree")

var lang = tree_sitter.NewLanguage(tree_sitter_python.Language())

// TODO: Maybe create a query that query all symbols of type object
// so we can match heph.myfun

// QueryResult holds the three categories of symbols extracted from a tree.
type QueryResult struct {
	Functions []*symbol.Symbol
	Variables []*symbol.Symbol
	Calls     []*symbol.Symbol
}

// QueryAll extracts functions, variables, and calls in one pass, with full
// scope awareness: variables and calls inside a function body are attached to
// that function's Symbols field.
func QueryAll(tree *tree_sitter.Tree, text []byte, source string) (*QueryResult, error) {
	funs, err := ExtractFunctions(tree, text, source)
	if err != nil {
		return nil, err
	}

	vars, err := ExtractVariables(tree, text, source, funs)
	if err != nil {
		return nil, err
	}

	calls, err := ExtractCalls(tree, text, source, funs)
	if err != nil {
		return nil, err
	}

	return &QueryResult{
		Functions: funs,
		Variables: vars,
		Calls:     calls,
	}, nil
}

func QuerySymbols(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
	result, err := QueryAll(tree, text, source)
	if err != nil {
		return nil, err
	}

	return append(result.Functions, result.Variables...), nil
}

// Necessary to do a re-run to get functions that are nested.
type funcEntry struct {
	sym     *symbol.Symbol
	defNode *tree_sitter.Node
}

// ExtractFunctions extracts function symbols from the tree.
// Nested functions (functions defined inside another function) are attached to
// their parent function's Symbols field rather than returned at the top level.
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

	entries := map[uintptr]*funcEntry{}

	for match := matches.Next(); match != nil; match = matches.Next() {
		currEntry := &funcEntry{sym: &symbol.Symbol{Kind: symbol.FunctionKind, Source: source}}
		var currParam *symbol.Parameter
		for _, capture := range match.Captures {
			currentNode := &capture.Node
			patternName := query.CaptureNames()[capture.Index]
			nodeRange := currentNode.Range()
			patternValue := currentNode.Utf8Text(text)

			switch patternName {
			case "function.name":
				// Params query repeats Captures. We use Function Name as id so we dont need to make multiple queries
				if e, ok := entries[currentNode.Id()]; ok {
					e.sym.Parameters = append(e.sym.Parameters, currEntry.sym.Parameters...)
					currEntry = e
				}

				// First capture group
				currEntry.sym.Position.RowStart = nodeRange.StartPoint.Row
				currEntry.sym.Position.ColumnStart = nodeRange.StartPoint.Column
				currEntry.sym.Position.ByteStart = nodeRange.StartByte

				currEntry.sym.Name = patternValue
				currEntry.sym.Signature = patternValue + "()" // empty params is the default
				currEntry.sym.FullyQualifiedName = patternValue
				currEntry.defNode = currentNode.Parent() // function_definition

				entries[currentNode.Id()] = currEntry
			case "function.params":
				currEntry.sym.Signature = currEntry.sym.Name + patternValue
			case "function.param":
				currParam = &symbol.Parameter{Name: patternValue}
				currEntry.sym.Parameters = append(currEntry.sym.Parameters, currParam)
			case "function.param.type":
				if currParam != nil {
					currParam.Type = ResolveType(patternValue)
				}
			case "function.param.value":
				if currParam != nil {
					currParam.Value = patternValue
				}
			case "function.docstring":
				currEntry.sym.DocString = sanitizeComment(patternValue)

				// Last capture group
				currEntry.sym.Position.RowEnd = nodeRange.EndPoint.Row
				currEntry.sym.Position.ColumnEnd = nodeRange.EndPoint.Column
				currEntry.sym.Position.ByteEnd = nodeRange.EndByte
			}
		}

		parseArgsFromDocstring(currEntry.sym.DocString, currEntry.sym.Parameters)
	}

	// Build a name->entry map for parent lookup.
	entryByName := map[string]*funcEntry{}
	for _, e := range entries {
		entryByName[e.sym.Name] = e
	}

	// Second pass: attach nested functions to their parent's Symbols.
	topLevel := []*symbol.Symbol{}
	for _, entry := range slices.Collect(maps.Values(entries)) {
		parentNameNode := getFunctionNameNodeIfExists(entry.defNode.Parent())
		if parentNameNode == nil {
			topLevel = append(topLevel, entry.sym)
			continue
		}

		parentName := parentNameNode.Utf8Text(text)
		if parentEntry, ok := entryByName[parentName]; ok {
			parentEntry.sym.Symbols = append(parentEntry.sym.Symbols, entry.sym)
		} else {
			topLevel = append(topLevel, entry.sym)
		}
	}

	return topLevel, nil
}

func ExtractVariables(tree *tree_sitter.Tree, text []byte, source string, funs []*symbol.Symbol) ([]*symbol.Symbol, error) {
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
		var lastNode *tree_sitter.Node

		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			node := &capture.Node
			lastNode = node

			patternValue := node.Utf8Text(text)
			nodeRange := node.Range()

			switch patternName {
			case "var.name":
				// First capture group
				currSymbol.Position.RowStart = nodeRange.StartPoint.Row
				currSymbol.Position.ColumnStart = nodeRange.StartPoint.Column
				currSymbol.Position.ByteStart = nodeRange.StartByte

				currSymbol.Name = patternValue
				currSymbol.Signature = patternValue
				currSymbol.FullyQualifiedName = patternValue
			case "var.comment":
				currSymbol.DocString = sanitizeComment(patternValue)
			case "var.value":
				currSymbol.Type = ResolveType(node.Kind())
				currSymbol.Value = patternValue

				// Last capture group — set position then check for enclosing function.
				currSymbol.Position.RowEnd = nodeRange.EndPoint.Row
				currSymbol.Position.ColumnEnd = nodeRange.EndPoint.Column
				currSymbol.Position.ByteEnd = nodeRange.EndByte
			}
		}

		if funNode := getFunctionNameNodeIfExists(lastNode); funNode != nil {
			funcName := funNode.Utf8Text(text)
			if fn, found := symbol.FindSymbol(funs, funcName); found {
				fn.Symbols = append(fn.Symbols, currSymbol)
			}
		} else {
			vars = append(vars, currSymbol)
		}
	}

	return vars, nil
}

func getFunctionNameNodeIfExists(node *tree_sitter.Node) *tree_sitter.Node {
	for funNode := node; funNode.Parent() != nil; funNode = funNode.Parent() {
		if funNode.Kind() != "function_definition" {
			continue
		}

		nameNode := funNode.ChildByFieldName("name")
		if nameNode == nil {
			break
		}

		return nameNode
	}

	return nil
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

func parseArgsFromDocstring(docstring string, params []*symbol.Parameter) {
	lines := strings.Split(docstring, "\n")
	inArgs := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Args:" {
			inArgs = true
			continue
		}
		if inArgs && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				namePart := strings.TrimSpace(parts[0])
				desc := strings.TrimSpace(parts[1])
				if idx := strings.Index(namePart, " ("); idx > 0 {
					paramName := namePart[:idx]
					for _, p := range params {
						if p.Name == paramName {
							p.DocString = desc
							break
						}
					}
				}
			}
		}
	}
}
