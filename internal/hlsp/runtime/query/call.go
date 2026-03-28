package query

import (
	"maps"
	"slices"
	"strconv"

	"github.com/hephbuild/heph/internal/hlsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const callQuery = `
(call function: (identifier) @call.name . (argument_list (_) @call.arg)? ) @call.stmt
`

// Necessary to do a re-run to get calls that are nested in functions.
type callEntry struct {
	sym      *symbol.Symbol
	stmtNode *tree_sitter.Node
}

// ExtractCalls extracts function call symbols from the tree.
// If funs is provided, calls made inside a function body are attached to that
// function's Symbols field instead of being returned at the top level.
func ExtractCalls(tree *tree_sitter.Tree, text []byte, source string, funs []*symbol.Symbol) ([]*symbol.Symbol, error) {
	if tree.RootNode() == nil {
		return nil, ErrEmptyTreeError
	}

	query, err := tree_sitter.NewQuery(lang, callQuery)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	matches := cursor.Matches(query, tree.RootNode(), text)
	entries := map[uintptr]*callEntry{}

	for match := matches.Next(); match != nil; match = matches.Next() {
		currEntry := &callEntry{sym: &symbol.Symbol{Kind: symbol.FunctionCallKind, Source: source}}
		for _, capture := range match.Captures {
			currentNode := &capture.Node
			patternName := query.CaptureNames()[capture.Index]
			nodeRange := currentNode.Range()
			patternValue := currentNode.Utf8Text(text)

			switch patternName {
			case "call.name":
				// Params query repeats Captures. We use NodeId so we dont need to make multiple queries
				if e, ok := entries[currentNode.Id()]; ok {
					e.sym.Parameters = append(e.sym.Parameters, currEntry.sym.Parameters...)
					currEntry = e
				}

				currEntry.sym.Name = patternValue
				currEntry.sym.FullyQualifiedName = patternValue
				currEntry.sym.Position.RowStart = nodeRange.StartPoint.Row
				currEntry.sym.Position.ColumnStart = nodeRange.StartPoint.Column
				currEntry.sym.Position.ByteStart = nodeRange.StartByte

				entries[currentNode.Id()] = currEntry
			case "call.arg":
				newParam := symbol.Parameter{Name: strconv.Itoa(len(currEntry.sym.Parameters)), Value: patternValue}

				// Is kwarg
				if currentNode.Kind() == "keyword_argument" {
					if name := currentNode.ChildByFieldName("name"); name != nil {
						newParam.Name = name.Utf8Text(text)
					}

					if value := currentNode.ChildByFieldName("value"); value != nil {
						newParam.Value = value.Utf8Text(text)
						newParam.Type = ResolveType(value.Kind())
					}
				} else {
					newParam.Type = ResolveType(currentNode.Kind())
				}

				currEntry.sym.Parameters = append(currEntry.sym.Parameters, &newParam)

				// Last pattern
				currEntry.sym.Position.RowEnd = nodeRange.EndPoint.Row
				currEntry.sym.Position.ColumnEnd = nodeRange.EndPoint.Column
			case "call.stmt":
				currEntry.sym.Signature = patternValue
				currEntry.sym.Position.ByteEnd = nodeRange.EndByte
				currEntry.stmtNode = currentNode
			}
		}
	}

	topLevel := []*symbol.Symbol{}
	for _, entry := range slices.Collect(maps.Values(entries)) {
		if funs != nil && entry.stmtNode != nil {
			if nameNode := getFunctionNameNodeIfExists(entry.stmtNode); nameNode != nil {
				funcName := nameNode.Utf8Text(text)
				if fn, found := symbol.FindSymbol(funs, funcName); found {
					fn.Symbols = append(fn.Symbols, entry.sym)
					continue
				}
			}
		}

		topLevel = append(topLevel, entry.sym)
	}

	return topLevel, nil
}

// QueryCalls is a convenience wrapper that returns all calls without scope awareness.
// TODO: bsena; remove this
func QueryCalls(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
	return ExtractCalls(tree, text, source, nil)
}
