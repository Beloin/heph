package query

import (
	"maps"
	"slices"
	"strconv"

	"github.com/hephbuild/heph/lsp/runtime/symbol"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const callQuery = `
(call function: (identifier) @call.name . (argument_list (_) @call.arg) ) @call.stmt
`

func QueryCalls(tree *tree_sitter.Tree, text []byte, source string) ([]*symbol.Symbol, error) {
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
	funs := map[string]*symbol.Symbol{}

	for match := matches.Next(); match != nil; match = matches.Next() {
		currSymbol := &symbol.Symbol{Kind: symbol.FunctionCallKind, Source: source}
		for _, capture := range match.Captures {
			patternName := query.CaptureNames()[capture.Index]
			nodeRange := capture.Node.Range()
			patternValue := capture.Node.Utf8Text(text)

			switch patternName {
			case "call.name":
				// Params query repeats Captures. We use Call Name as id so we dont need to make multiple queries
				if ss, ok := funs[patternValue]; ok {
					ss.Parameters = append(ss.Parameters, currSymbol.Parameters...)
					currSymbol = ss
				}

				currSymbol.Name = patternValue
				currSymbol.FullyQualifiedName = patternValue
				currSymbol.Position.RowStart = nodeRange.StartPoint.Row
				currSymbol.Position.ColumnStart = nodeRange.StartPoint.Column

				funs[patternValue] = currSymbol
			case "call.arg":
				newParam := symbol.Parameter{Name: strconv.Itoa(len(currSymbol.Parameters)), Value: patternValue}
				currSymbol.Parameters = append(currSymbol.Parameters, &newParam)
			case "call.stmt":
				currSymbol.Signature = patternValue
				currSymbol.Position.RowEnd = nodeRange.EndPoint.Row
				currSymbol.Position.ColumnEnd = nodeRange.EndPoint.Column
			}
		}
	}

	return slices.Collect(maps.Values(funs)), nil
}
